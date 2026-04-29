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
  chart     = "./unitcost-engine"
  namespace = "k8s-cost-insights"

  values = [
    yamlencode({
      # ------------------------------------------------------------
      # Image
      # ------------------------------------------------------------
      image = {
        repository  = "spideralxjoshi/unit-cost-api"
        tag         = "f2dd411fbbaaa921b5c8001e1fe162c4f8bf50b1"
        pullPolicy  = "IfNotPresent"
      }

      # ------------------------------------------------------------
      # Deployment
      # ------------------------------------------------------------
      replicaCount = 1

      # ------------------------------------------------------------
      # Service
      # ------------------------------------------------------------
      service = {
        name = "unitcost-engine"
        type = "ClusterIP"
        port = 7000
      }

      # ------------------------------------------------------------
      # Ingress
      # ------------------------------------------------------------
      ingress = {
        enabled  = true
        className = "traefik"
        host     = "unitcost.127.0.0.1.nip.io"
        path     = "/"
      }

      # ------------------------------------------------------------
      # Secrets (IMPORTANT)
      # ------------------------------------------------------------
      envFromSecret = "unitcost-engine-secret"

      # ------------------------------------------------------------
      # Resources
      # ------------------------------------------------------------
      resources = {
        requests = {
          cpu    = "200m"
          memory = "256Mi"
        }
        limits = {
          cpu    = "1"
          memory = "1Gi"
        }
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