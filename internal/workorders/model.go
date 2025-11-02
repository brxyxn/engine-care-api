package workorders

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type WorkOrderStatus string
type WorkOrderPriority string

const (
	WorkOrderStatusDraft    WorkOrderStatus   = "draft"
	WorkOrderPriorityNormal WorkOrderPriority = "normal"
)

type WorkOrder struct {
	//id              UUID PRIMARY KEY             DEFAULT gen_random_uuid(),
	//organization_id UUID                NOT NULL REFERENCES organizations (id) ON DELETE CASCADE,
	//project_id      UUID                REFERENCES projects (id) ON DELETE SET NULL, -- optional grouping
	//customer_id     UUID                NOT NULL REFERENCES customers (id) ON DELETE RESTRICT,
	//vehicle_id      UUID                NOT NULL REFERENCES vehicles (id) ON DELETE RESTRICT,
	//status          work_order_status   NOT NULL DEFAULT 'draft',
	//priority        work_order_priority NOT NULL DEFAULT 'normal',
	//title           TEXT                NOT NULL,
	//description     TEXT,
	//-- lifecycle timestamps
	//opened_at       TIMESTAMPTZ         NOT NULL DEFAULT now(),
	//scheduled_at    TIMESTAMPTZ,
	//started_at      TIMESTAMPTZ,
	//completed_at    TIMESTAMPTZ,
	//closed_at       TIMESTAMPTZ,
	//-- audit
	//created_by      UUID                NOT NULL REFERENCES app_users (id) ON DELETE RESTRICT,
	//updated_by      UUID                REFERENCES app_users (id) ON DELETE SET NULL,
	//created_at      TIMESTAMPTZ         NOT NULL DEFAULT now(),
	//updated_at      TIMESTAMPTZ         NOT NULL DEFAULT now(),
	//-- quick totals (optional denormalization)
	//subtotal_cents  BIGINT              NOT NULL DEFAULT 0,
	//tax_cents       BIGINT              NOT NULL DEFAULT 0,
	//total_cents     BIGINT              NOT NULL DEFAULT 0
	bun.BaseModel `bun:"table:work_orders,alias:wo"`

	ID             uuid.UUID  `bun:"id,pk,default:gen_random_uuid()" json:"id"`
	OrganizationID uuid.UUID  `bun:"organization_id,notnull" json:"organization_id"`
	ProjectID      *uuid.UUID `bun:"project_id" json:"project_id,omitempty"`
	CustomerID     uuid.UUID  `bun:"customer_id,notnull" json:"customer_id"`
	VehicleID      uuid.UUID  `bun:"vehicle_id,notnull" json:"vehicle_id"`

	Status   WorkOrderStatus   `bun:"status,type:work_order_status,notnull,default:draft" json:"status"`
	Priority WorkOrderPriority `bun:"priority,type:work_order_priority,notnull,default:normal" json:"priority"`

	Title       string  `bun:"title,notnull" json:"title"`
	Description *string `bun:"description" json:"description,omitempty"`

	OpenedAt    time.Time  `bun:"opened_at,notnull,default:now()" json:"opened_at"`
	ScheduledAt *time.Time `bun:"scheduled_at" json:"scheduled_at,omitempty"`
	StartedAt   *time.Time `bun:"started_at" json:"started_at,omitempty"`
	CompletedAt *time.Time `bun:"completed_at" json:"completed_at,omitempty"`
	ClosedAt    *time.Time `bun:"closed_at" json:"closed_at,omitempty"`

	CreatedBy uuid.UUID  `bun:"created_by,notnull" json:"created_by"`
	UpdatedBy *uuid.UUID `bun:"updated_by" json:"updated_by,omitempty"`
	CreatedAt time.Time  `bun:"created_at,notnull,default:now()" json:"created_at"`
	UpdatedAt time.Time  `bun:"updated_at,notnull,default:now()" json:"updated_at"`

	SubtotalCents int64 `bun:"subtotal_cents,notnull,default:0" json:"subtotal_cents"`
	TaxCents      int64 `bun:"tax_cents,notnull,default:0" json:"tax_cents"`
	TotalCents    int64 `bun:"total_cents,notnull,default:0" json:"total_cents"`
}
