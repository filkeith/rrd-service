package storage

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"aerospike.com/rrd/internal/models"
)

const (
	testHost       = "localhost"
	testPort       = 3000
	testNamespace  = "test"
	testMaxRecords = 5
	udfPath        = "../../../udf/"
	testCounter    = int64(10)
)

func testRecord(ts int64) models.Record {
	return models.Record{
		Timestamp:   ts,
		MetricValue: 3.5,
	}
}

func TestStorage_Set(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	storage, err := NewStorage(testHost, testPort, testNamespace, testMaxRecords, udfPath, logger)
	require.NoError(t, err)

	err = storage.Set(context.Background(), testRecord(1))
	require.NoError(t, err)
}

func TestStorage_GetByRange(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	storage, err := NewStorage(testHost, testPort, testNamespace, testMaxRecords, udfPath, logger)
	require.NoError(t, err)

	for i := 0; i < 10; i++ {
		err = storage.Set(context.Background(), testRecord(time.Now().UnixMicro()))
		require.NoError(t, err)
	}
	result, err := storage.GetByRange(context.Background(), 0, time.Now().UnixMicro())
	require.NoError(t, err)
	require.Equal(t, testMaxRecords, len(result))
}

func TestStorage_SetCounter(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	storage, err := NewStorage(testHost, testPort, testNamespace, testMaxRecords, udfPath, logger)
	require.NoError(t, err)

	storage.SetCounter(context.Background(), testCounter)
	val, err := storage.GetCounter(context.Background())
	require.NoError(t, err)
	require.Equal(t, testCounter, val)
}
