package sql

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type DataQueryOptions struct {
	Table     string
	Filters   map[string]string
	StartTime *time.Time
	EndTime   *time.Time
	Columns   []string
	GroupBy   []string
	Limit     *int
}

const (
	DefaultLimit = 1000
	MaxLimit     = 10000
)

func BuildDataQuery(opts DataQueryOptions) (string, []any, error) {
	if err := validateColumns(opts.Columns); err != nil {
		return "", nil, err
	}
	if err := validateFilters(opts.Filters); err != nil {
		return "", nil, err
	}
	if err := validateGroupBy(opts.Columns, opts.GroupBy); err != nil {
		return "", nil, err
	}
	if (opts.StartTime == nil) != (opts.EndTime == nil) {
		return "", nil, fmt.Errorf("start_time and end_time must be provided together")
	}
	if opts.StartTime != nil && opts.StartTime.After(*opts.EndTime) {
		return "", nil, fmt.Errorf("start_time must be before end_time")
	}
	limit, err := validateLimit(opts.Limit)
	if err != nil {
		return "", nil, err
	}

	args := make([]any, 0)
	selects := buildSelects(opts.Columns, len(opts.GroupBy) > 0)
	where, whereArgs := buildWhere(opts.Table, opts.Filters, opts.StartTime, opts.EndTime)
	args = append(args, whereArgs...)

	query := "SELECT " + strings.Join(selects, ", ") +
		" FROM " + opts.Table +
		" WHERE " + strings.Join(where, " AND ")

	if len(opts.GroupBy) > 0 {
		groupBy := buildGroupBy(opts.GroupBy)
		query += " GROUP BY " + strings.Join(groupBy, ", ")
		query += " ORDER BY " + strings.Join(groupBy, ", ")
	} else {
		query += " ORDER BY month_year, resource_type, ifNull(region, '')"
	}
	query += " LIMIT ?"
	args = append(args, limit)

	return query, args, nil
}

func BuildOpenCostQuery(table string) (string, []any) {
	query := "SELECT " +
		"max(month_year) AS end_timestamp, " +
		"resource_type, " +
		"ifNull(region, '') AS region, " +
		"sum(total_cost) / nullIf(sum(total_usage), 0) AS unit_cost " +
		"FROM " + table + " " +
		"WHERE month_year = (SELECT max(month_year) FROM " + table + ") " +
		"GROUP BY resource_type, ifNull(region, '') " +
		"ORDER BY resource_type, ifNull(region, '')"

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
		"GROUP BY cloud_provider, ifNull(finops_env, ''), ifNull(region, ''), resource_type, cost_unit, usage_unit"
}

func BuildDataTimestampQuery(table string) string {
	return "SELECT ifNull(finops_env, '') AS finops_env, max(month_year) AS data_timestamp " +
		"FROM " + table + " GROUP BY ifNull(finops_env, '')"
}

func buildSelects(columns []string, grouped bool) []string {
	selects := make([]string, 0, len(columns))
	for _, column := range columns {
		if grouped && aggregateColumns[column] {
			selects = append(selects, aggregateSelect(column))
			continue
		}
		selects = append(selects, nullableSelect(column))
	}
	return selects
}

func aggregateSelect(column string) string {
	switch column {
	case "total_cost":
		return "sum(total_cost) AS total_cost"
	case "total_usage":
		return "sum(total_usage) AS total_usage"
	case "unit_cost":
		return "sum(total_cost) / nullIf(sum(total_usage), 0) AS unit_cost"
	default:
		return column
	}
}

func nullableSelect(column string) string {
	switch column {
	case "finops_env", "region":
		return "ifNull(" + column + ", '') AS " + column
	default:
		return column
	}
}

func buildGroupBy(columns []string) []string {
	groupBy := make([]string, 0, len(columns))
	for _, column := range columns {
		switch column {
		case "finops_env", "region":
			groupBy = append(groupBy, "ifNull("+column+", '')")
		default:
			groupBy = append(groupBy, column)
		}
	}
	return groupBy
}

func validateLimit(limit *int) (int, error) {
	if limit == nil {
		return DefaultLimit, nil
	}
	if *limit <= 0 {
		return 0, fmt.Errorf("limit must be greater than 0")
	}
	if *limit > MaxLimit {
		return 0, fmt.Errorf("limit cannot be greater than %d", MaxLimit)
	}
	return *limit, nil
}

func buildWhere(table string, filters map[string]string, startTime *time.Time, endTime *time.Time) ([]string, []any) {
	where := make([]string, 0, len(filters)+1)
	args := make([]any, 0, len(filters)+2)

	if startTime != nil && endTime != nil {
		where = append(where, "month_year >= ?")
		args = append(args, startTime.UTC())
		where = append(where, "month_year <= ?")
		args = append(args, endTime.UTC())
	} else {
		where = append(where, "month_year = (SELECT max(month_year) FROM "+table+")")
	}

	filterColumns := make([]string, 0, len(filters))
	for column := range filters {
		filterColumns = append(filterColumns, column)
	}
	sort.Strings(filterColumns)

	for _, column := range filterColumns {
		where = append(where, column+" = ?")
		args = append(args, filters[column])
	}

	return where, args
}
