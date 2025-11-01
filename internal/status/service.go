package status

import (
	"github.com/brxyxn/go-logger"
)

type Service interface {
	Status() error
}

type service struct {
	log *logger.Logger
}

func NewService(log *logger.Logger) Service {
	return &service{log}
}

func (s *service) Status() error {
	// todo: implement real health checks (e.g., database connectivity)
	s.log.Info().Msg("status check successful")
	return nil
}
