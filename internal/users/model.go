package users

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type User struct {
	bun.BaseModel `bun:"table:users,alias:u"`

	ID          uuid.UUID `bun:"id,pk,default:gen_random_uuid()" json:"id"`
	StackUserID string    `bun:"stack_user_id,unique,notnull" json:"stack_user_id"`
	Email       string    `bun:"email,notnull" json:"email"`
	DisplayName *string   `bun:"display_name" json:"display_name,omitempty"`
	AvatarURL   *string   `bun:"avatar_url" json:"avatar_url,omitempty"`
	IsActive    bool      `bun:"is_active,notnull,default:true" json:"is_active"`
	CreatedAt   time.Time `bun:"created_at,notnull,default:now()" json:"created_at"`
	UpdatedAt   time.Time `bun:"updated_at,notnull,default:now()" json:"updated_at"`
}

type UserPhoneNumber struct {
	bun.BaseModel `bun:"table:user_phone_numbers,alias:upn"`

	ID            uuid.UUID  `bun:"id,pk,default:gen_random_uuid()" json:"id"`
	UserID        *uuid.UUID `bun:"user_id" json:"user_id,omitempty"`
	PhoneNumberID uuid.UUID  `bun:"phone_number_id,notnull" json:"phone_number_id"`
	IsPrimary     bool       `bun:"is_primary,notnull,default:false" json:"is_primary"`
	CreatedAt     time.Time  `bun:"created_at,notnull,default:now()" json:"created_at"`
}
