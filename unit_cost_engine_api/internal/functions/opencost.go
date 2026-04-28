package functions

import (
	"context"
	"database/sql"
	"encoding/csv"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	ch "github.com/ClickHouse/clickhouse-go/v2"

	"unit-cost-engine-api/internal/models"
	querysql "unit-cost-engine-api/internal/sql"
)

func OpenCost(db ch.Conn, table string, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		statement, args := querysql.BuildOpenCostQuery(table)

		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		rows, err := db.Query(ctx, statement, args...)
		if err != nil {
			logger.Error("opencost query failed", "path", r.URL.Path, "error", err)
			writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "database query failed"})
			return
		}
		defer rows.Close()

		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", `attachment; filename="opencost-prices.csv"`)
		w.WriteHeader(http.StatusOK)

		writer := csv.NewWriter(w)
		defer writer.Flush()

		_ = writer.Write([]string{
			"EndTimestamp",
			"InstanceID",
			"Region",
			"AssetClass",
			"InstanceIDField",
			"InstanceType",
			"MarketPriceHourly",
			"Version",
		})

		count := 0
		for rows.Next() {
			var endTimestamp time.Time
			var resourceType string
			var region string
			var unitCost sql.NullFloat64

			if err := rows.Scan(&endTimestamp, &resourceType, &region, &unitCost); err != nil {
				logger.Error("opencost row scan failed", "path", r.URL.Path, "error", err)
				return
			}

			_ = writer.Write([]string{
				endTimestamp.UTC().Format(time.RFC3339),
				"placeholder",
				region,
				models.AssetClass(resourceType),
				"",
				resourceType,
				formatNullableFloat(unitCost),
				"v1",
			})
			count++
		}

		if err := rows.Err(); err != nil {
			logger.Error("opencost rows failed", "path", r.URL.Path, "error", err)
			return
		}

		logger.Info("opencost csv completed",
			"path", r.URL.Path,
			"rows", count,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	}
}

func formatNullableFloat(value sql.NullFloat64) string {
	if !value.Valid {
		return ""
	}
	return strconv.FormatFloat(value.Float64, 'f', -1, 64)
}
