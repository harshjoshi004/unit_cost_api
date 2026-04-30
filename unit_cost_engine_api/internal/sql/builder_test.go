package sql

import (
	"strings"
	"testing"
)

func TestBuildUnitCostQueryNoFilters(t *testing.T) {
	query, args := BuildUnitCostQuery("unit_cost.unit_cost_table", UnitCostFilters{})

	expected := "SELECT resource_type AS instance_type, region, cloud_provider, finops_env, sum(total_cost) / nullIf(sum(total_usage), 0) AS unit_cost FROM unit_cost.unit_cost_table WHERE month_year = (SELECT max(month_year) FROM unit_cost.unit_cost_table) GROUP BY resource_type, region, cloud_provider, finops_env ORDER BY resource_type"
	if query != expected {
		t.Fatalf("query mismatch\nwant: %s\n got: %s", expected, query)
	}

	if len(args) != 0 {
		t.Fatalf("unexpected args: %#v", args)
	}
}

func TestBuildUnitCostQueryWithFilters(t *testing.T) {
	region := "eastus"
	cloudProvider := "azure"
	finopsEnv := "prod"

	query, args := BuildUnitCostQuery("unit_cost.unit_cost_table", UnitCostFilters{
		Region:        &region,
		CloudProvider: &cloudProvider,
		FinopsEnv:     &finopsEnv,
	})

	expected := "SELECT resource_type AS instance_type, region, cloud_provider, finops_env, sum(total_cost) / nullIf(sum(total_usage), 0) AS unit_cost FROM unit_cost.unit_cost_table WHERE month_year = (SELECT max(month_year) FROM unit_cost.unit_cost_table) AND region = ? AND cloud_provider = ? AND finops_env = ? GROUP BY resource_type, region, cloud_provider, finops_env ORDER BY resource_type"
	if query != expected {
		t.Fatalf("query mismatch\nwant: %s\n got: %s", expected, query)
	}

	if len(args) != 3 || args[0] != region || args[1] != cloudProvider || args[2] != finopsEnv {
		t.Fatalf("unexpected args: %#v", args)
	}
}

func TestBuildUnitCostQueryUsesSafeWeightedMath(t *testing.T) {
	query, _ := BuildUnitCostQuery("unit_cost.unit_cost_table", UnitCostFilters{})

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
