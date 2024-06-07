package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/steinfletcher/apitest"

	"aerospike.com/rrd/internal/models"
)

const (
	testMetric  = 3.5
	errorMetric = 0
)

var errTest = errors.New("test error")

func testRecord() models.Record {
	return models.Record{
		Timestamp:   time.Now().UnixMicro(),
		MetricValue: testMetric,
	}
}

func errorRecord() models.Record {
	return models.Record{
		Timestamp:   time.Now().UnixMicro(),
		MetricValue: errorMetric,
	}
}

func testBody() string {
	body, _ := json.Marshal(testRecord())
	return string(body)
}

func errorBody() string {
	body, _ := json.Marshal(errorRecord())
	return string(body)
}

type getterMock struct{}

func (mock getterMock) GetByRange(_ context.Context, min, max int64) ([]models.Record, error) {
	if min < 0 || max < 0 {
		return nil, fmt.Errorf("failed to get by range: %w", errTest)
	}
	return []models.Record{testRecord()}, nil
}

type setterMock struct{}

func (mock setterMock) Create(_ context.Context, record models.Record) error {
	if record.MetricValue != testMetric {
		return fmt.Errorf("failed to set: %w", errTest)
	}
	return nil
}

func newRRDMock() *RRD {
	return &RRD{
		getter: getterMock{},
		setter: setterMock{},
		logger: slog.New(slog.NewJSONHandler(os.Stdout, nil)),
	}
}

func TestRRD_Create(t *testing.T) {
	t.Parallel()
	h := newRRDMock()
	router := mux.NewRouter()
	router.HandleFunc(
		"/metrics",
		h.Create,
	).Methods(http.MethodPut)

	testCases := []struct {
		method     string
		statusCode int
		body       string
	}{
		{http.MethodPut, http.StatusOK, testBody()},
		{http.MethodPut, http.StatusBadRequest, ""},
		{http.MethodPut, http.StatusInternalServerError, errorBody()},
		{http.MethodPost, http.StatusMethodNotAllowed, testBody()},
		{http.MethodConnect, http.StatusMethodNotAllowed, testBody()},
		{http.MethodDelete, http.StatusMethodNotAllowed, testBody()},
		{http.MethodPatch, http.StatusMethodNotAllowed, testBody()},
		{http.MethodGet, http.StatusMethodNotAllowed, testBody()},
		{http.MethodTrace, http.StatusMethodNotAllowed, testBody()},
	}

	for _, tt := range testCases {
		apitest.New().
			Handler(router).
			Method(tt.method).
			URL("/metrics").
			Body(tt.body).
			Expect(t).
			Status(tt.statusCode).
			End()
	}
}

func TestRRD_GetByRange(t *testing.T) {
	t.Parallel()
	h := newRRDMock()
	router := mux.NewRouter()
	router.HandleFunc(
		"/metrics",
		h.GetByRange,
	).Methods(http.MethodGet)

	testCases := []struct {
		method     string
		statusCode int
		start      string
		end        string
	}{
		{http.MethodGet, http.StatusOK, "0", "10"},
		{http.MethodGet, http.StatusBadRequest, "a", "b"},
		{http.MethodGet, http.StatusBadRequest, "-1", "0"},
		{http.MethodGet, http.StatusBadRequest, "5", "1"},
		{http.MethodGet, http.StatusBadRequest, "0", "-1"},
		{http.MethodPost, http.StatusMethodNotAllowed, "0", "10"},
		{http.MethodConnect, http.StatusMethodNotAllowed, "0", "10"},
		{http.MethodDelete, http.StatusMethodNotAllowed, "0", "10"},
		{http.MethodPatch, http.StatusMethodNotAllowed, "0", "10"},
		{http.MethodPut, http.StatusMethodNotAllowed, "0", "10"},
		{http.MethodTrace, http.StatusMethodNotAllowed, "0", "10"},
	}

	for _, tt := range testCases {
		apitest.New().
			Handler(router).
			Method(tt.method).
			URL("/metrics").
			QueryParams(map[string]string{"start": tt.start, "end": tt.end}).
			Expect(t).
			Status(tt.statusCode).
			End()
	}
}
