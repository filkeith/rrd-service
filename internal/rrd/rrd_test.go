package rrd

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"aerospike.com/rrd/internal/models"
)

const (
	testMetric    = 3.5
	errorMetric   = 0
	testTimestamp = 1717745157997559
)

var errTest = errors.New("test error")

func testRecord() models.Record {
	return models.Record{
		Timestamp:   testTimestamp,
		MetricValue: testMetric,
	}
}

func errorRecord() models.Record {
	return models.Record{
		Timestamp:   testTimestamp,
		MetricValue: errorMetric,
	}
}

type storageGetterMock struct{}

func (mock storageGetterMock) GetByRange(_ context.Context, min, max int64) ([]models.Record, error) {
	if min < 0 {
		return nil, fmt.Errorf("failed to get by range: %w", errTest)
	}
	return []models.Record{testRecord()}, nil
}

type storageSetterMock struct{}

func (mock storageSetterMock) Set(_ context.Context, record models.Record) error {
	if record.MetricValue != testMetric {
		return fmt.Errorf("failed to set: %w", errTest)
	}
	return nil
}

func newServiceMock() *Service {
	return &Service{
		storageGetter: storageGetterMock{},
		storageSetter: storageSetterMock{},
	}
}

func TestService_Create(t *testing.T) {
	t.Parallel()
	srv := newServiceMock()
	testCases := []struct {
		record models.Record
		err    error
	}{
		{testRecord(), nil},
		{errorRecord(), errTest},
	}

	for i, tt := range testCases {
		err := srv.Create(context.Background(), tt.record)
		require.ErrorIs(t, err, tt.err, fmt.Sprintf("case %d", i))
	}
}

func TestService_GetByRange(t *testing.T) {
	t.Parallel()
	srv := newServiceMock()
	testCases := []struct {
		min, max int64
		records  []models.Record
		err      error
	}{
		{0, 10, []models.Record{testRecord()}, nil},
		{0, 0, []models.Record{testRecord()}, nil},
		{-1, 0, nil, errTest},
	}

	for i, tt := range testCases {
		result, err := srv.GetByRange(context.Background(), tt.min, tt.max)
		require.ErrorIs(t, err, tt.err, fmt.Sprintf("case %d", i))
		require.Equal(t, tt.records, result, fmt.Sprintf("case %d", i))
	}
}
