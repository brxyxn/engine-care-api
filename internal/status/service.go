package status

import "github.com/rs/zerolog"

type Service interface {
	Status() error
}

type service struct {
	log *zerolog.Logger
}

func NewService(log *zerolog.Logger) Service {
	return &service{log}
}

func (s *service) Status() error {
	// todo: implement real health checks (e.g., database connectivity)
	s.log.Info().Msg("status check successful")
	return nil
}
