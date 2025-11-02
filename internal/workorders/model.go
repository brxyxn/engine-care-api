package workorders

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// Status represents the work order status
type Status string

const (
	StatusDraft            Status = "draft"
	StatusNew              Status = "new"
	StatusScheduled        Status = "scheduled"
	StatusAwaitingCustomer Status = "awaiting_customer"
	StatusInProgress       Status = "in_progress"
	StatusWaitingParts     Status = "waiting_parts"
	StatusAwaitingApproval Status = "awaiting_approval"
	StatusReadyForPickup   Status = "ready_for_pickup"
	StatusReadyForDeliver  Status = "ready_for_deliver"
	StatusEnRoute          Status = "en_route"
	StatusCompleted        Status = "completed"
	StatusCanceled         Status = "canceled"
)

// Priority represents the work order status
type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityNormal Priority = "normal"
	PriorityHigh   Priority = "high"
	PriorityUrgent Priority = "urgent"
)

type LineItemType string

const (
	LineItemTypeLabor LineItemType = "labor"
	LineItemTypePart  LineItemType = "part"
	LineItemTypeFee   LineItemType = "fee"
	LineItemTypeOther LineItemType = "other"
)

type WorkOrder struct {
	bun.BaseModel `bun:"table:work_orders,alias:wo"`

	ID             uuid.UUID  `bun:"id,pk,default:gen_random_uuid()" json:"id"`
	OrganizationID uuid.UUID  `bun:"organization_id,notnull" json:"organization_id"`
	ProjectID      *uuid.UUID `bun:"project_id" json:"project_id,omitempty"`
	CustomerID     uuid.UUID  `bun:"customer_id,notnull" json:"customer_id"`
	VehicleID      uuid.UUID  `bun:"vehicle_id,notnull" json:"vehicle_id"`

	Status   Status   `bun:"status,type:work_order_status,notnull,default:draft" json:"status"`
	Priority Priority `bun:"priority,type:work_order_priority,notnull,default:normal" json:"priority"`

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

type Item struct {
	bun.BaseModel `bun:"table:work_order_items,alias:woi"`

	ID             uuid.UUID    `bun:"id,pk,default:gen_random_uuid()" json:"id"`
	WorkOrderID    uuid.UUID    `bun:"work_order_id,notnull" json:"work_order_id"`
	ItemType       LineItemType `bun:"item_type,type:line_item_type,notnull" json:"item_type"`
	SKU            *string      `bun:"sku" json:"sku,omitempty"`
	Name           string       `bun:"name,notnull" json:"name"`
	Qty            float64      `bun:"qty,notnull,default:1" json:"qty"`
	UnitPriceCents int64        `bun:"unit_price_cents,notnull,default:0" json:"unit_price_cents"`
	TaxRatePct     float64      `bun:"tax_rate_pct,notnull,default:0" json:"tax_rate_pct"`
	Position       int          `bun:"position,notnull,default:0" json:"position"`
	CreatedAt      time.Time    `bun:"created_at,notnull,default:now()" json:"created_at"`
	UpdatedAt      time.Time    `bun:"updated_at,notnull,default:now()" json:"updated_at"`
}

type Event struct {
	bun.BaseModel `bun:"table:work_order_events,alias:woe"`

	ID          uuid.UUID  `bun:"id,pk,default:gen_random_uuid()" json:"id"`
	WorkOrderID uuid.UUID  `bun:"work_order_id,notnull" json:"work_order_id"`
	EventType   string     `bun:"event_type,notnull" json:"event_type"`
	FromStatus  *Status    `bun:"from_status,type:work_order_status" json:"from_status,omitempty"`
	ToStatus    *Status    `bun:"to_status,type:work_order_status" json:"to_status,omitempty"`
	Message     *string    `bun:"message" json:"message,omitempty"`
	CreatedBy   *uuid.UUID `bun:"created_by" json:"created_by,omitempty"`
	CreatedAt   time.Time  `bun:"created_at,notnull,default:now()" json:"created_at"`
}
