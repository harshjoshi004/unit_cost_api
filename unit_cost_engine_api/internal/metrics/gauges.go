package metrics

import "github.com/prometheus/client_golang/prometheus"

var unitCostGauge = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "finops_unit_cost",
		Help: "Unit cost calculated as SUM(total_cost) / SUM(total_usage) for the latest snapshot.",
	},
	[]string{"cloud_provider", "finops_env", "region", "resource_type", "cost_unit", "usage_unit", "assetClass"},
)

var totalCostGauge = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "finops_total_cost",
		Help: "Total cost calculated as SUM(total_cost) for the latest snapshot.",
	},
	[]string{"cloud_provider", "finops_env", "region", "resource_type", "cost_unit", "assetClass"},
)

var totalUsageGauge = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "finops_total_usage",
		Help: "Total usage calculated as SUM(total_usage) for the latest snapshot.",
	},
	[]string{"cloud_provider", "finops_env", "region", "resource_type", "usage_unit", "assetClass"},
)

var dataTimestampGauge = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "finops_data_timestamp",
		Help: "Latest month_year timestamp available, exposed as Unix seconds.",
	},
	[]string{"finops_env"},
)

func init() {
	prometheus.MustRegister(unitCostGauge)
	prometheus.MustRegister(totalCostGauge)
	prometheus.MustRegister(totalUsageGauge)
	prometheus.MustRegister(dataTimestampGauge)
}
