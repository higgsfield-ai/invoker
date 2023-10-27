package collector

import (
	"context"

	"github.com/pkg/errors"
)

type Service struct {
	db *DB
}

func NewService(db *DB) *Service {
	return &Service{db: db}
}

func (s *Service) NewLogsToCollect(ctx context.Context, ltcs []LogToCollect) error {
	for _, ltc := range ltcs {
		if err := s.db.InsertLog(ctx, ltc); err != nil {
			return errors.WithMessagef(err, "failed to insert log")
		}
	}

	return nil
}

func (s *Service) NewStopOrErr(ctx context.Context, soe StopOrErr) error {
	if err := s.db.InsertStop(ctx, soe); err != nil {
		return errors.WithMessagef(err, "failed to insert stop or err")
	}

	return nil
}
