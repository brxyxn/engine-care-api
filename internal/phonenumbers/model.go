package phonenumbers

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type PhoneNumber struct {
	bun.BaseModel `bun:"table:phone_numbers,alias:pn"`

	ID             uuid.UUID `bun:"id,pk,default:gen_random_uuid()" json:"id"`
	RawNumber      string    `bun:"raw_number,notnull" json:"raw_number"`
	E164           string    `bun:"e164,notnull" json:"e164"`
	CountryCode    string    `bun:"country_code,notnull" json:"country_code"`
	NationalNumber string    `bun:"national_number,notnull" json:"national_number"`
	CreatedAt      time.Time `bun:"created_at,notnull,default:now()" json:"created_at"`
}
