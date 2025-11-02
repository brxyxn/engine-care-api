package status

import (
	"github.com/brxyxn/go-logger"
	"github.com/gorilla/mux"
	"github.com/uptrace/bun"

	"github.com/brxyxn/engine-care-api/api"
	"github.com/brxyxn/engine-care-api/config"
	"github.com/brxyxn/engine-care-api/internal/middleware"
	"github.com/brxyxn/engine-care-api/pkg/mwchain"
)

func Routes(v1 *mux.Router, log *logger.Logger, cfg config.Config, db *bun.DB) {
	stsLog := log.With().Str("route", "status").Logger()
	statusHandler := NewHandler(stsLog, cfg, db)
	v1.Handle("/status", mwchain.NewChain(middleware.Logger(stsLog)).Then(statusHandler.Status())).Methods(api.GET)
}
