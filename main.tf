variable "image_tag" {
  type = string
}

resource "helm_release" "prometheus" {
  name       = "prometheus"
  repository = "https://prometheus-community.github.io/helm-charts"
  chart      = "kube-prometheus-stack"
  namespace  = "default"

  values = [
    yamlencode({
      grafana = {
        enabled = true
      }

      prometheus = {
        prometheusSpec = {
          scrapeInterval = "30s"
        }
      }

      defaultRules = {
        disabled = {
          NodeClockNotSynchronising = true
        }
      }
    })
  ]
}

resource "helm_release" "unit_cost_api" {
  name      = "unit-cost-api"
  chart = "./api-chart"
  namespace = "k8s-cost-insights"

  values = [
    yamlencode({
      image = {
        repository = "spideralxjoshi/unit-cost-api"
        tag        = var.image_tag   
      }

      service = {
        name = "api"
        port = 7000
      }

      ingress = {
        enabled = true
        host    = "unitcost.127.0.0.1.nip.io"
      }

      env = {
        CLICKHOUSE_ADDR     = "clickhouse.k8s-cost-insights.svc.cluster.local:9000"
        CLICKHOUSE_DATABASE = "test"
        CLICKHOUSE_USERNAME = "metabasetest"
        CLICKHOUSE_PASSWORD = "123456"
        CLICKHOUSE_TABLE    = "data"
      }
    })
  ]

  depends_on = [
    helm_release.prometheus
  ]
}

resource "helm_release" "api_monitor" {
  name      = "api-monitor"
  chart = "./service-monitor"
  namespace = "k8s-cost-insights"

  values = [
    yamlencode({
      name = "api-monitor"

      selector = {
        app = "api"
      }

      endpoint = {
        port     = "http"
        path     = "/metrics"
        interval = "30s"
      }
    })
  ]

  depends_on = [
    helm_release.prometheus,
    helm_release.unit_cost_api
  ]
}