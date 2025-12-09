package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tluolamo/url-status-checker/internal/checker"
	"github.com/tluolamo/url-status-checker/internal/config"
	"github.com/tluolamo/url-status-checker/internal/metrics"
	"github.com/tluolamo/url-status-checker/internal/models"
)

const (
	contentTypeHeader = "Content-Type"
	contentTypeJSON   = "application/json"
	contentTypeHTML   = "text/html; charset=utf-8"
)

// Server represents the HTTP server.
type Server struct {
	router    *chi.Mux
	config    *config.Config
	checker   *checker.Checker
	startTime time.Time
	logger    *slog.Logger
}

// NewServer creates a new HTTP server.
func NewServer(cfg *config.Config, logger *slog.Logger) *Server {
	s := &Server{
		router:    chi.NewRouter(),
		config:    cfg,
		checker:   checker.New(cfg.DefaultTimeout, cfg.MaxWorkers),
		startTime: time.Now(),
		logger:    logger,
	}

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.Timeout(60 * time.Second))

	s.router.Route("/api/v1", func(r chi.Router) {
		r.Post("/check", s.handleCheckURLs)
		r.Get("/health", s.handleHealth)
	})

	s.router.Handle("/metrics", promhttp.Handler())
	s.router.Get("/", s.handleDashboard)
}

func (s *Server) handleCheckURLs(w http.ResponseWriter, r *http.Request) {
	metrics.RequestsInFlight.Inc()
	defer metrics.RequestsInFlight.Dec()

	var req models.CheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.Error("failed to decode request", "error", err)
		http.Error(w, fmt.Sprintf("invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	if len(req.URLs) == 0 {
		http.Error(w, "urls field is required and must not be empty", http.StatusBadRequest)
		return
	}

	if len(req.URLs) > 1000 {
		http.Error(w, "maximum 1000 URLs allowed per request", http.StatusBadRequest)
		return
	}

	timeout := s.config.DefaultTimeout
	if req.Timeout > 0 {
		timeout = req.Timeout
	}

	maxWorkers := s.config.MaxWorkers
	if req.MaxWorkers > 0 {
		maxWorkers = req.MaxWorkers
	}

	urlChecker := checker.New(timeout, maxWorkers)

	start := time.Now()
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	results := urlChecker.CheckURLs(ctx, req.URLs)
	totalTime := time.Since(start)

	for _, result := range results {
		status := "success"
		if result.Error != "" {
			status = "failure"
		}
		metrics.URLChecksTotal.WithLabelValues(status).Inc()
		metrics.URLCheckDuration.WithLabelValues(fmt.Sprintf("%d", result.StatusCode)).Observe(float64(result.ResponseTimeMs) / 1000.0)
	}

	availableCount := 0
	for _, result := range results {
		if result.Available {
			availableCount++
		}
	}

	response := models.CheckResponse{
		Results:        results,
		TotalChecked:   len(results),
		TotalAvailable: availableCount,
		TotalTimeMs:    totalTime.Milliseconds(),
	}

	w.Header().Set(contentTypeHeader, contentTypeJSON)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.logger.Error("failed to encode response", "error", err)
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	uptime := time.Since(s.startTime)

	response := models.HealthResponse{
		Status:  "healthy",
		Version: s.config.Version,
		Uptime:  uptime.String(),
		Time:    time.Now(),
	}

	w.Header().Set(contentTypeHeader, contentTypeJSON)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.logger.Error("failed to encode health response", "error", err)
	}
}

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>URL Status Checker</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            padding: 20px;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background: white;
            border-radius: 10px;
            box-shadow: 0 20px 60px rgba(0,0,0,0.3);
            padding: 40px;
        }
        h1 {
            color: #333;
            margin-bottom: 10px;
            font-size: 32px;
        }
        .subtitle {
            color: #666;
            margin-bottom: 30px;
        }
        .input-section {
            margin-bottom: 30px;
        }
        textarea {
            width: 100%;
            min-height: 150px;
            padding: 15px;
            border: 2px solid #e0e0e0;
            border-radius: 5px;
            font-size: 14px;
            font-family: 'Courier New', monospace;
            resize: vertical;
        }
        textarea:focus {
            outline: none;
            border-color: #667eea;
        }
        .controls {
            display: flex;
            gap: 10px;
            margin-top: 10px;
        }
        button {
            padding: 12px 24px;
            background: #667eea;
            color: white;
            border: none;
            border-radius: 5px;
            font-size: 16px;
            cursor: pointer;
            transition: background 0.3s;
        }
        button:hover { background: #5568d3; }
        button:disabled {
            background: #ccc;
            cursor: not-allowed;
        }
        .results { margin-top: 30px; }
        .result-item {
            background: #f8f9fa;
            border-left: 4px solid #28a745;
            padding: 15px;
            margin-bottom: 10px;
            border-radius: 4px;
        }
        .result-item.unavailable { border-left-color: #dc3545; }
        .url { font-weight: 600; color: #333; margin-bottom: 5px; }
        .details { font-size: 14px; color: #666; }
        .status-badge {
            display: inline-block;
            padding: 3px 8px;
            border-radius: 3px;
            font-size: 12px;
            font-weight: 600;
            margin-right: 10px;
        }
        .status-success { background: #d4edda; color: #155724; }
        .status-error { background: #f8d7da; color: #721c24; }
        .summary {
            background: #e7f3ff;
            padding: 15px;
            border-radius: 5px;
            margin-bottom: 20px;
        }
        .spinner {
            border: 3px solid #f3f3f3;
            border-top: 3px solid #667eea;
            border-radius: 50%;
            width: 40px;
            height: 40px;
            animation: spin 1s linear infinite;
            margin: 20px auto;
        }
        @keyframes spin { 0% { transform: rotate(0deg); } 100% { transform: rotate(360deg); } }
    </style>
</head>
<body>
    <div class="container">
        <h1>🚀 URL Status Checker</h1>
        <p class="subtitle">Check multiple URLs concurrently with Go's powerful goroutines</p>
        <div class="input-section">
            <textarea id="urlInput" placeholder="Enter URLs (one per line):
https://google.com
https://github.com
https://example.com"></textarea>
            <div class="controls">
                <button onclick="checkURLs()" id="checkBtn">Check URLs</button>
                <button onclick="clearResults()" id="clearBtn">Clear</button>
            </div>
        </div>
        <div id="results" class="results"></div>
    </div>

    <script>
        async function checkURLs() {
            const textarea = document.getElementById('urlInput');
            const urls = textarea.value.split('\n').map(u => u.trim()).filter(Boolean);

            if (urls.length === 0) {
                alert('Please enter at least one URL');
                return;
            }

            const btn = document.getElementById('checkBtn');
            btn.disabled = true;
            btn.textContent = 'Checking...';

            const resultsDiv = document.getElementById('results');
            resultsDiv.innerHTML = '<div class="spinner"></div>';

            try {
                const response = await fetch('/api/v1/check', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ urls })
                });

                const data = await response.json();
                displayResults(data);
            } catch (error) {
                resultsDiv.innerHTML = '<div class="result-item unavailable">' +
                    '<div class="url">Error</div>' +
                    '<div class="details">' + error.message + '</div>' +
                    '</div>';
            } finally {
                btn.disabled = false;
                btn.textContent = 'Check URLs';
            }
        }

        function displayResults(data) {
            const resultsDiv = document.getElementById('results');
            let html = '<div class="summary">' +
                '<strong>Summary:</strong> ' +
                'Checked ' + data.total_checked + ' URLs in ' + data.total_time_ms + 'ms | ' +
                'Available: ' + data.total_available + ' | ' +
                'Unavailable: ' + (data.total_checked - data.total_available) +
                '</div>';

            data.results.forEach(result => {
                const statusClass = result.available ? 'status-success' : 'status-error';
                const itemClass = result.available ? '' : 'unavailable';
                const statusText = result.available ? '✓ Available' : '✗ Unavailable';

                html += '<div class="result-item ' + itemClass + '">' +
                    '<div class="url">' + escapeHtml(result.url) + '</div>' +
                    '<div class="details">' +
                        '<span class="status-badge ' + statusClass + '">' + statusText + '</span>' +
                        'Status: ' + (result.status_code || 'N/A') + ' | ' +
                        'Response Time: ' + result.response_time_ms + 'ms' +
                        (result.error ? '<br><strong>Error:</strong> ' + escapeHtml(result.error) : '') +
                    '</div>' +
                '</div>';
            });

            resultsDiv.innerHTML = html;
        }

        function clearResults() {
            document.getElementById('urlInput').value = '';
            document.getElementById('results').innerHTML = '';
        }

        function escapeHtml(text) {
            const div = document.createElement('div');
            div.textContent = text;
            return div.innerHTML;
        }

        document.getElementById('urlInput').addEventListener('keydown', (e) => {
            if (e.ctrlKey && e.key === 'Enter') {
                checkURLs();
            }
        });
    </script>
</body>
</html>`

	w.Header().Set(contentTypeHeader, contentTypeHTML)
	if _, err := io.WriteString(w, html); err != nil {
		s.logger.Error("failed to write dashboard", "error", err)
	}
}

// Start runs the HTTP server.
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.config.Port)
	s.logger.Info("starting server", "address", addr)
	server := &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	return server.ListenAndServe()
}
