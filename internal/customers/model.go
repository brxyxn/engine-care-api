package customers

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Customer struct {
	bun.BaseModel `bun:"table:customers,alias:c"`

	ID             uuid.UUID  `bun:"id,pk,default:gen_random_uuid()" json:"id"`
	OrganizationID uuid.UUID  `bun:"organization_id,notnull" json:"organization_id"`
	FullName       string     `bun:"full_name,notnull" json:"full_name"`
	Email          *string    `bun:"email" json:"email,omitempty"`
	Notes          *string    `bun:"notes" json:"notes,omitempty"`
	CreatedBy      *uuid.UUID `bun:"created_by" json:"created_by,omitempty"`
	UpdatedBy      *uuid.UUID `bun:"updated_by" json:"updated_by,omitempty"`
	CreatedAt      time.Time  `bun:"created_at,notnull,default:now()" json:"created_at"`
	UpdatedAt      time.Time  `bun:"updated_at,notnull,default:now()" json:"updated_at"`
}

type CustomerPhoneNumber struct {
	bun.BaseModel `bun:"table:customer_phone_numbers,alias:cpn"`

	ID            uuid.UUID  `bun:"id,pk,default:gen_random_uuid()" json:"id"`
	CustomerID    *uuid.UUID `bun:"customer_id" json:"customer_id,omitempty"`
	PhoneNumberID uuid.UUID  `bun:"phone_number_id,notnull" json:"phone_number_id"`
	IsPrimary     bool       `bun:"is_primary,notnull,default:false" json:"is_primary"`
	CreatedAt     time.Time  `bun:"created_at,notnull,default:now()" json:"created_at"`
}
