package status

import (
	"github.com/rs/zerolog"
	"github.com/uptrace/bun"
)

type Service interface {
	Status() error
}

type service struct {
	log zerolog.Logger
	db  *bun.DB
}

func NewService(log zerolog.Logger, db *bun.DB) Service {
	return &service{log, db}
}

func (s *service) Status() error {
	err := s.db.Ping()
	if err != nil {
		s.log.Error().Err(err).Msg("database ping failed")
		return err
	}

	s.log.Info().Msg("status check successful")
	return nil
}
