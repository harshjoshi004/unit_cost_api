package metrics

import (
	"context"
	"log/slog"
	"time"

	ch "github.com/ClickHouse/clickhouse-go/v2"

	"unit-cost-engine-api/internal/models"
	querysql "unit-cost-engine-api/internal/sql"
)

func StartUpdater(ctx context.Context, db ch.Conn, table string, logger *slog.Logger, interval time.Duration) {
	go func() {
		update(ctx, db, table, logger)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				update(ctx, db, table, logger)
			}
		}
	}()
}

func update(ctx context.Context, db ch.Conn, table string, logger *slog.Logger) {
	start := time.Now()
	if err := updateCostMetrics(ctx, db, table); err != nil {
		logger.Error("metrics cost update failed", "error", err)
		return
	}
	if err := updateTimestampMetric(ctx, db, table); err != nil {
		logger.Error("metrics timestamp update failed", "error", err)
		return
	}

	logger.Info("metrics updated", "duration_ms", time.Since(start).Milliseconds())
}

func updateCostMetrics(ctx context.Context, db ch.Conn, table string) error {
	queryCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	rows, err := db.Query(queryCtx, querysql.BuildMetricsQuery(table))
	if err != nil {
		return err
	}
	defer rows.Close()

	unitCostGauge.Reset()
	totalCostGauge.Reset()
	totalUsageGauge.Reset()

	for rows.Next() {
		var row costMetricRow
		if err := rows.Scan(
			&row.CloudProvider,
			&row.FinopsEnv,
			&row.Region,
			&row.ResourceType,
			&row.CostUnit,
			&row.UsageUnit,
			&row.TotalCost,
			&row.TotalUsage,
		); err != nil {
			return err
		}

		assetClass := models.AssetClass(row.ResourceType)
		totalCostGauge.WithLabelValues(
			row.CloudProvider,
			row.FinopsEnv,
			row.Region,
			row.ResourceType,
			row.CostUnit,
			assetClass,
		).Set(row.TotalCost)

		totalUsageGauge.WithLabelValues(
			row.CloudProvider,
			row.FinopsEnv,
			row.Region,
			row.ResourceType,
			row.UsageUnit,
			assetClass,
		).Set(row.TotalUsage)

		if row.TotalUsage != 0 {
			unitCostGauge.WithLabelValues(
				row.CloudProvider,
				row.FinopsEnv,
				row.Region,
				row.ResourceType,
				row.CostUnit,
				row.UsageUnit,
				assetClass,
			).Set(row.TotalCost / row.TotalUsage)
		}
	}

	return rows.Err()
}

func updateTimestampMetric(ctx context.Context, db ch.Conn, table string) error {
	queryCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	rows, err := db.Query(queryCtx, querysql.BuildDataTimestampQuery(table))
	if err != nil {
		return err
	}
	defer rows.Close()

	dataTimestampGauge.Reset()

	for rows.Next() {
		var finopsEnv string
		var dataTimestamp time.Time
		if err := rows.Scan(&finopsEnv, &dataTimestamp); err != nil {
			return err
		}

		dataTimestampGauge.WithLabelValues(finopsEnv).Set(float64(dataTimestamp.Unix()))
	}

	return rows.Err()
}

type costMetricRow struct {
	CloudProvider string
	FinopsEnv     string
	Region        string
	ResourceType  string
	CostUnit      string
	UsageUnit     string
	TotalCost     float64
	TotalUsage    float64
}
