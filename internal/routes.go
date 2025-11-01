package internal

import (
	"context"

	"github.com/brxyxn/go-logger"
	"github.com/gorilla/mux"
	"github.com/uptrace/bun"

	"github.com/brxyxn/engine-care-api/config"
	"github.com/brxyxn/engine-care-api/internal/middleware"
	"github.com/brxyxn/engine-care-api/internal/status"
	"github.com/brxyxn/engine-care-api/pkg/mwchain"
)

var (
	POST = "POST"
	GET  = "GET"
)

type Routes struct {
	rtr *mux.Router
	ctx context.Context
	log *logger.Logger
	cfg config.Config
	db  *bun.DB
}

func NewRoutes(ctx context.Context, cfg config.Config, log *logger.Logger, db *bun.DB) *Routes {
	return &Routes{
		rtr: mux.NewRouter(),
		ctx: ctx,
		log: log,
		cfg: cfg,
		db:  db,
	}
}

// ConfigRoutes initializes the router and sets up the routes for the API.
func (r Routes) ConfigRoutes() *mux.Router {
	// Initialize the router.
	r.rtr = r.rtr.PathPrefix("/v1").Subrouter()

	// Public endpoints.
	statusHandler := status.NewHandler(r.log, r.cfg)
	r.rtr.Handle("/status", mwchain.NewChain(middleware.Logger(r.log)).Then(statusHandler.Status())).Methods(GET)

	//authHandler := auth.NewHandler(ctx, db, log.Logger, conf)
	//r.Handle("/register", mwchain.NewChain(middleware.Logger(log.Logger)).Then(authHandler.Register())).Methods(POST)
	//r.Handle("/login", mwchain.NewChain(middleware.Logger(log.Logger)).Then(authHandler.Login())).Methods(POST)

	return r.rtr
}
