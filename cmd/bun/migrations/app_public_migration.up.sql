-- Enable UUID generation
-- UUID + crypto-safe IDs
CREATE EXTENSION IF NOT EXISTS pgcrypto;
-- For nice text search later (optional but handy for search boxes)
CREATE EXTENSION IF NOT EXISTS unaccent;
CREATE EXTENSION IF NOT EXISTS citext;

CREATE SCHEMA IF NOT EXISTS app;
CREATE SCHEMA IF NOT EXISTS config;

CREATE TYPE app.org_role AS ENUM ('owner','admin','manager','mechanic','viewer');
CREATE TYPE app.work_order_status AS ENUM (
    'draft','new','checking','scheduled','awaiting_customer','in_progress','waiting_parts','awaiting_approval','ready_for_pickup','ready_for_deliver','en_route','completed','canceled'
    );
CREATE TYPE app.work_order_priority AS ENUM ('low','normal','high','urgent');
CREATE TYPE app.line_item_type AS ENUM ('labor','part','fee','other');
CREATE TYPE app.notify_channel AS ENUM ('email','sms','whatsapp','push','webhook');
CREATE TYPE app.appointment_status AS ENUM ('pending','confirmed','cancelled','no_show','completed');

-- =========================
-- 0) Phone Numbers (normalization)
-- =========================
-- Store phone numbers in E.164 format for consistency and lookup.
-- You can use a library like libphonenumber to parse/format numbers on input.
-- This table holds unique phone numbers.

CREATE TABLE IF NOT EXISTS public.phone_numbers
(
    id              UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    raw_number      TEXT        NOT NULL,
    e164            TEXT        NOT NULL,
    country_code    TEXT        NOT NULL,
    national_number TEXT        NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (country_code, national_number)
);
CREATE UNIQUE INDEX idx_phone_numbers_e164 ON public.phone_numbers (e164);

-- =========================
-- 1) Stack Auth <-> Internal mapping
-- =========================

-- Users you control internally, mapped to Stack Auth user.id (string).
CREATE TABLE app.users
(
    id            UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    -- Stack Auth user id is a string; keep it as TEXT and unique
    stack_user_id TEXT UNIQUE NOT NULL,
    email         TEXT        NOT NULL, -- store for convenience/joins; keep in sync via webhook
    display_name  TEXT,
    avatar_url    TEXT,
    is_active     BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE app.user_phone_numbers
(
    id              UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    user_id         UUID REFERENCES app.users (id) ON DELETE CASCADE,
    phone_number_id UUID        NOT NULL REFERENCES public.phone_numbers (id) ON DELETE CASCADE,
    is_primary      BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, phone_number_id),
    UNIQUE (user_id, is_primary)
);
CREATE INDEX idx_user_phone_numbers_user ON app.user_phone_numbers (user_id);
CREATE INDEX idx_user_phone_numbers_phone ON app.user_phone_numbers (phone_number_id);

-- Your organizations (tenants). You own the concept, but you can also
-- keep a pointer to Stack's Team id if you use their Teams feature.
CREATE TABLE app.organizations
(
    id            UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    name          TEXT        NOT NULL,
    stack_team_id TEXT UNIQUE, -- optional link to Stack "team" if you enable teams there
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Organization membership and roles (your own RBAC).
-- You can keep this fully internal; Stack Auth’s team objects are minimal. :contentReference[oaicite:1]{index=1}
CREATE TABLE app.organization_members
(
    id              UUID PRIMARY KEY      DEFAULT gen_random_uuid(),
    organization_id UUID         NOT NULL REFERENCES app.organizations (id) ON DELETE CASCADE,
    user_id         UUID         NOT NULL REFERENCES app.users (id) ON DELETE CASCADE,
    role            app.org_role NOT NULL DEFAULT 'viewer',
    invited_by      UUID         REFERENCES app.users (id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    UNIQUE (organization_id, user_id)
);

-- Optional: projects under an organization, kept internally
CREATE TABLE app.projects
(
    id              UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    organization_id UUID        NOT NULL REFERENCES app.organizations (id) ON DELETE CASCADE,
    name            TEXT        NOT NULL,
    description     TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (organization_id, name)
);

-- =========================
-- 2) Customers & Vehicles
-- =========================

-- Basic customer record scoped to an organization (tenant).
CREATE TABLE public.customers
(
    id              UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    organization_id UUID        NOT NULL REFERENCES app.organizations (id) ON DELETE CASCADE,
    full_name       TEXT        NOT NULL,
    email           TEXT,
    notes           TEXT,
    created_by      UUID        REFERENCES app.users (id) ON DELETE SET NULL,
    updated_by      UUID        REFERENCES app.users (id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_customers_org ON public.customers (organization_id);
CREATE INDEX idx_customers_search ON public.customers USING GIN ((to_tsvector('simple',
                                                                              coalesce(full_name, '') || ' ' ||
                                                                              coalesce(email, ''))));

-- Link customer to multiple phone numbers
CREATE TABLE public.customer_phone_numbers
(
    id              UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    customer_id     UUID REFERENCES public.customers (id) ON DELETE CASCADE,
    phone_number_id UUID        NOT NULL REFERENCES public.phone_numbers (id) ON DELETE CASCADE,
    is_primary      BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (customer_id, phone_number_id),
    UNIQUE (customer_id, is_primary)
);
CREATE INDEX idx_cust_phone_numbers_cust ON public.customer_phone_numbers (customer_id);
CREATE INDEX idx_cust_phone_numbers_phone ON public.customer_phone_numbers (phone_number_id);

-- Vehicles belong to a customer (and implicitly to the same org).
CREATE TABLE public.vehicles
(
    id              UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    organization_id UUID        NOT NULL REFERENCES app.organizations (id) ON DELETE CASCADE,
    customer_id     UUID        NOT NULL REFERENCES public.customers (id) ON DELETE CASCADE,
    vin             TEXT, -- consider UNIQUE per org if needed
    plate_number    TEXT,
    make            TEXT,
    model           TEXT,
    year            INT,
    color           TEXT,
    mileage_km      INT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (organization_id, vin) DEFERRABLE INITIALLY DEFERRED
);
CREATE INDEX idx_vehicles_org ON public.vehicles (organization_id);
CREATE INDEX idx_vehicles_customer ON public.vehicles (customer_id);
CREATE INDEX idx_vehicles_plate ON public.vehicles (plate_number);

-- =========================
-- 3) Work Orders
-- =========================

CREATE TABLE app.work_orders
(
    id              UUID PRIMARY KEY                 DEFAULT gen_random_uuid(),
    organization_id UUID                    NOT NULL REFERENCES app.organizations (id) ON DELETE CASCADE,
    project_id      UUID                    REFERENCES app.projects (id) ON DELETE SET NULL, -- optional grouping
    customer_id     UUID                    NOT NULL REFERENCES public.customers (id) ON DELETE RESTRICT,
    vehicle_id      UUID                    NOT NULL REFERENCES public.vehicles (id) ON DELETE RESTRICT,
    status          app.work_order_status   NOT NULL DEFAULT 'draft',
    priority        app.work_order_priority NOT NULL DEFAULT 'normal',
    title           TEXT                    NOT NULL,
    description     TEXT,
    -- lifecycle timestamps
    opened_at       TIMESTAMPTZ             NOT NULL DEFAULT now(),
    scheduled_at    TIMESTAMPTZ,
    started_at      TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ,
    closed_at       TIMESTAMPTZ,
    -- audit
    created_by      UUID                    NOT NULL REFERENCES app.users (id) ON DELETE RESTRICT,
    updated_by      UUID                    REFERENCES app.users (id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ             NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ             NOT NULL DEFAULT now(),
    -- quick totals (optional denormalization)
    subtotal_cents  BIGINT                  NOT NULL DEFAULT 0,
    tax_cents       BIGINT                  NOT NULL DEFAULT 0,
    total_cents     BIGINT                  NOT NULL DEFAULT 0
);
CREATE INDEX idx_work_orders_org ON app.work_orders (organization_id);
CREATE INDEX idx_work_orders_customer ON app.work_orders (customer_id);
CREATE INDEX idx_work_orders_vehicle ON app.work_orders (vehicle_id);
CREATE INDEX idx_work_orders_status ON app.work_orders (status);
CREATE INDEX idx_work_orders_priority ON app.work_orders (priority);

-- Line items for labor/parts
CREATE TABLE app.work_order_items
(
    id               UUID PRIMARY KEY            DEFAULT gen_random_uuid(),
    work_order_id    UUID               NOT NULL REFERENCES app.work_orders (id) ON DELETE CASCADE,
    item_type        app.line_item_type NOT NULL,
    sku              TEXT,
    name             TEXT               NOT NULL,
    qty              NUMERIC(12, 2)     NOT NULL DEFAULT 1,
    unit_price_cents BIGINT             NOT NULL DEFAULT 0,
    tax_rate_pct     NUMERIC(5, 2)      NOT NULL DEFAULT 0,
    position         INT                NOT NULL DEFAULT 0,
    created_at       TIMESTAMPTZ        NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ        NOT NULL DEFAULT now()
);
CREATE INDEX idx_items_work_order ON app.work_order_items (work_order_id);

-- Event log for status changes and communications (for automation & audit)
CREATE TABLE app.work_order_events
(
    id            UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    work_order_id UUID        NOT NULL REFERENCES app.work_orders (id) ON DELETE CASCADE,
    event_type    TEXT        NOT NULL, -- e.g., status_changed, note_added, photo_uploaded, customer_notified
    from_status   app.work_order_status,
    to_status     app.work_order_status,
    message       TEXT,
    created_by    UUID        REFERENCES app.users (id) ON DELETE SET NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_events_work_order ON app.work_order_events (work_order_id);
CREATE INDEX idx_events_type ON app.work_order_events (event_type);

-- Outbound notification log (so you can automate and show history)

CREATE TABLE app.notification_logs
(
    id              UUID PRIMARY KEY            DEFAULT gen_random_uuid(),
    organization_id UUID               NOT NULL REFERENCES app.organizations (id) ON DELETE CASCADE,
    work_order_id   UUID REFERENCES app.work_orders (id) ON DELETE CASCADE,
    event_id        UUID               REFERENCES app.work_order_events (id) ON DELETE SET NULL,
    customer_id     UUID               REFERENCES public.customers (id) ON DELETE SET NULL,
    channel         app.notify_channel NOT NULL,
    recipient       TEXT               NOT NULL, -- phone/email/endpoint
    template_key    TEXT,                        -- for idempotency & analytics
    sent_at         TIMESTAMPTZ        NOT NULL DEFAULT now(),
    meta            JSONB              NOT NULL DEFAULT '{}'
);
CREATE INDEX idx_notifications_org ON app.notification_logs (organization_id);
CREATE INDEX idx_notifications_work_order ON app.notification_logs (work_order_id);

-- =========================
-- 4) Schedule (Appointments / Calendar)
-- =========================

CREATE TABLE app.appointments
(
    id              UUID PRIMARY KEY                DEFAULT gen_random_uuid(),
    organization_id UUID                   NOT NULL REFERENCES app.organizations (id) ON DELETE CASCADE,
    customer_id     UUID                   NOT NULL REFERENCES public.customers (id) ON DELETE RESTRICT,
    vehicle_id      UUID                   REFERENCES public.vehicles (id) ON DELETE SET NULL,
    title           TEXT                   NOT NULL, -- e.g., "Initial inspection"
    notes           TEXT,
    status          app.appointment_status NOT NULL DEFAULT 'pending',
    start_time      TIMESTAMPTZ            NOT NULL,
    end_time        TIMESTAMPTZ            NOT NULL,
    created_by      UUID                   NOT NULL REFERENCES app.users (id) ON DELETE RESTRICT,
    created_at      TIMESTAMPTZ            NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ            NOT NULL DEFAULT now(),
    UNIQUE (organization_id, start_time, end_time, customer_id)
);
CREATE INDEX idx_appointments_org_time ON app.appointments (organization_id, start_time);

-- Link an appointment to a work order once the customer agrees.
CREATE TABLE app.appointment_work_orders
(
    appointment_id UUID PRIMARY KEY REFERENCES app.appointments (id) ON DELETE CASCADE,
    work_order_id  UUID UNIQUE NOT NULL REFERENCES app.work_orders (id) ON DELETE CASCADE,
    linked_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- =========================
-- 5) Useful policies / guardrails you can add later
-- =========================
-- - Row Level Security (RLS) scoped by organization_id for multi-tenant isolation.
-- - Materialized view or trigger to maintain work_orders totals from items.
-- - Webhook consumer to sync Stack Auth profile/email → users.

-- (optional but recommended) Make sure your DB/role can see it by default
-- Choose ONE of these:

-- A) Per-session:
SET search_path = public, app;

-- -- B) Per-role (persists for that role):
-- ALTER ROLE your_app_db_role IN DATABASE your_database
--   SET search_path = public, app;

-- C) Per-database (affects all roles unless overridden):
ALTER DATABASE engine_care
    SET search_path = public, app;


-- =========
-- Helpers
-- =========
CREATE OR REPLACE FUNCTION app.current_user_id()
    RETURNS uuid
    LANGUAGE sql
    STABLE AS
$$
SELECT NULLIF(current_setting('app.user_id', true), '')::uuid;
$$;

CREATE OR REPLACE FUNCTION app.current_org_id()
    RETURNS uuid
    LANGUAGE sql
    STABLE AS
$$
SELECT NULLIF(current_setting('app.organization_id', true), '')::uuid;
$$;

CREATE OR REPLACE FUNCTION app.is_org_member(org uuid)
    RETURNS boolean
    LANGUAGE sql
    STABLE AS
$$
SELECT EXISTS (SELECT 1
               FROM organization_members m
               WHERE m.organization_id = org
                 AND m.user_id = app.current_user_id());
$$;

CREATE OR REPLACE FUNCTION app.has_org_role(org uuid, roles text[])
    RETURNS boolean
    LANGUAGE sql
    STABLE AS
$$
SELECT EXISTS (SELECT 1
               FROM organization_members m
               WHERE m.organization_id = org
                 AND m.user_id = app.current_user_id()
                 AND m.role::text = ANY (roles));
$$;

-- inside a tx, this is an example to validate that the settings work:
-- BEGIN;
-- SET LOCAL app.user_id = '00000000-0000-0000-0000-000000000000';
-- SET LOCAL app.organization_id = '11111111-1111-1111-1111-111111111111';
-- SELECT app.current_user_id(), app.current_org_id();
-- COMMIT;


-- =========
-- Enable RLS
-- =========
ALTER TABLE app.organizations
    ENABLE ROW LEVEL SECURITY;
ALTER TABLE app.organization_members
    ENABLE ROW LEVEL SECURITY;
ALTER TABLE app.projects
    ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.customers
    ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.vehicles
    ENABLE ROW LEVEL SECURITY;
ALTER TABLE app.work_orders
    ENABLE ROW LEVEL SECURITY;
ALTER TABLE app.work_order_items
    ENABLE ROW LEVEL SECURITY;
ALTER TABLE app.work_order_events
    ENABLE ROW LEVEL SECURITY;
ALTER TABLE app.notification_logs
    ENABLE ROW LEVEL SECURITY;
ALTER TABLE app.appointments
    ENABLE ROW LEVEL SECURITY;
ALTER TABLE app.appointment_work_orders
    ENABLE ROW LEVEL SECURITY;

-- =========
-- Policies
-- =========
-- Organizations: members can see their org; only owner/admin can modify.
DROP POLICY IF EXISTS org_select ON app.organizations;
CREATE POLICY org_select ON app.organizations
    FOR SELECT
    USING (id = app.current_org_id() AND app.is_org_member(id));

DROP POLICY IF EXISTS org_modify ON app.organizations;
CREATE POLICY org_modify ON app.organizations
    FOR UPDATE
    USING (id = app.current_org_id() AND app.has_org_role(id, ARRAY ['owner','admin']));

CREATE POLICY org_delete ON app.organizations
    FOR DELETE
    USING (id = app.current_org_id() AND app.has_org_role(id, ARRAY ['owner','admin']));

DROP POLICY IF EXISTS org_insert ON app.organizations;
CREATE POLICY org_insert ON app.organizations
    FOR INSERT
    WITH CHECK (true);
-- allow system flows; typically orgs are created by service role

-- Organization members
DROP POLICY IF EXISTS orgmem_select ON app.organization_members;
CREATE POLICY orgmem_select ON app.organization_members
    FOR SELECT
    USING (organization_id = app.current_org_id() AND app.is_org_member(organization_id));

-- Insert: only owner/admin can create memberships (INSERT has ONLY WITH CHECK)
DROP POLICY IF EXISTS orgmem_insert ON app.organization_members;
CREATE POLICY orgmem_insert ON app.organization_members
    FOR INSERT
    WITH CHECK (
    organization_id = app.current_org_id()
        AND app.has_org_role(organization_id, ARRAY ['owner','admin'])
    );

-- Update: owner/admin can update
DROP POLICY IF EXISTS orgmem_update ON app.organization_members;
CREATE POLICY orgmem_update ON app.organization_members
    FOR UPDATE
    USING (
    organization_id = app.current_org_id()
        AND app.has_org_role(organization_id, ARRAY ['owner','admin'])
    )
    WITH CHECK (
    organization_id = app.current_org_id()
        AND app.has_org_role(organization_id, ARRAY ['owner','admin'])
    );

-- Delete: owner/admin can delete
DROP POLICY IF EXISTS orgmem_delete ON app.organization_members;
CREATE POLICY orgmem_delete ON app.organization_members
    FOR DELETE
    USING (
    organization_id = app.current_org_id()
        AND app.has_org_role(organization_id, ARRAY ['owner','admin'])
    );

-- Generic “org-scoped” template for the rest
ALTER TABLE app.work_order_items
    ADD COLUMN organization_id uuid;
ALTER TABLE app.work_order_events
    ADD COLUMN organization_id uuid;

-- backfill
UPDATE app.work_order_items i
SET organization_id = w.organization_id
FROM app.work_orders w
WHERE w.id = i.work_order_id;

UPDATE app.work_order_events e
SET organization_id = w.organization_id
FROM app.work_orders w
WHERE w.id = e.work_order_id;

-- not null after backfill
ALTER TABLE app.work_order_items
    ALTER COLUMN organization_id SET NOT NULL;
ALTER TABLE app.work_order_events
    ALTER COLUMN organization_id SET NOT NULL;

-- keep it in sync going forward
CREATE OR REPLACE FUNCTION app.sync_child_org_from_work_order()
    RETURNS trigger
    LANGUAGE plpgsql AS
$$
BEGIN
    SELECT organization_id
    INTO NEW.organization_id
    FROM app.work_orders
    WHERE id = NEW.work_order_id;
    RETURN NEW;
END
$$;

DROP TRIGGER IF EXISTS trg_items_org_sync ON app.work_order_items;
CREATE TRIGGER trg_items_org_sync
    BEFORE INSERT OR UPDATE OF work_order_id
    ON app.work_order_items
    FOR EACH ROW
EXECUTE FUNCTION app.sync_child_org_from_work_order();

DROP TRIGGER IF EXISTS trg_events_org_sync ON app.work_order_events;
CREATE TRIGGER trg_events_org_sync
    BEFORE INSERT OR UPDATE OF work_order_id
    ON app.work_order_events
    FOR EACH ROW
EXECUTE FUNCTION app.sync_child_org_from_work_order();


CREATE OR REPLACE FUNCTION app.recalc_work_order_totals(p_work_order_id uuid)
    RETURNS void
    LANGUAGE plpgsql AS
$$
DECLARE
    v_subtotal BIGINT := 0;
    v_tax      BIGINT := 0;
    v_total    BIGINT := 0;
BEGIN
    SELECT COALESCE(SUM((unit_price_cents * qty)::bigint), 0),
           COALESCE(SUM(((unit_price_cents * qty) * (tax_rate_pct / 100.0))::bigint), 0)
    INTO v_subtotal, v_tax
    FROM app.work_order_items
    WHERE work_order_id = p_work_order_id;

    v_total := v_subtotal + v_tax;

    UPDATE app.work_orders
    SET subtotal_cents = v_subtotal,
        tax_cents      = v_tax,
        total_cents    = v_total,
        updated_at     = now()
    WHERE id = p_work_order_id;
END
$$;

CREATE OR REPLACE FUNCTION app.work_order_items_recalc_trg()
    RETURNS trigger
    LANGUAGE plpgsql AS
$$
BEGIN
    PERFORM app.recalc_work_order_totals(
            COALESCE(NEW.work_order_id, OLD.work_order_id)
            );
    RETURN COALESCE(NEW, OLD);
END
$$;

DROP TRIGGER IF EXISTS trg_items_recalc_ins ON app.work_order_items;
CREATE TRIGGER trg_items_recalc_ins
    AFTER INSERT
    ON app.work_order_items
    FOR EACH ROW
EXECUTE FUNCTION app.work_order_items_recalc_trg();

DROP TRIGGER IF EXISTS trg_items_recalc_upd ON app.work_order_items;
CREATE TRIGGER trg_items_recalc_upd
    AFTER UPDATE
    ON app.work_order_items
    FOR EACH ROW
EXECUTE FUNCTION app.work_order_items_recalc_trg();

DROP TRIGGER IF EXISTS trg_items_recalc_del ON app.work_order_items;
CREATE TRIGGER trg_items_recalc_del
    AFTER DELETE
    ON app.work_order_items
    FOR EACH ROW
EXECUTE FUNCTION app.work_order_items_recalc_trg();


CREATE OR REPLACE FUNCTION app.work_orders_status_event_trg()
    RETURNS trigger
    LANGUAGE plpgsql AS
$$
BEGIN
    IF TG_OP = 'UPDATE' AND NEW.status IS DISTINCT FROM OLD.status THEN
        INSERT INTO app.work_order_events (id, work_order_id, event_type, from_status, to_status, message, created_by)
        VALUES (gen_random_uuid(), NEW.id, 'status_changed', OLD.status, NEW.status,
                format('Status changed %s → %s', OLD.status, NEW.status),
                NEW.updated_by);
    END IF;
    RETURN NEW;
END
$$;

DROP TRIGGER IF EXISTS trg_wo_status_event ON app.work_orders;
CREATE TRIGGER trg_wo_status_event
    AFTER UPDATE
    ON app.work_orders
    FOR EACH ROW
EXECUTE FUNCTION app.work_orders_status_event_trg();



-- appointment_work_orders already has PK=appointment_id and UNIQUE(work_order_id).
-- Optional: Prevent linking across orgs.
CREATE OR REPLACE FUNCTION app.enforce_same_org_appointment_link()
    RETURNS trigger
    LANGUAGE plpgsql AS
$$
DECLARE
    ao uuid; wo uuid;
BEGIN
    SELECT organization_id INTO ao FROM appointments WHERE id = NEW.appointment_id;
    SELECT organization_id INTO wo FROM work_orders WHERE id = NEW.work_order_id;

    IF ao IS NULL OR wo IS NULL OR ao <> wo THEN
        RAISE EXCEPTION 'Appointment and Work Order must belong to the same organization';
    END IF;

    RETURN NEW;
END
$$;

DROP TRIGGER IF EXISTS trg_appt_link_org ON app.appointment_work_orders;
CREATE TRIGGER trg_appt_link_org
    BEFORE INSERT
    ON app.appointment_work_orders
    FOR EACH ROW
EXECUTE FUNCTION app.enforce_same_org_appointment_link();



-- Allow only certain roles to write to work_orders
-- INSERT policy (INSERT can only use WITH CHECK)
DROP POLICY IF EXISTS wo_insert ON app.work_orders;
CREATE POLICY wo_insert ON app.work_orders
    FOR INSERT
    WITH CHECK (
    organization_id = app.current_org_id()
        AND app.has_org_role(organization_id, ARRAY ['owner','admin','manager','mechanic'])
    );

-- UPDATE policy (can use both USING + WITH CHECK)
DROP POLICY IF EXISTS wo_update ON app.work_orders;
CREATE POLICY wo_update ON app.work_orders
    FOR UPDATE
    USING (
    organization_id = app.current_org_id()
        AND app.has_org_role(organization_id, ARRAY ['owner','admin','manager','mechanic'])
    )
    WITH CHECK (
    organization_id = app.current_org_id()
        AND app.has_org_role(organization_id, ARRAY ['owner','admin','manager','mechanic'])
    );

-- DELETE policy (only USING)
DROP POLICY IF EXISTS wo_delete ON app.work_orders;
CREATE POLICY wo_delete ON app.work_orders
    FOR DELETE
    USING (
    organization_id = app.current_org_id()
        AND app.has_org_role(organization_id, ARRAY ['owner','admin','manager','mechanic'])
    );

-- Work Order Items
DROP POLICY IF EXISTS woi_select ON app.work_order_items;
CREATE POLICY woi_select ON app.work_order_items
    FOR SELECT
    USING (organization_id = app.current_org_id() AND app.is_org_member(organization_id));
DROP POLICY IF EXISTS woi_insert ON app.work_order_items;
CREATE POLICY woi_insert ON app.work_order_items
    FOR INSERT
    WITH CHECK (
    organization_id = app.current_org_id()
        AND app.has_org_role(organization_id, ARRAY ['owner','admin','manager','mechanic'])
    );
DROP POLICY IF EXISTS woi_update ON app.work_order_items;
CREATE POLICY woi_update ON app.work_order_items
    FOR UPDATE
    USING (
    organization_id = app.current_org_id()
        AND app.has_org_role(organization_id, ARRAY ['owner','admin','manager','mechanic'])
    )
    WITH CHECK (
    organization_id = app.current_org_id()
        AND app.has_org_role(organization_id, ARRAY ['owner','admin','manager','mechanic'])
    );
DROP POLICY IF EXISTS woi_delete ON app.work_order_items;
CREATE POLICY woi_delete ON app.work_order_items
    FOR DELETE
    USING (
    organization_id = app.current_org_id()
        AND app.has_org_role(organization_id, ARRAY ['owner','admin','manager','mechanic'])
    );

-- Work Order Events
DROP POLICY IF EXISTS woe_select ON app.work_order_events;
CREATE POLICY woe_select ON app.work_order_events
    FOR SELECT
    USING (organization_id = app.current_org_id() AND app.is_org_member(organization_id));
DROP POLICY IF EXISTS woe_insert ON app.work_order_events;
CREATE POLICY woe_insert ON app.work_order_events
    FOR INSERT
    WITH CHECK (
    organization_id = app.current_org_id()
        AND app.has_org_role(organization_id, ARRAY ['owner','admin','manager','mechanic'])
    );
DROP POLICY IF EXISTS woe_update ON app.work_order_events;
CREATE POLICY woe_update ON app.work_order_events
    FOR UPDATE
    USING (
    organization_id = app.current_org_id()
        AND app.has_org_role(organization_id, ARRAY ['owner','admin','manager','mechanic'])
    )
    WITH CHECK (
    organization_id = app.current_org_id()
        AND app.has_org_role(organization_id, ARRAY ['owner','admin','manager','mechanic'])
    );
DROP POLICY IF EXISTS woe_delete ON app.work_order_events;
CREATE POLICY woe_delete ON app.work_order_events
    FOR DELETE
    USING (
    organization_id = app.current_org_id()
        AND app.has_org_role(organization_id, ARRAY ['owner','admin','manager','mechanic'])
    );

-- Notification Logs
DROP POLICY IF EXISTS nl_select ON app.notification_logs;
CREATE POLICY nl_select ON app.notification_logs
    FOR SELECT
    USING (organization_id = app.current_org_id() AND app.is_org_member(organization_id));
DROP POLICY IF EXISTS nl_insert ON app.notification_logs;
CREATE POLICY nl_insert ON app.notification_logs
    FOR INSERT
    WITH CHECK (
    organization_id = app.current_org_id()
        AND app.has_org_role(organization_id, ARRAY ['owner','admin','manager','mechanic'])
    );
DROP POLICY IF EXISTS nl_update ON app.notification_logs;
CREATE POLICY nl_update ON app.notification_logs
    FOR UPDATE
    USING (
    organization_id = app.current_org_id()
        AND app.has_org_role(organization_id, ARRAY ['owner','admin','manager','mechanic'])
    )
    WITH CHECK (
    organization_id = app.current_org_id()
        AND app.has_org_role(organization_id, ARRAY ['owner','admin','manager','mechanic'])
    );
DROP POLICY IF EXISTS nl_delete ON app.notification_logs;
CREATE POLICY nl_delete ON app.notification_logs
    FOR DELETE
    USING (
    organization_id = app.current_org_id()
        AND app.has_org_role(organization_id, ARRAY ['owner','admin','manager','mechanic'])
    );

-- Appointments
DROP POLICY IF EXISTS appt_select ON app.appointments;
CREATE POLICY appt_select ON app.appointments
    FOR SELECT
    USING (organization_id = app.current_org_id() AND app.is_org_member(organization_id));
DROP POLICY IF EXISTS appt_insert ON app.appointments;
CREATE POLICY appt_insert ON app.appointments
    FOR INSERT
    WITH CHECK (
    organization_id = app.current_org_id()
        AND app.has_org_role(organization_id, ARRAY ['owner','admin','manager','mechanic'])
    );
DROP POLICY IF EXISTS appt_update ON app.appointments;
CREATE POLICY appt_update ON app.appointments
    FOR UPDATE
    USING (
    organization_id = app.current_org_id()
        AND app.has_org_role(organization_id, ARRAY ['owner','admin','manager','mechanic'])
    )
    WITH CHECK (
    organization_id = app.current_org_id()
        AND app.has_org_role(organization_id, ARRAY ['owner','admin','manager','mechanic'])
    );
DROP POLICY IF EXISTS appt_delete ON app.appointments;
CREATE POLICY appt_delete ON app.appointments
    FOR DELETE
    USING (
    organization_id = app.current_org_id()
        AND app.has_org_role(organization_id, ARRAY ['owner','admin','manager','mechanic'])
    );

-- Appointment Work Orders
DROP POLICY IF EXISTS apptwo_select ON app.appointment_work_orders;
CREATE POLICY apptwo_select ON app.appointment_work_orders
    FOR SELECT
    USING (
    EXISTS (SELECT 1
           FROM app.appointments a
           WHERE a.id = appointment_id
             AND a.organization_id = app.current_org_id()
             AND app.is_org_member(a.organization_id))
    );
DROP POLICY IF EXISTS apptwo_insert ON app.appointment_work_orders;
CREATE POLICY apptwo_insert ON app.appointment_work_orders
    FOR INSERT
    WITH CHECK (
    EXISTS (SELECT 1
           FROM app.appointments a
           WHERE a.id = appointment_id
             AND a.organization_id = app.current_org_id()
             AND app.has_org_role(a.organization_id, ARRAY ['owner','admin','manager','mechanic']))
    );
DROP POLICY IF EXISTS apptwo_update ON app.appointment_work_orders;
CREATE POLICY apptwo_update ON app.appointment_work_orders
    FOR UPDATE
    USING (
    EXISTS (SELECT 1
           FROM app.appointments a
           WHERE a.id = appointment_id
             AND a.organization_id = app.current_org_id()
             AND app.has_org_role(a.organization_id, ARRAY ['owner','admin','manager','mechanic']))
    )
    WITH CHECK (
    EXISTS (SELECT 1
           FROM app.appointments a
           WHERE a.id = appointment_id
             AND a.organization_id = app.current_org_id()
             AND app.has_org_role(a.organization_id, ARRAY ['owner','admin','manager','mechanic']))
    );
DROP POLICY IF EXISTS apptwo_delete ON app.appointment_work_orders;
CREATE POLICY apptwo_delete ON app.appointment_work_orders
    FOR DELETE
    USING (
    EXISTS (SELECT 1
           FROM app.appointments a
           WHERE a.id = appointment_id
             AND a.organization_id = app.current_org_id()
             AND app.has_org_role(a.organization_id, ARRAY ['owner','admin','manager','mechanic']))
    );
