package vehicles

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Vehicle struct {
	bun.BaseModel `bun:"table:vehicles,alias:v"`

	ID             uuid.UUID `bun:"id,pk,default:gen_random_uuid()" json:"id"`
	OrganizationID uuid.UUID `bun:"organization_id,notnull" json:"organization_id"`
	CustomerID     uuid.UUID `bun:"customer_id,notnull" json:"customer_id"`
	VIN            *string   `bun:"vin" json:"vin,omitempty"`
	PlateNumber    *string   `bun:"plate_number" json:"plate_number,omitempty"`
	Make           *string   `bun:"make" json:"make,omitempty"`
	Model          *string   `bun:"model" json:"model,omitempty"`
	Year           *int      `bun:"year" json:"year,omitempty"`
	Color          *string   `bun:"color" json:"color,omitempty"`
	MileageKM      *int      `bun:"mileage_km" json:"mileage_km,omitempty"`
	CreatedAt      time.Time `bun:"created_at,notnull,default:now()" json:"created_at"`
	UpdatedAt      time.Time `bun:"updated_at,notnull,default:now()" json:"updated_at"`
}
