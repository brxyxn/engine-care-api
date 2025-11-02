package appointments

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Status string

const (
	StatusPending   Status = "pending"
	StatusConfirmed Status = "confirmed"
	StatusCancelled Status = "cancelled"
	StatusNoShow    Status = "no_show"
	StatusCompleted Status = "completed"
)

type Appointment struct {
	bun.BaseModel `bun:"table:appointments,alias:a"`

	ID             uuid.UUID  `bun:"id,pk,default:gen_random_uuid()" json:"id"`
	OrganizationID uuid.UUID  `bun:"organization_id,notnull" json:"organization_id"`
	CustomerID     uuid.UUID  `bun:"customer_id,notnull" json:"customer_id"`
	VehicleID      *uuid.UUID `bun:"vehicle_id" json:"vehicle_id,omitempty"`
	Title          string     `bun:"title,notnull" json:"title"`
	Notes          *string    `bun:"notes" json:"notes,omitempty"`
	Status         Status     `bun:"status,type:appointment_status,notnull,default:pending" json:"status"`
	StartTime      time.Time  `bun:"start_time,notnull" json:"start_time"`
	EndTime        time.Time  `bun:"end_time,notnull" json:"end_time"`
	CreatedBy      uuid.UUID  `bun:"created_by,notnull" json:"created_by"`
	CreatedAt      time.Time  `bun:"created_at,notnull,default:now()" json:"created_at"`
	UpdatedAt      time.Time  `bun:"updated_at,notnull,default:now()" json:"updated_at"`
}

type AppointmentWorkOrder struct {
	bun.BaseModel `bun:"table:appointment_work_orders,alias:awo"`

	AppointmentID uuid.UUID `bun:"appointment_id,pk" json:"appointment_id"`
	WorkOrderID   uuid.UUID `bun:"work_order_id,unique,notnull" json:"work_order_id"`
	LinkedAt      time.Time `bun:"linked_at,notnull,default:now()" json:"linked_at"`
}
