pipeline {
  agent {
    kubernetes {
      yaml """
apiVersion: v1
kind: Pod
spec:
  containers:
    - name: golang
      image: golang:1.24
      command: ["sh", "-c", "sleep infinity"]

    - name: kaniko
      image: gcr.io/kaniko-project/executor:debug
      command: ["/busybox/sh", "-c", "sleep infinity"]

    - name: helm
      image: alpine/helm:3.14.4
      command: ["sh", "-c", "sleep infinity"]
"""
    }
  }

  environment {
    IMAGE = "spideralxjoshi/unit-cost-api"
    TAG   = "${BUILD_NUMBER}"
  }

  stages {

    stage('Checkout') {
      steps {
        git url: 'Yhttps://github.com/harshjoshi004/unit_cost_api', branch: 'main'
      }
    }

    stage('Unit Tests') {
      steps {
        container('golang') {
          sh '''
            go mod tidy
            go test ./...
          '''
        }
      }
    }

    stage('Build & Push Image') {
      steps {
        container('kaniko') {
          withCredentials([usernamePassword(
            credentialsId: 'dockerhub-creds',
            usernameVariable: 'DOCKER_USER',
            passwordVariable: 'DOCKER_PASS'
          )]) {
            sh '''
mkdir -p /kaniko/.docker

cat > /kaniko/.docker/config.json <<EOF
{
  "auths": {
    "registry-1.docker.io": {
      "username": "$DOCKER_USER",
      "password": "$DOCKER_PASS"
    }
  }
}
EOF

/kaniko/executor \
  --dockerfile=Dockerfile \
  --context=/home/jenkins/agent/workspace/unit-cost-api \
  --destination=${IMAGE}:${TAG}
'''
          }
        }
      }
    }

    stage('Deploy ClickHouse (Test)') {
      steps {
        container('helm') {
          sh '''
helm upgrade --install clickhouse-test bitnami/clickhouse
'''
        }
      }
    }

    stage('Deploy API (Test)') {
      steps {
        container('helm') {
          sh '''
helm upgrade --install api-test helm/api-chart \
  --set image.repository=${IMAGE} \
  --set image.tag=${TAG}
'''
        }
      }
    }

    stage('Wait for API') {
      steps {
        container('golang') {
          sh '''
sleep 20
'''
        }
      }
    }

    stage('Integration Test') {
      steps {
        container('golang') {
          sh '''
curl http://api-test:7000/api/v1/health
curl http://api-test:7000/metrics
'''
        }
      }
    }

    stage('Deploy Prod') {
      steps {
        container('helm') {
          sh '''
helm upgrade --install api helm/api-chart \
  --set image.repository=${IMAGE} \
  --set image.tag=${TAG}
'''
        }
      }
    }
  }

  post {
    always {
      container('helm') {
        sh '''
helm uninstall clickhouse-test || true
helm uninstall api-test || true
'''
      }
    }
  }
}