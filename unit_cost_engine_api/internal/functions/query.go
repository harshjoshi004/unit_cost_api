package functions

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	ch "github.com/ClickHouse/clickhouse-go/v2"

	"unit-cost-engine-api/internal/models"
	querysql "unit-cost-engine-api/internal/sql"
)

func Query(db ch.Conn, table string, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		var req models.QueryRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid JSON body"})
			return
		}

		statement, args, err := querysql.BuildDataQuery(querysql.DataQueryOptions{
			Table:     table,
			Filters:   req.Filters,
			StartTime: req.StartTime,
			EndTime:   req.EndTime,
			Columns:   req.Columns,
			GroupBy:   req.GroupBy,
			Limit:     req.Limit,
		})
		if err != nil {
			logger.Warn("query validation failed",
				"path", r.URL.Path,
				"filters", req.Filters,
				"columns", req.Columns,
				"group_by", req.GroupBy,
				"error", err,
			)
			writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		rows, err := db.Query(ctx, statement, args...)
		if err != nil {
			logger.Error("query failed", "path", r.URL.Path, "error", err)
			writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "database query failed"})
			return
		}
		defer rows.Close()

		data, err := scanRows(rows)
		if err != nil {
			logger.Error("row scan failed", "path", r.URL.Path, "error", err)
			writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "database row scan failed"})
			return
		}

		logger.Info("data query completed",
			"path", r.URL.Path,
			"filters", req.Filters,
			"columns", req.Columns,
			"group_by", req.GroupBy,
			"rows", len(data),
			"duration_ms", time.Since(start).Milliseconds(),
		)

		writeJSON(w, http.StatusOK, models.QueryResponse{Data: data})
	}
}

type clickhouseRows interface {
	Columns() []string
	Scan(dest ...any) error
	Next() bool
	Err() error
}

func scanRows(rows clickhouseRows) ([]map[string]any, error) {
	columns := rows.Columns()
	result := make([]map[string]any, 0)

	for rows.Next() {
		values := make([]any, len(columns))
		dest := make([]any, len(columns))
		for i := range values {
			dest[i] = &values[i]
		}

		if err := rows.Scan(dest...); err != nil {
			return nil, err
		}

		item := make(map[string]any, len(columns))
		for i, column := range columns {
			item[column] = jsonSafeValue(values[i])
		}
		result = append(result, item)
	}

	return result, rows.Err()
}

func jsonSafeValue(value any) any {
	switch v := value.(type) {
	case time.Time:
		return v.UTC().Format(time.RFC3339)
	default:
		return v
	}
}
