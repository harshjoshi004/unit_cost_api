package router

import (
	"log/slog"
	"net/http"
	"time"

	ch "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"unit-cost-engine-api/internal/functions"
)

func New(db ch.Conn, table string, logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/v1/health", functions.Health())
	mux.HandleFunc("GET /api/v1/unit-cost", functions.UnitCost(db, table, logger))
	mux.HandleFunc("GET /api/v1/data/opencost", functions.OpenCost(db, table, logger))
	mux.Handle("GET /metrics", promhttp.Handler())

	return logRequests(mux, logger)
}

type responseRecorder struct {
	http.ResponseWriter
	status int
}

func (r *responseRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func logRequests(next http.Handler, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &responseRecorder{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(recorder, r)

		logger.Info("request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"query", r.URL.RawQuery,
			"status", recorder.status,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}
