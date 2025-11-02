package status

import (
	"github.com/brxyxn/go-logger"
	"github.com/uptrace/bun"
)

type Service interface {
	Status() error
}

type service struct {
	log *logger.Logger
	db  *bun.DB
}

func NewService(log *logger.Logger, db *bun.DB) Service {
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
