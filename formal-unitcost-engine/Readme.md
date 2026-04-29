# Unit Cost Engine API

Helm chart for deploying the **Unit Cost Engine API**, a service that computes instance-type–level unit costs from ClickHouse cost data and exposes them via HTTP endpoints and Prometheus metrics.

---

## Overview

The Unit Cost Engine calculates:

```
unit_cost = total_cost / total_usage
```

Key characteristics:

* Aggregation level: **instance type (`resource_type`)**
* Filters supported:

  * `region`
  * `cloud_provider`
  * `finops_env`
* Data source: ClickHouse
* Output formats:

  * JSON (API)
  * CSV (OpenCost-compatible)
  * Prometheus metrics

---

## Prerequisites

The following must be available before installing the chart:

* A running ClickHouse cluster accessible from the Kubernetes cluster
* A namespace (e.g. `k8s-cost-insights`)
* A Kubernetes Secret containing ClickHouse connection details

---

## Required Secret

This chart does not create secrets. You must create the required secret prior to installation.

```bash
kubectl apply -f secrets/unitcost-engine-secret.yaml -n k8s-cost-insights
```

### Required keys

```yaml
CLICKHOUSE_ADDR=
CLICKHOUSE_DATABASE=
CLICKHOUSE_USERNAME=
CLICKHOUSE_PASSWORD=
CLICKHOUSE_TABLE=
```

### Example

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: unitcost-engine-secret
  namespace: k8s-cost-insights
type: Opaque
stringData:
  CLICKHOUSE_ADDR: clickhouse.k8s-cost-insights.svc.cluster.local:9000
  CLICKHOUSE_DATABASE: test
  CLICKHOUSE_USERNAME: default
  CLICKHOUSE_PASSWORD: ""
  CLICKHOUSE_TABLE: data
```

---

## Installation

```bash
helm install unit-cost-api ./charts/unit-cost-api \
  -n k8s-cost-insights
```

To override the image tag:

```bash
helm install unit-cost-api ./charts/unit-cost-api \
  -n k8s-cost-insights \
  --set image.tag=<image-tag>
```

---

## Configuration

Key configurable values in `values.yaml`:

```yaml
image:
  repository: spideralxjoshi/unit-cost-api
  tag: latest
  pullPolicy: IfNotPresent

service:
  port: 7000

resources: {}
```

---

## API Endpoints

### Unit Cost API

```
GET /api/v1/unit-cost
```

#### Query Parameters

| Parameter      | Required | Description              |
| -------------- | -------- | ------------------------ |
| region         | No       | Filter by region         |
| cloud_provider | No       | Filter by cloud provider |
| finops_env     | No       | Filter by environment    |

#### Example

```bash
curl "http://<host>/api/v1/unit-cost?region=eastus&cloud_provider=azure&finops_env=prod"
```

#### Response

```json
[
  {
    "instance_type": "D2s_v3",
    "region": "eastus",
    "cloud_provider": "azure",
    "finops_env": "prod",
    "unit_cost": 0.28
  }
]
```

---

### OpenCost-Compatible Endpoint

```
GET /api/v1/data/opencost
```

* Returns CSV output
* Intended for integration with OpenCost-compatible systems

---

### Prometheus Metrics

```
GET /metrics
```

Exposes metrics such as:

* `finops_total_cost`
* `finops_unit_cost`

---

## Data Contract (ClickHouse)

The configured table must contain the following columns:

```sql
resource_type   -- instance type
region
cloud_provider
finops_env
total_cost
total_usage
month_year
```

---

## Deployment Notes

* Connects to ClickHouse using the native TCP protocol (port 9000)
* Queries operate on the latest snapshot using `max(month_year)`
* Time Window query currently not supported
* Stateless service
* Designed for use behind an Ingress
