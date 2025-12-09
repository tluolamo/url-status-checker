package models

import "time"

// CheckRequest represents a request to check multiple URLs.
type CheckRequest struct {
	URLs       []string      `json:"urls"`
	Timeout    time.Duration `json:"timeout,omitempty"`
	MaxWorkers int           `json:"max_workers,omitempty"`
}

// CheckResult represents the result of checking a single URL.
type CheckResult struct {
	CheckedAt      time.Time `json:"checked_at"`
	URL            string    `json:"url"`
	Error          string    `json:"error,omitempty"`
	ResponseTimeMs int64     `json:"response_time_ms"`
	StatusCode     int       `json:"status_code"`
	Available      bool      `json:"available"`
}

// CheckResponse represents the response containing all check results.
type CheckResponse struct {
	Results        []CheckResult `json:"results"`
	TotalChecked   int           `json:"total_checked"`
	TotalAvailable int           `json:"total_available"`
	TotalTimeMs    int64         `json:"total_time_ms"`
}

// HealthResponse represents a health check response.
type HealthResponse struct {
	Time    time.Time `json:"time"`
	Status  string    `json:"status"`
	Version string    `json:"version"`
	Uptime  string    `json:"uptime"`
}
