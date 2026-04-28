# Unit Cost Engine API

This project is a small Go API that exposes processed cloud cost data from ClickHouse.

It does not write data and it does not run cost calculations pipelines. Upstream jobs already aggregate the data. This API is only a read layer for:

- Grafana dashboards through Prometheus metrics
- OpenCost through CSV ingestion
- Internal tools and CI pipelines through JSON

## Folder Walkthrough

```text
/cmd/server.go
```

Starts the HTTP server, connects to ClickHouse, starts the metrics updater, and registers routes.

```text
/internal/router/router.go
```

Defines the HTTP routes and request logging middleware.

```text
/internal/functions/query.go
```

Handles `POST /api/v1/data/query`. It validates the request, asks the SQL builder for a safe query, runs it, and returns JSON.

```text
/internal/functions/opencost.go
```

Handles `GET /api/v1/data/opencost`. It converts the latest ClickHouse snapshot into the CSV shape OpenCost expects.

```text
/internal/functions/health.go
```

Simple health check endpoint.

```text
/internal/sql/builder.go
/internal/sql/filters.go
```

Builds ClickHouse SQL from whitelisted columns, filters, and groupings. Endpoint functions do not contain raw SQL.

```text
/internal/metrics/gauges.go
/internal/metrics/updater.go
```

Defines Prometheus gauges and refreshes them in the background.

```text
/internal/models
```

Request/response structs and reusable asset class detection.

```text
/pkg/clickhouse/connection.go
```

Reads ClickHouse environment variables and opens the database connection.

## Request Flow

```text
router -> function -> sql -> db -> response
```

The router maps a path to a function. The function validates HTTP input and calls the SQL builder. The SQL builder only accepts known-safe column names. The function then queries ClickHouse and writes the response.

## API

Base path:

```text
/api/v1
```

### Health

```bash
curl http://localhost:7000/api/v1/health
```

### Flexible Data Query

```bash
curl -X POST http://localhost:7000/api/v1/data/query \
  -H 'Content-Type: application/json' \
  -d '{
    "filters": {
      "finops_env": "prod",
      "region": "asia-south1"
    },
    "start_time": "2026-04-01T00:00:00Z",
    "end_time": "2026-04-30T23:59:59Z",
    "columns": ["resource_type", "total_cost"],
    "group_by": ["resource_type"],
    "limit": 1000
  }'
```

If `group_by` is present, aggregate columns are grouped:

- `total_cost` -> `sum(total_cost)`
- `total_usage` -> `sum(total_usage)`
- `unit_cost` -> `SUM(total_cost) / nullIf(SUM(total_usage), 0)`

If `start_time` and `end_time` are missing, the API returns the latest `month_year` snapshot.

`limit` is optional. If it is missing, the API uses `1000`. The maximum allowed value is `10000`.

When `SUM(total_usage)` is zero, grouped `unit_cost` returns `null` instead of dividing by zero.

### OpenCost CSV

```bash
curl http://localhost:7000/api/v1/data/opencost
```

CSV columns:

```text
EndTimestamp,InstanceID,Region,AssetClass,InstanceIDField,InstanceType,MarketPriceHourly,Version
```

Mapping:

- `resource_type` -> `InstanceType`
- `region` -> `Region`
- `unit_cost` -> `MarketPriceHourly`
- `InstanceID` is always `placeholder`
- `InstanceIDField` is empty
- `Version` is always `v1`

`AssetClass` is inferred from `resource_type`:

- GPU-like names -> `gpu`
- disk/storage/PVC-like names -> `pv`
- everything else -> `node`

## Prometheus Metrics

Metrics are exposed at:

```text
/metrics
```

The background updater refreshes every 45 seconds.

Metrics:

- `finops_unit_cost`: `SUM(total_cost) / SUM(total_usage)`, skipped when usage is zero
- `finops_total_cost`: `SUM(total_cost)`
- `finops_total_usage`: `SUM(total_usage)`
- `finops_data_timestamp`: `MAX(month_year)` as Unix seconds

## Environment Variables

```text
PORT=7000
CLICKHOUSE_ADDR=localhost:9000
CLICKHOUSE_DATABASE=default
CLICKHOUSE_USERNAME=default
CLICKHOUSE_PASSWORD=
CLICKHOUSE_TABLE=cloud_costs
```

`CLICKHOUSE_TABLE` defaults to `cloud_costs` because the schema was provided but the table name was not. Set it to the real ClickHouse table name before running.

## Local Setup With Rancher Desktop

1. Start Rancher Desktop.
2. Make sure the Kubernetes context points to the cluster where ClickHouse is running.
3. Expose ClickHouse locally through an ingress or port-forward. Example:

```bash
kubectl port-forward svc/clickhouse 9000:9000
```

4. Export environment variables:

```bash
export CLICKHOUSE_ADDR=localhost:9000
export CLICKHOUSE_DATABASE=default
export CLICKHOUSE_USERNAME=default
export CLICKHOUSE_PASSWORD=''
export CLICKHOUSE_TABLE=cloud_costs
export PORT=7000
```

5. Run the API:

```bash
go run ./cmd
```

## Quick Test Script

```python
import requests

base = "http://localhost:7000"

payload = {
    "filters": {"finops_env": "prod"},
    "columns": ["resource_type", "region", "total_cost"],
    "group_by": ["resource_type", "region"],
}

print(requests.post(f"{base}/api/v1/data/query", json=payload).json())
print(requests.get(f"{base}/api/v1/data/opencost").text[:500])
print(requests.get(f"{base}/metrics").text[:500])
```
