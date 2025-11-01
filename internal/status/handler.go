package status

import (
	"net/http"

	"github.com/brxyxn/go-logger"

	"github.com/brxyxn/engine-care-api/config"
	"github.com/brxyxn/engine-care-api/internal/api"
)

type Handler interface {
	Status() http.HandlerFunc
}

type handler struct {
	svc Service
	log *logger.Logger
	cfg config.Config
}

func NewHandler(log *logger.Logger, cfg config.Config) Handler {
	svc := NewService(log)
	return &handler{svc, log, cfg}
}

func (h *handler) Status() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := h.svc.Status()
		if err != nil {
			h.log.Error().Err(err).Msg("status check failed")
			http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
			return
		}

		api.Success(w, http.StatusOK, Response{
			Message: "Service is healthy",
			Status:  http.StatusOK,
		})
	}
}
