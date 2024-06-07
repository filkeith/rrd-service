package models

// Record represents a record with timestamp and metric value.
type Record struct {
	// It would be great to add Source fields, so we can save metrics from different services/sources.
	// But in requirements, there is no such field.
	// Source string `json:"string"`
	Timestamp   int64 `json:"timestamp"`
	MetricValue any   `json:"metric_value"`
}
