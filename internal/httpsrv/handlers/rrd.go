package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"aerospike.com/rrd/internal/models"
)

type RRDGetter interface {
	GetByRange(ctx context.Context, start, end int64) ([]models.Record, error)
}

type RRDSetter interface {
	Create(ctx context.Context, record models.Record) error
}

// RRD contains handlers for processing http requests.
type RRD struct {
	getter RRDGetter
	setter RRDSetter
	logger *slog.Logger
}

// NewRRD returns new handlers struct.
func NewRRD(getter RRDGetter, setter RRDSetter, logger *slog.Logger) *RRD {
	return &RRD{
		getter: getter,
		setter: setter,
		logger: logger,
	}
}

// Create validates request and creates record in database.
func (h *RRD) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		h.logger.Error("failed to create record, wrong method",
			slog.String("method", r.Method),
		)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var record models.Record
	err := json.NewDecoder(r.Body).Decode(&record)
	if err != nil {
		h.logger.Error("failed to create record, failed to decode request", slog.Any("error", err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err = h.setter.Create(r.Context(), record); err != nil {
		h.logger.Error("failed to create record",
			slog.Any("record", record),
			slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Here must be http.StatusCreated, but requirements say http.StatusOK.
	w.WriteHeader(http.StatusOK)
}

// GetByRange validates request and returns records from database by range.
func (h *RRD) GetByRange(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.logger.Error("failed to get records, wrong method",
			slog.String("method", r.Method),
		)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	startString := r.URL.Query().Get("start")
	endString := r.URL.Query().Get("end")

	var (
		start, end int64
		err        error
	)

	if startString != "" {
		start, err = strconv.ParseInt(startString, 10, 64)
		if err != nil {
			h.logger.Error("failed to get records, failed to parse int64",
				slog.String("startString", startString),
			)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	if endString != "" {
		end, err = strconv.ParseInt(endString, 10, 64)
		if err != nil {
			h.logger.Error("failed to get records, failed to parse int64",
				slog.String("endString", endString),
			)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	if start > end || start < 0 || end < 0 {
		h.logger.Error("failed to get records, invalid range",
			slog.Int64("start", start),
			slog.Int64("end", end),
		)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	result, err := h.getter.GetByRange(r.Context(), start, end)
	if err != nil {
		h.logger.Error("failed to get records",
			slog.Int64("start", start),
			slog.Int64("end", end),
			slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(result) == 0 {
		h.logger.Error("failed to get records, not found",
			slog.Int("len(result)", len(result)),
		)
		// I think that it must be http.StatusNotFound.
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err = json.NewEncoder(w).Encode(result); err != nil {
		h.logger.Error("failed to get records, failed to encode", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
