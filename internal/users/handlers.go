package users

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/brxyxn/go-logger"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/uptrace/bun"

	"github.com/brxyxn/engine-care-api/api"
)

type h interface {
	Create() http.HandlerFunc
	List() http.HandlerFunc
	ByEmail() http.HandlerFunc
	ByID() http.HandlerFunc
}

type pnh interface {
	Add() http.HandlerFunc
	List() http.HandlerFunc
	ByNumber() http.HandlerFunc
	ByID() http.HandlerFunc
}

type Hdlr struct {
	ctx context.Context
	db  *bun.DB
	log zerolog.Logger
	svc Svc
}

type PhoneNumberHdlr struct {
	ctx context.Context
	db  *bun.DB
	log *logger.Logger
	svc PhoneNumberSvc
}

var _ h = (*Hdlr)(nil)
var _ pnh = (*PhoneNumberHdlr)(nil)

func Handler(ctx context.Context, log zerolog.Logger, db *bun.DB) Hdlr {
	svc := Service(ctx, log, db)
	return Hdlr{ctx, db, log, svc}
}

type CreateUser struct {
	StackUserID string  `json:"stack_user_id"`
	Email       string  `json:"email"`
	FirstName   string  `json:"first_name"`
	LastName    *string `json:"last_name,omitempty"`
	AvatarURL   *string `json:"avatar_url,omitempty"`
}

func (h *Hdlr) Create() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := CreateUser{}
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			return
		}

		displayName := fmt.Sprintf("%s %s", data.FirstName, *data.LastName)
		user := User{
			ID:          uuid.New(), // todo: validate if StackUserID is defined and unique otherwise create a new one
			StackUserID: data.StackUserID,
			Email:       data.Email,
			DisplayName: displayName,
			AvatarURL:   data.AvatarURL,
		}

		err = h.svc.Create(user)
		if err != nil {
			// todo: validate error type for conflict vs internal server error
			h.log.Debug().Err(err).Msg("error creating user")
			api.Error(w, http.StatusInternalServerError, api.ErrorResponse{Stack: err.Error()})
			return
		}

		api.Success[User](w, http.StatusCreated, user)
	}
}

func (h *Hdlr) List() http.HandlerFunc {
	//TODO implement me
	panic("implement me")
}

func (h *Hdlr) ByEmail() http.HandlerFunc {
	//TODO implement me
	panic("implement me")
}

func (h *Hdlr) ByID() http.HandlerFunc {
	//TODO implement me
	panic("implement me")
}

func (p PhoneNumberHdlr) Add() http.HandlerFunc {
	//TODO implement me
	panic("implement me")
}

func (p PhoneNumberHdlr) List() http.HandlerFunc {
	//TODO implement me
	panic("implement me")
}

func (p PhoneNumberHdlr) ByNumber() http.HandlerFunc {
	//TODO implement me
	panic("implement me")
}

func (p PhoneNumberHdlr) ByID() http.HandlerFunc {
	//TODO implement me
	panic("implement me")
}
