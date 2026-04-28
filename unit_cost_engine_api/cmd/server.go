package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"unit-cost-engine-api/internal/metrics"
	"unit-cost-engine-api/internal/router"
	"unit-cost-engine-api/pkg/clickhouse"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	cfg, err := clickhouse.ConfigFromEnv()
	if err != nil {
		logger.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	db, err := clickhouse.Open(cfg)
	if err != nil {
		logger.Error("clickhouse connection failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	metrics.StartUpdater(context.Background(), db, cfg.Table, logger, 45*time.Second)

	port := getEnv("PORT", "7000")
	server := &http.Server{
		Addr:              ":" + port,
		Handler:           router.New(db, cfg.Table, logger),
		ReadHeaderTimeout: 10 * time.Second,
	}

	logger.Info("server started", "port", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
