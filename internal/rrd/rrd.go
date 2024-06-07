package rrd

import (
	"context"
	"fmt"
	"time"

	"aerospike.com/rrd/internal/models"
)

type storageGetter interface {
	GetByRange(ctx context.Context, min, max int64) ([]models.Record, error)
}

type storageSetter interface {
	Set(ctx context.Context, record models.Record) error
}

type Service struct {
	storageGetter storageGetter
	storageSetter storageSetter
}

func NewService(storageGetter storageGetter, storageSetter storageSetter) *Service {
	return &Service{
		storageGetter: storageGetter,
		storageSetter: storageSetter,
	}
}

func (s *Service) Create(ctx context.Context, record models.Record) error {
	if err := s.storageSetter.Set(ctx, record); err != nil {
		return fmt.Errorf("failed to create record: %w", err)
	}
	return nil
}

func (s *Service) GetByRange(ctx context.Context, start, end int64) ([]models.Record, error) {
	// if start = 0 and end = 0 we select all records.
	if start == 0 && end == 0 {
		end = time.Now().UnixMicro()
	}
	records, err := s.storageGetter.GetByRange(ctx, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to get records: %w", err)
	}
	return records, nil
}
