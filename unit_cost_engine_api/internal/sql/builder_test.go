package sql

import (
	"strings"
	"testing"
	"time"
)

func TestBuildDataQueryLatestSnapshotGrouped(t *testing.T) {
	query, args, err := BuildDataQuery(DataQueryOptions{
		Table: "cloud_costs",
		Filters: map[string]string{
			"region":     "asia-south1",
			"finops_env": "prod",
		},
		Columns: []string{"resource_type", "total_cost"},
		GroupBy: []string{"resource_type"},
	})
	if err != nil {
		t.Fatalf("BuildDataQuery returned error: %v", err)
	}

	expected := "SELECT resource_type, sum(total_cost) AS total_cost FROM cloud_costs WHERE month_year = (SELECT max(month_year) FROM cloud_costs) AND finops_env = ? AND region = ? GROUP BY resource_type ORDER BY resource_type LIMIT ?"
	if query != expected {
		t.Fatalf("query mismatch\nwant: %s\n got: %s", expected, query)
	}

	if len(args) != 3 || args[0] != "prod" || args[1] != "asia-south1" || args[2] != DefaultLimit {
		t.Fatalf("unexpected args: %#v", args)
	}
}

func TestBuildDataQueryTimeRange(t *testing.T) {
	start := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 4, 30, 23, 59, 59, 0, time.UTC)

	query, args, err := BuildDataQuery(DataQueryOptions{
		Table:     "cloud_costs",
		StartTime: &start,
		EndTime:   &end,
		Columns:   []string{"resource_type", "region", "total_usage"},
		GroupBy:   []string{"resource_type", "region"},
	})
	if err != nil {
		t.Fatalf("BuildDataQuery returned error: %v", err)
	}

	if !strings.Contains(query, "month_year >= ? AND month_year <= ?") {
		t.Fatalf("query did not contain time range: %s", query)
	}

	if len(args) != 3 || args[0] != start || args[1] != end || args[2] != DefaultLimit {
		t.Fatalf("unexpected args: %#v", args)
	}
}

func TestBuildDataQueryUnitCostUsesSafeWeightedMath(t *testing.T) {
	query, _, err := BuildDataQuery(DataQueryOptions{
		Table:   "cloud_costs",
		Columns: []string{"resource_type", "unit_cost"},
		GroupBy: []string{"resource_type"},
	})
	if err != nil {
		t.Fatalf("BuildDataQuery returned error: %v", err)
	}

	if !strings.Contains(query, "sum(total_cost) / nullIf(sum(total_usage), 0) AS unit_cost") {
		t.Fatalf("query did not contain safe unit_cost math: %s", query)
	}
}

func TestBuildOpenCostQueryUsesSafeWeightedMathAndOrdering(t *testing.T) {
	query, _ := BuildOpenCostQuery("cloud_costs")

	if !strings.Contains(query, "sum(total_cost) / nullIf(sum(total_usage), 0) AS unit_cost") {
		t.Fatalf("query did not contain safe unit_cost math: %s", query)
	}

	if !strings.Contains(query, "ORDER BY resource_type, region") {
		t.Fatalf("query did not contain stable ordering: %s", query)
	}
}

func TestBuildDataQueryCustomLimit(t *testing.T) {
	limit := 250

	_, args, err := BuildDataQuery(DataQueryOptions{
		Table:   "cloud_costs",
		Columns: []string{"resource_type"},
		Limit:   &limit,
	})
	if err != nil {
		t.Fatalf("BuildDataQuery returned error: %v", err)
	}

	if len(args) != 1 || args[0] != limit {
		t.Fatalf("unexpected args: %#v", args)
	}
}

func TestBuildDataQueryRejectsTooLargeLimit(t *testing.T) {
	limit := MaxLimit + 1

	_, _, err := BuildDataQuery(DataQueryOptions{
		Table:   "cloud_costs",
		Columns: []string{"resource_type"},
		Limit:   &limit,
	})
	if err == nil {
		t.Fatal("expected invalid limit error")
	}
}

func TestBuildDataQueryRejectsUnsafeColumn(t *testing.T) {
	_, _, err := BuildDataQuery(DataQueryOptions{
		Table:   "cloud_costs",
		Columns: []string{"resource_type; DROP TABLE users"},
	})
	if err == nil {
		t.Fatal("expected invalid column error")
	}
}

func TestBuildDataQueryRequiresBothTimes(t *testing.T) {
	start := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)

	_, _, err := BuildDataQuery(DataQueryOptions{
		Table:     "cloud_costs",
		StartTime: &start,
		Columns:   []string{"resource_type"},
	})
	if err == nil {
		t.Fatal("expected missing end_time error")
	}
}
