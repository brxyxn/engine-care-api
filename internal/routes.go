package internal

import (
	"context"

	"github.com/brxyxn/go-logger"
	"github.com/gorilla/mux"

	"github.com/brxyxn/engine-care-api/config"
	"github.com/brxyxn/engine-care-api/internal/middleware"
	"github.com/brxyxn/engine-care-api/internal/status"
	"github.com/brxyxn/engine-care-api/pkg/mwchain"
)

var (
	POST = "POST"
	GET  = "GET"
)

// ConfigRoutes initializes the router and sets up the routes for the API.
func ConfigRoutes(_ context.Context, log *logger.Logger, cfg config.Config) *mux.Router {
	// Initialize the router.
	r := mux.NewRouter().PathPrefix("/v1").Subrouter()

	// Public endpoints.
	statusHandler := status.NewHandler(log.Logger, cfg)
	r.Handle("/status", mwchain.NewChain(middleware.Logger(log.Logger)).Then(statusHandler.Status())).Methods(GET)

	//authHandler := auth.NewHandler(ctx, db, log.Logger, conf)
	//r.Handle("/register", mwchain.NewChain(middleware.Logger(log.Logger)).Then(authHandler.Register())).Methods(POST)
	//r.Handle("/login", mwchain.NewChain(middleware.Logger(log.Logger)).Then(authHandler.Login())).Methods(POST)

	return r
}
