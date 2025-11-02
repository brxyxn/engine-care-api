package organizations

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// OrgRole mirrors app.org_role enum
type OrgRole string

const (
	OrgRoleOwner    OrgRole = "owner"
	OrgRoleAdmin    OrgRole = "admin"
	OrgRoleManager  OrgRole = "manager"
	OrgRoleMechanic OrgRole = "mechanic"
	OrgRoleViewer   OrgRole = "viewer"
)

type Organization struct {
	bun.BaseModel `bun:"table:organizations,alias:org"`

	ID          uuid.UUID `bun:"id,pk,default:gen_random_uuid()" json:"id"`
	Name        string    `bun:"name,notnull" json:"name"`
	StackTeamID *string   `bun:"stack_team_id" json:"stack_team_id,omitempty"`
	CreatedAt   time.Time `bun:"created_at,notnull,default:now()" json:"created_at"`
	UpdatedAt   time.Time `bun:"updated_at,notnull,default:now()" json:"updated_at"`
}

type OrganizationMember struct {
	bun.BaseModel `bun:"table:organization_members,alias:om"`

	ID             uuid.UUID  `bun:"id,pk,default:gen_random_uuid()" json:"id"`
	OrganizationID uuid.UUID  `bun:"organization_id,notnull" json:"organization_id"`
	UserID         uuid.UUID  `bun:"user_id,notnull" json:"user_id"`
	Role           OrgRole    `bun:"role,type:org_role,notnull,default:viewer" json:"role"`
	InvitedBy      *uuid.UUID `bun:"invited_by" json:"invited_by,omitempty"`
	CreatedAt      time.Time  `bun:"created_at,notnull,default:now()" json:"created_at"`
}
