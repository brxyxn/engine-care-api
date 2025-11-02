package users

import (
	"context"

	"github.com/brxyxn/go-logger"
	"github.com/gorilla/mux"
	"github.com/uptrace/bun"

	"github.com/brxyxn/engine-care-api/api"
	"github.com/brxyxn/engine-care-api/internal/middleware"
	"github.com/brxyxn/engine-care-api/pkg/mwchain"
)

func Routes(ctx context.Context, v1 *mux.Router, log *logger.Logger, db *bun.DB) {
	u := v1.PathPrefix("/users").Subrouter()
	usrLog := log.With().Str("route", "users").Logger()
	usrHandler := Handler(ctx, usrLog, db)
	u.Handle("", mwchain.NewChain(middleware.Logger(usrLog)).Then(api.Placeholder())).Methods(api.GET)
	u.Handle("/by-id", mwchain.NewChain(middleware.Logger(usrLog)).Then(api.Placeholder())).Methods(api.GET)
	u.Handle("/by-email", mwchain.NewChain(middleware.Logger(usrLog)).Then(api.Placeholder())).Methods(api.GET)
	u.Handle("/create", mwchain.NewChain(middleware.Logger(usrLog)).Then(usrHandler.Create())).Methods(api.POST)
}
