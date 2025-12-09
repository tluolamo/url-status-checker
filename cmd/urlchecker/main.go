package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/tluolamo/url-status-checker/internal/api"
	"github.com/tluolamo/url-status-checker/internal/config"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Setup logger
	logLevel := slog.LevelInfo
	switch cfg.LogLevel {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))

	slog.SetDefault(logger)

	// Print banner
	fmt.Println("╔═══════════════════════════════════════════════╗")
	fmt.Println("║   URL Status Checker v" + cfg.Version + "            ║")
	fmt.Println("║   High-Concurrency URL Testing with Go       ║")
	fmt.Println("╚═══════════════════════════════════════════════╝")
	fmt.Println()

	// Create and start server
	server := api.NewServer(cfg, logger)

	logger.Info("server configuration",
		"port", cfg.Port,
		"max_workers", cfg.MaxWorkers,
		"timeout", cfg.DefaultTimeout,
		"log_level", cfg.LogLevel,
	)

	fmt.Printf("🚀 Server starting on http://localhost:%d\n", cfg.Port)
	fmt.Printf("📊 Dashboard: http://localhost:%d/\n", cfg.Port)
	fmt.Printf("🔍 API: http://localhost:%d/api/v1/check\n", cfg.Port)
	fmt.Printf("💚 Health: http://localhost:%d/api/v1/health\n", cfg.Port)
	fmt.Printf("📈 Metrics: http://localhost:%d/metrics\n", cfg.Port)
	fmt.Println()

	if err := server.Start(); err != nil {
		logger.Error("server failed to start", "error", err)
		os.Exit(1)
	}
}
