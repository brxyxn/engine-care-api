package status

import (
	"net/http"

	"github.com/brxyxn/go-logger"
	"github.com/uptrace/bun"

	"github.com/brxyxn/engine-care-api/api"
	"github.com/brxyxn/engine-care-api/config"
)

type Handler interface {
	Status() http.HandlerFunc
}

type handler struct {
	svc Service
	log *logger.Logger
	cfg config.Config
	db  *bun.DB
}

func NewHandler(log *logger.Logger, cfg config.Config, db *bun.DB) Handler {
	svc := NewService(log, db)
	return &handler{svc, log, cfg, db}
}

func (h *handler) Status() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := h.svc.Status()
		if err != nil {
			h.log.Error().Err(err).Msg("status check failed")
			http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
			return
		}

		api.Success[Response](w, http.StatusOK, Response{
			Message:   "Service is healthy",
			SvrStatus: http.StatusText(http.StatusOK),
			DBStatus:  http.StatusText(http.StatusOK),
		})
	}
}
