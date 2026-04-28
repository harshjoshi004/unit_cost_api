package models

type UnitCostRow struct {
	InstanceType  string  `json:"instance_type"`
	Region        string  `json:"region"`
	CloudProvider string  `json:"cloud_provider"`
	FinopsEnv     string  `json:"finops_env"`
	UnitCost      float64 `json:"unit_cost"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
