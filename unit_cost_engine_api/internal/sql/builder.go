package sql

import "strings"

const (
	unitCostTable = "test.data"
)

type UnitCostFilters struct {
	Region        *string
	CloudProvider *string
	FinopsEnv     *string
}

func BuildUnitCostQuery(filters UnitCostFilters) (string, []any) {
	where := []string{"month_year = (SELECT max(month_year) FROM " + unitCostTable + ")"}
	args := make([]any, 0, 3)

	if filters.Region != nil {
		where = append(where, "region = ?")
		args = append(args, *filters.Region)
	}
	if filters.CloudProvider != nil {
		where = append(where, "cloud_provider = ?")
		args = append(args, *filters.CloudProvider)
	}
	if filters.FinopsEnv != nil {
		where = append(where, "finops_env = ?")
		args = append(args, *filters.FinopsEnv)
	}

	query := "SELECT " +
		"resource_type AS instance_type, " +
		"region, " +
		"cloud_provider, " +
		"finops_env, " +
		"sum(total_cost) / nullIf(sum(total_usage), 0) AS unit_cost " +
		"FROM " + unitCostTable + " " +
		"WHERE " + strings.Join(where, " AND ") + " " +
		"GROUP BY resource_type, region, cloud_provider, finops_env " +
		"ORDER BY resource_type"

	return query, args
}

func BuildOpenCostQuery(table string) (string, []any) {
	query := "SELECT " +
		"max(month_year) AS end_timestamp, " +
		"resource_type, " +
		"ifNull(region, '') AS region, " +
		"sum(total_cost) / nullIf(sum(total_usage), 0) AS unit_cost " +
		"FROM " + table + " " +
		"WHERE month_year = (SELECT max(month_year) FROM " + table + ") " +
		"GROUP BY resource_type, region " +   
		"ORDER BY resource_type, region"

	return query, nil
}

func BuildMetricsQuery(table string) string {
	return "SELECT " +
		"cloud_provider, " +
		"ifNull(finops_env, '') AS finops_env, " +
		"ifNull(region, '') AS region, " +
		"resource_type, " +
		"cost_unit, " +
		"usage_unit, " +
		"sum(total_cost) AS total_cost, " +
		"sum(total_usage) AS total_usage " +
		"FROM " + table + " " +
		"WHERE month_year = (SELECT max(month_year) FROM " + table + ") " +
		"GROUP BY cloud_provider, finops_env, region, resource_type, cost_unit, usage_unit"
}

func BuildDataTimestampQuery(table string) string {
	return "SELECT ifNull(finops_env, '') AS finops_env, max(month_year) AS data_timestamp " +
		"FROM " + table + " GROUP BY finops_env"
}
