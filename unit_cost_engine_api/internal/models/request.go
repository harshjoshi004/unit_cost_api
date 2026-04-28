package models

import "time"

type QueryRequest struct {
	Filters   map[string]string `json:"filters"`
	StartTime *time.Time        `json:"start_time"`
	EndTime   *time.Time        `json:"end_time"`
	Columns   []string          `json:"columns"`
	GroupBy   []string          `json:"group_by"`
	Limit     *int              `json:"limit"`
}
