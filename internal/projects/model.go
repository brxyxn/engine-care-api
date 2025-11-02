package projects

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Project struct {
	bun.BaseModel `bun:"table:projects,alias:p"`

	ID             uuid.UUID `bun:"id,pk,default:gen_random_uuid()" json:"id"`
	OrganizationID uuid.UUID `bun:"organization_id,notnull" json:"organization_id"`
	Name           string    `bun:"name,notnull" json:"name"`
	Description    *string   `bun:"description" json:"description,omitempty"`
	CreatedAt      time.Time `bun:"created_at,notnull,default:now()" json:"created_at"`
	UpdatedAt      time.Time `bun:"updated_at,notnull,default:now()" json:"updated_at"`
}
