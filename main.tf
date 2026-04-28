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

resource "helm_release" "jenkins" {
  name       = "jenkins"
  repository = "https://charts.jenkins.io"
  chart      = "jenkins"
  namespace  = "default"

  values = [
    yamlencode({
      controller = {
        serviceType = "ClusterIP"
      }
    })
  ]
}

resource "helm_release" "unit_cost_api" {
  name      = "unit-cost-api"
  chart     = "api-chart"
  namespace = "default"

  values = [
    yamlencode({
      image = {
        repository = "spideralxjoshi/unit-cost-api"
        tag        = "latest"
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
        CLICKHOUSE_ADDR     = "clickhouse:9000"
        CLICKHOUSE_DATABASE = "default"
        CLICKHOUSE_USERNAME = "default"
        CLICKHOUSE_PASSWORD = ""
        CLICKHOUSE_TABLE    = "cloud_costs"
      }
    })
  ]

  depends_on = [
    helm_release.prometheus
  ]
}
