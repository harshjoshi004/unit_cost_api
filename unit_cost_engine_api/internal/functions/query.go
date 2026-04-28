package functions

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"time"

	ch "github.com/ClickHouse/clickhouse-go/v2"

	"unit-cost-engine-api/internal/models"
	querysql "unit-cost-engine-api/internal/sql"
)

func UnitCost(db ch.Conn, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		filters := querysql.UnitCostFilters{
			Region:        queryParam(r, "region"),
			CloudProvider: queryParam(r, "cloud_provider"),
			FinopsEnv:     queryParam(r, "finops_env"),
		}

		statement, args := querysql.BuildUnitCostQuery(filters)

		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		rows, err := db.Query(ctx, statement, args...)
		if err != nil {
			logger.Error("unit cost query failed", "path", r.URL.Path, "error", err)
			writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "database query failed"})
			return
		}
		defer rows.Close()

		data := make([]models.UnitCostRow, 0)
		for rows.Next() {
			var row models.UnitCostRow
			var region sql.NullString
			var cloudProvider sql.NullString
			var finopsEnv sql.NullString
			var unitCost sql.NullFloat64

			if err := rows.Scan(&row.InstanceType, &region, &cloudProvider, &finopsEnv, &unitCost); err != nil {
				logger.Error("unit cost row scan failed", "path", r.URL.Path, "error", err)
				writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "database row scan failed"})
				return
			}

			row.Region = nullString(region)
			row.CloudProvider = nullString(cloudProvider)
			row.FinopsEnv = nullString(finopsEnv)
			if unitCost.Valid {
				row.UnitCost = unitCost.Float64
			}

			data = append(data, row)
		}

		if err := rows.Err(); err != nil {
			logger.Error("unit cost rows failed", "path", r.URL.Path, "error", err)
			writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "database row scan failed"})
			return
		}

		logger.Info("unit cost query completed",
			"path", r.URL.Path,
			"region", filterValue(filters.Region),
			"cloud_provider", filterValue(filters.CloudProvider),
			"finops_env", filterValue(filters.FinopsEnv),
			"rows", len(data),
			"duration_ms", time.Since(start).Milliseconds(),
		)

		writeJSON(w, http.StatusOK, data)
	}
}

func queryParam(r *http.Request, key string) *string {
	value := r.URL.Query().Get(key)
	if value == "" {
		return nil
	}
	return &value
}

func nullString(value sql.NullString) string {
	if !value.Valid {
		return ""
	}
	return value.String
}

func filterValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
