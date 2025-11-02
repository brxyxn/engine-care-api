package internal

import (
	"context"

	"github.com/brxyxn/go-logger"
	"github.com/gorilla/mux"
	"github.com/uptrace/bun"

	"github.com/brxyxn/engine-care-api/config"
	"github.com/brxyxn/engine-care-api/internal/status"
	"github.com/brxyxn/engine-care-api/internal/users"
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
	ctx := r.ctx
	cfg := r.cfg
	log := r.log
	db := r.db

	// Initialize the router.
	v1 := r.rtr.PathPrefix("/v1").Subrouter()

	// Public endpoints.
	status.Routes(v1, log, cfg, db)

	// Private endpoints
	users.Routes(ctx, v1, log, db)

	return r.rtr
}
