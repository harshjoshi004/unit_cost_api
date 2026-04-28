package sql

import "fmt"

var allowedFilters = map[string]bool{
	"cloud_provider": true,
	"finops_env":     true,
	"region":         true,
	"resource_type":  true,
	"cost_unit":      true,
	"usage_unit":     true,
}

var allowedColumns = map[string]bool{
	"cloud_provider": true,
	"finops_env":     true,
	"region":         true,
	"month_year":     true,
	"resource_type":  true,
	"total_cost":     true,
	"cost_unit":      true,
	"total_usage":    true,
	"usage_unit":     true,
	"unit_cost":      true,
	"execution_day":  true,
}

var groupableColumns = map[string]bool{
	"cloud_provider": true,
	"finops_env":     true,
	"region":         true,
	"month_year":     true,
	"resource_type":  true,
	"cost_unit":      true,
	"usage_unit":     true,
	"execution_day":  true,
}

var aggregateColumns = map[string]bool{
	"total_cost":  true,
	"total_usage": true,
	"unit_cost":   true,
}

func validateFilters(filters map[string]string) error {
	for column := range filters {
		if !allowedFilters[column] {
			return fmt.Errorf("filter %q is not allowed", column)
		}
	}
	return nil
}

func validateColumns(columns []string) error {
	if len(columns) == 0 {
		return fmt.Errorf("columns is required")
	}

	seen := map[string]bool{}
	for _, column := range columns {
		if !allowedColumns[column] {
			return fmt.Errorf("column %q is not allowed", column)
		}
		if seen[column] {
			return fmt.Errorf("column %q is duplicated", column)
		}
		seen[column] = true
	}
	return nil
}

func validateGroupBy(columns []string, groupBy []string) error {
	if len(groupBy) == 0 {
		return nil
	}

	selected := map[string]bool{}
	for _, column := range columns {
		selected[column] = true
	}

	grouped := map[string]bool{}
	for _, column := range groupBy {
		if !groupableColumns[column] {
			return fmt.Errorf("group_by column %q is not allowed", column)
		}
		if grouped[column] {
			return fmt.Errorf("group_by column %q is duplicated", column)
		}
		if !selected[column] {
			return fmt.Errorf("group_by column %q must also be in columns", column)
		}
		grouped[column] = true
	}

	for _, column := range columns {
		if aggregateColumns[column] {
			continue
		}
		if !grouped[column] {
			return fmt.Errorf("column %q must be in group_by or be an aggregate column", column)
		}
	}

	return nil
}
