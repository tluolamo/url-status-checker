package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// URLChecksTotal counts the total number of URL checks performed.
	URLChecksTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "url_checks_total",
			Help: "Total number of URL checks performed",
		},
		[]string{"status"},
	)

	// URLCheckDuration tracks the duration of URL checks.
	URLCheckDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "url_check_duration_seconds",
			Help:    "Time taken to check URLs",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"status_code"},
	)

	// ActiveWorkers tracks the number of active worker goroutines.
	ActiveWorkers = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "url_checker_active_workers",
			Help: "Number of active worker goroutines",
		},
	)

	// RequestsInFlight tracks the number of requests currently being processed.
	RequestsInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "url_checker_requests_in_flight",
			Help: "Number of requests currently being processed",
		},
	)
)
