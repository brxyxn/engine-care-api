package users

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/brxyxn/go-logger"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/uptrace/bun"

	"github.com/brxyxn/engine-care-api/internal/phonenumbers"
)

type s interface {
	Create(user User) error
	List() ([]*User, error)
	ByEmail(email string) (*User, error)
	ByID(id string) (*User, error)
	// todo: Update, Delete
}

type pn interface {
	Add(userID uuid.UUID, countryCode, phoneNumber string, isPrimary bool) error
	List(userID uuid.UUID) ([]*UserPhoneNumber, error)
	// todo: Remove, SetPrimary, Update
}

type Svc struct {
	ctx context.Context
	db  *bun.DB
	log zerolog.Logger
}

type PhoneNumberSvc struct {
	ctx context.Context
	db  *bun.DB
	log *logger.Logger
}

var _ s = (*Svc)(nil)
var _ pn = (*PhoneNumberSvc)(nil)

func Service(ctx context.Context, log zerolog.Logger, db *bun.DB) Svc {
	return Svc{
		ctx: ctx,
		db:  db,
		log: log,
	}
}

func PhoneNumberService(ctx context.Context, log *logger.Logger, db *bun.DB) PhoneNumberSvc {
	return PhoneNumberSvc{
		ctx: ctx,
		db:  db,
		log: log,
	}
}

// Create creates a new user.
func (s *Svc) Create(user User) error {
	_, err := s.db.NewInsert().Model(&user).Returning("id").Exec(s.ctx)
	if err != nil {
		return err
	}
	return nil
}

// List lists all users.
func (s *Svc) List() ([]*User, error) {
	var users []*User
	err := s.db.NewSelect().Model(&users).Scan(s.ctx)
	if err != nil {
		return nil, err
	}
	return users, nil
}

// ByEmail gets a user by email.
func (s *Svc) ByEmail(email string) (*User, error) {
	var user User
	err := s.db.NewSelect().Model(&user).Where("email = ?", email).Scan(s.ctx)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// ByID gets a user by ID, with phone numbers preloaded.
func (s *Svc) ByID(id string) (*User, error) {
	var user User
	err := s.db.NewSelect().
		Model(&user).
		Where("u.id = ?", id).
		Relation("PhoneNumbers").
		Relation("PhoneNumbers.PhoneNumber").
		Scan(s.ctx)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Add adds a phone number to a user.
func (p *PhoneNumberSvc) Add(userID uuid.UUID, countryCode, phoneNumber string, isPrimary bool) error {
	raw := fmt.Sprintf("%s%s", countryCode, phoneNumber)
	e164 := fmt.Sprintf("+%s%s", countryCode, phoneNumber)
	// Normalize/build phone number record
	// NOTE: If you use a lib to format to E.164, set pn.E164 accordingly.
	pn := phonenumbers.PhoneNumber{
		ID:             uuid.New(), // let app set, DB also can default
		RawNumber:      raw,
		E164:           e164, // assume input is already E.164; replace if you normalize
		CountryCode:    countryCode,
		NationalNumber: phoneNumber,
		CreatedAt:      time.Time{}, // DB default now()
	}

	return p.db.RunInTx(p.ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		// Upsert phone number by E.164 and return its ID
		_, err := tx.NewInsert().
			Model(&pn).
			Column("id", "raw_number", "e164", "country_code", "national_number").
			On("CONFLICT (e164) DO UPDATE SET raw_number = EXCLUDED.raw_number").
			Returning("id").
			Exec(ctx)
		if err != nil {
			return err
		}

		// If marking as primary, demote others first
		if isPrimary {
			_, err = tx.NewUpdate().
				Table("user_phone_numbers").
				Set("is_primary = FALSE").
				Where("user_id = ?", userID).
				Where("is_primary = TRUE").
				Exec(ctx)
			if err != nil {
				return err
			}
		}

		upn := UserPhoneNumber{
			UserID:        &userID,
			PhoneNumberID: pn.ID,
			IsPrimary:     isPrimary,
		}

		// Link user to phone number (ignore duplicate links)
		_, err = tx.NewInsert().
			Model(&upn).
			On("CONFLICT (user_id, phone_number_id) DO NOTHING").
			Exec(ctx)
		if err != nil {
			p.log.Err(err).Msg("failed to link user to phone number")
			return err
		}

		return nil
	})
}

// List lists all phone numbers for a user with phone data.
func (p *PhoneNumberSvc) List(userID uuid.UUID) ([]*UserPhoneNumber, error) {
	var upn []*UserPhoneNumber
	err := p.db.NewSelect().
		Model(&upn).
		Where("upn.user_id = ?", userID).
		Relation("PhoneNumber").
		Scan(p.ctx)
	if err != nil {
		p.log.Debug().Err(err).Msg("failed to list phone_numbers")
		return nil, err
	}
	return upn, nil
}
