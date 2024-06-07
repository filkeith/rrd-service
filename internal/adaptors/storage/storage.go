package storage

import (
	"context"
	"fmt"
	"log/slog"
	"sync/atomic"

	"github.com/aerospike/aerospike-client-go/v7"

	"aerospike.com/rrd/internal/models"
)

const (
	setNameMetrics     = "metrics"
	setNameCounter     = "counter"
	binNameTimestamp   = "timestamp"
	binNameMetricValue = "metric_value"
	binNameCounter     = "counter"
	udfFileName        = "find_oldest.lua"
)

// Storage contains database logic.
type Storage struct {
	namespace  string
	maxRecords uint64

	counter atomic.Uint64
	client  *aerospike.Client

	logger *slog.Logger
}

// NewStorage returns new storage for processing time series data.
func NewStorage(host string, port int, namespace string, maxRecords uint64, udfPath string, logger *slog.Logger,
) (*Storage, error) {
	aerospike.SetLuaPath(udfPath)
	client, err := aerospike.NewClient(host, port)
	// Why it returns custom error type!?
	if err != nil {
		return nil, fmt.Errorf("failed to initialize aerospike client: %w", err)
	}

	// create index.
	indexTask, err := client.CreateIndex(nil,
		namespace,
		setNameMetrics,
		"idx_timestamp",
		binNameTimestamp,
		aerospike.NUMERIC)
	if err != nil {
		return nil, fmt.Errorf("failed to create index: %w", err)
	}
	<-indexTask.OnComplete()

	task, err := client.RegisterUDFFromFile(nil, fmt.Sprintf("%s%s", udfPath, udfFileName), udfFileName, aerospike.LUA)
	if err != nil {
		return nil, fmt.Errorf("failed to register udf: %w", err)
	}
	// Wait for the registration to complete
	<-task.OnComplete()

	storage := &Storage{
		namespace:  namespace,
		maxRecords: maxRecords,
		client:     client,
		logger:     logger,
	}

	// Initialize counter.
	counter, errCounter := storage.GetCounter(context.Background())
	if errCounter == nil {
		storage.counter.Store(uint64(counter))
	}
	logger.Debug("initialized counter", slog.Int64(binNameCounter, counter))

	return storage, nil
}

// Close closes connection to aerospike instance.
func (s *Storage) Close() {
	s.client.Close()
}

// Set saves record to the database.
func (s *Storage) Set(ctx context.Context, record models.Record) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context error: %w", err)
	}
	if s.counter.Load() >= s.maxRecords {
		if err := s.evict(ctx); err != nil {
			return fmt.Errorf("failed to evict records: %w", err)
		}
	}

	key, err := aerospike.NewKey(s.namespace, setNameMetrics, record.Timestamp)
	if err != nil {
		return fmt.Errorf("failed to create aerospike key: %w", err)
	}

	bin := aerospike.BinMap{
		binNameTimestamp:   record.Timestamp,
		binNameMetricValue: record.MetricValue,
	}

	writePolicy := aerospike.NewWritePolicy(0, aerospike.TTLDontExpire)

	err = s.client.Put(writePolicy, key, bin)
	if err != nil {
		return fmt.Errorf("failed to put bins: %w", err)
	}

	// Increase counter only if we have less than `maxRecords`
	if s.counter.Load() < s.maxRecords {
		s.counter.Add(1)
		s.SetCounter(ctx, int64(s.counter.Load()))
	}

	return nil
}

// GetByRange returns records from a database by range.
func (s *Storage) GetByRange(ctx context.Context, min, max int64) ([]models.Record, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context error: %w", err)
	}

	stmt := aerospike.NewStatement(s.namespace, setNameMetrics)
	if err := stmt.SetFilter(aerospike.NewRangeFilter(binNameTimestamp, min, max)); err != nil {
		return nil, fmt.Errorf("failed to set statement filter: %w", err)
	}

	recordset, err := s.client.Query(nil, stmt)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	defer recordset.Close()

	results := make([]models.Record, 0)
	for res := range recordset.Results() {
		if res.Err != nil {
			return nil, fmt.Errorf("failed to itterate over result: %w", res.Err)
		}
		timestamp, ok := res.Record.Bins[binNameTimestamp].(int)
		if !ok {
			return nil, fmt.Errorf("failed to cast timestamp to int64")
		}
		metricValue, ok := res.Record.Bins[binNameMetricValue]
		if !ok {
			return nil, fmt.Errorf("failed to cast metric_value to float64")
		}
		one := models.Record{
			Timestamp:   int64(timestamp),
			MetricValue: metricValue,
		}
		results = append(results, one)
	}

	return results, nil
}

// evict finds oldest record in a database and delete it.
func (s *Storage) evict(ctx context.Context) error {
	oldestKey, errKey := s.FindOldestKey(ctx)
	if errKey != nil {
		return fmt.Errorf("failed to find oldest key: %w", errKey)
	}
	if oldestKey != nil {
		_, err := s.client.Delete(nil, oldestKey)
		if err != nil {
			return fmt.Errorf("failed to delete oldest key: %w", err)
		}
	}

	return nil
}

// SetCounter saves counter do db. As this function will be called in goroutine, we don't return errors here.
func (s *Storage) SetCounter(ctx context.Context, val int64) {
	if err := ctx.Err(); err != nil {
		s.logger.Error("context error", slog.Any("error", err))
	}

	key, err := aerospike.NewKey(s.namespace, setNameCounter, 0)
	if err != nil {
		s.logger.Error("failed to create aerospike key", slog.Any("error", err))
	}

	bin := aerospike.BinMap{
		binNameCounter: val,
	}

	err = s.client.Put(nil, key, bin)
	if err != nil {
		s.logger.Error("failed to put bins", slog.Any("error", err))
	}
}

// GetCounter retrieves counter from a database for an initial load.
func (s *Storage) GetCounter(ctx context.Context) (int64, error) {
	if err := ctx.Err(); err != nil {
		return 0, fmt.Errorf("context error: %w", err)
	}

	key, err := aerospike.NewKey(s.namespace, setNameCounter, 0)
	if err != nil {
		return 0, fmt.Errorf("failed to create aerospike key: %w", err)
	}

	record, err := s.client.Get(nil, key)
	if err != nil {
		return 0, fmt.Errorf("failed to get counter: %w", err)
	}

	counter, ok := record.Bins[binNameCounter].(int)
	if !ok {
		return 0, fmt.Errorf("failed to cast counter to int")
	}

	return int64(counter), nil
}

// FindOldestKey returns key of the oldest record for eviction.
func (s *Storage) FindOldestKey(ctx context.Context) (*aerospike.Key, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context error: %w", err)
	}

	stmt := aerospike.NewStatement(s.namespace, setNameMetrics)
	recordset, err := s.client.QueryAggregate(nil, stmt, "find_oldest", "find_oldest")
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer recordset.Close()

	for res := range recordset.Results() {
		if res.Err != nil {
			return nil, res.Err
		}
		var key *aerospike.Key
		if record, ok := res.Record.Bins["SUCCESS"].(map[interface{}]interface{}); ok {
			timestamp := int64(record[binNameTimestamp].(int))
			key, err = aerospike.NewKey(s.namespace, setNameMetrics, timestamp)
			if err != nil {
				return nil, fmt.Errorf("failed to create aerospike key: %w", err)
			}
			return key, nil
		}
	}

	return nil, nil
}
