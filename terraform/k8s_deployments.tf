variable "oauth_client" {
  description = "OAuth client id for the application"
  type        = string
  sensitive   = true
}


resource "kubernetes_secret" "oauth_credentials" {
  metadata {
    name      = "oauth-credentials"
    namespace = "default"
  }

  data = {
    oauth_client = var.oauth_client
  }

  type = "Opaque"
}


data "aws_ecr_repository" "auth_serv_repo" {
  name = "yalecrush/auth"
}

data "aws_ecr_repository" "match_serv_repo" {
  name = "yalecrush/match"
}

data "aws_ecr_repository" "user_serv_repo" {
  name = "yalecrush/user"
}

resource "kubernetes_service" "auth_service" {
  metadata {
    name      = "auth-service"
    namespace = "default"
  }

  spec {
    selector = {
      app = "auth"
    }

    port {
      port        = 80
      target_port = 5678
    }

    type = "NodePort"
  }
}

resource "kubernetes_service" "user_service" {
  metadata {
    name      = "user-service"
    namespace = "default"
  }

  spec {
    selector = {
      app = "user"
    }

    port {
      port        = 80
      target_port = 6000
    }

    type = "NodePort"
  }
}

resource "kubernetes_service" "match_service" {
  metadata {
    name      = "match-service"
    namespace = "default"
  }

  spec {
    selector = {
      app = "match"
    }

    port {
      port        = 80
      target_port = 5678
    }

    type = "NodePort"
  }
}

resource "kubernetes_deployment" "auth_deployment" {
  metadata {
    name = "auth"
    labels = {
      app = "auth"
    }
  }

  spec {
    replicas = 1

    selector {
      match_labels = {
        app = "auth"
      }
    }

    template {
      metadata {
        labels = {
          app = "auth"
        }
      }

      spec {
        affinity {
          pod_anti_affinity {
            required_during_scheduling_ignored_during_execution {
              label_selector {
                match_expressions {
                  key      = "app"
                  operator = "In"
                  values   = ["auth"]
                }
              }
              topology_key = "kubernetes.io/hostname"
            }
          }
        }
        container {
          name              = "auth"
          image             = "${data.aws_ecr_repository.auth_serv_repo.repository_url}:latest"
          image_pull_policy = "Always"

          port {
            container_port = 5678
          }
          env {
            name = "OAUTH_CLIENT"
            value_from {
              secret_key_ref {
                name = kubernetes_secret.oauth_credentials.metadata[0].name
                key  = "oauth_client"
              }
            }
          }

          env {
            name = "DB_USERNAME"
            value_from {
              secret_key_ref {
                name = kubernetes_secret.rds_credentials.metadata[0].name
                key  = "db_username"
              }
            }
          }
          env {
            name = "DB_PASSWORD"
            value_from {
              secret_key_ref {
                name = kubernetes_secret.rds_credentials.metadata[0].name
                key  = "db_password"
              }
            }
          }
          env {
            name = "DB_ENDPOINT"
            value_from {
              secret_key_ref {
                name = kubernetes_secret.rds_credentials.metadata[0].name
                key  = "db_endpoint"
              }
            }
          }
          env {
            name = "DB_NAME"
            value_from {
              secret_key_ref {
                name = kubernetes_secret.rds_credentials.metadata[0].name
                key  = "db_name"
              }
            }
          }
          env {
            name = "DB_PORT"
            value_from {
              secret_key_ref {
                name = kubernetes_secret.rds_credentials.metadata[0].name
                key  = "db_port"
              }
            }
          }
        }

      }
    }
  }
}

resource "kubernetes_deployment" "user_deployment" {
  metadata {
    name = "user"
    labels = {
      app = "user"
    }
  }

  spec {
    replicas = 1

    selector {
      match_labels = {
        app = "user"
      }
    }

    template {
      metadata {
        labels = {
          app = "user"
        }
      }

      spec {
        service_account_name = kubernetes_service_account.presign_service_account.metadata[0].name

        affinity {
          pod_anti_affinity {
            required_during_scheduling_ignored_during_execution {
              label_selector {
                match_expressions {
                  key      = "app"
                  operator = "In"
                  values   = ["user"]
                }
              }
              topology_key = "kubernetes.io/hostname"
            }
          }
        }
        container {
          name              = "user"
          image             = "${data.aws_ecr_repository.user_serv_repo.repository_url}:latest"
          image_pull_policy = "Always"

          port {
            container_port = 6000
          }
          env {
            name  = "S3_REGION"
            value = "us-east-2"
          }
          env {
            name  = "S3_BUCKET"
            value = aws_s3_bucket.images_bucket.bucket
          }
          env {
            name = "OPENSEARCH_ENDPOINT"
            value_from {
              secret_key_ref {
                name = kubernetes_secret.opensearch_credentials.metadata[0].name
                key  = "opensearch_endpoint"
              }
            }
          }
          env {
            name = "OAUTH_CLIENT"
            value_from {
              secret_key_ref {
                name = kubernetes_secret.oauth_credentials.metadata[0].name
                key  = "oauth_client"
              }
            }
          }
          env {
            name = "DB_USERNAME"
            value_from {
              secret_key_ref {
                name = kubernetes_secret.rds_credentials.metadata[0].name
                key  = "db_username"
              }
            }
          }
          env {
            name = "DB_PASSWORD"
            value_from {
              secret_key_ref {
                name = kubernetes_secret.rds_credentials.metadata[0].name
                key  = "db_password"
              }
            }
          }
          env {
            name = "DB_ENDPOINT"
            value_from {
              secret_key_ref {
                name = kubernetes_secret.rds_credentials.metadata[0].name
                key  = "db_endpoint"
              }
            }
          }
          env {
            name = "DB_NAME"
            value_from {
              secret_key_ref {
                name = kubernetes_secret.rds_credentials.metadata[0].name
                key  = "db_name"
              }
            }
          }
          env {
            name = "DB_PORT"
            value_from {
              secret_key_ref {
                name = kubernetes_secret.rds_credentials.metadata[0].name
                key  = "db_port"
              }
            }
          }

        }
      }
    }
  }
}

resource "kubernetes_deployment" "match_deployment" {
  metadata {
    name = "match"
    labels = {
      app = "match"
    }
  }

  spec {
    replicas = 1

    selector {
      match_labels = {
        app = "match"
      }
    }

    template {
      metadata {
        labels = {
          app = "match"
        }
      }

      spec {
        affinity {
          pod_anti_affinity {
            required_during_scheduling_ignored_during_execution {
              label_selector {
                match_expressions {
                  key      = "app"
                  operator = "In"
                  values   = ["match"]
                }
              }
              topology_key = "kubernetes.io/hostname"
            }
          }
        }
        container {
          name  = "match"
          image = "hashicorp/http-echo:0.2.3"
          args = [
            "-text=Hello from Match Dummy Server!"
          ]

          port {
            container_port = 5678
          }
          env {
            name  = "MATCH_QUEUE_URL"
            value = aws_sqs_queue.match_queue.url
          }
          env {
            name = "OPENSEARCH_ENDPOINT"
            value_from {
              secret_key_ref {
                name = kubernetes_secret.opensearch_credentials.metadata[0].name
                key  = "opensearch_endpoint"
              }
            }
          }
          env {
            name = "OAUTH_CLIENT"
            value_from {
              secret_key_ref {
                name = kubernetes_secret.oauth_credentials.metadata[0].name
                key  = "oauth_client"
              }
            }
          }
          env {
            name = "DB_USERNAME"
            value_from {
              secret_key_ref {
                name = kubernetes_secret.rds_credentials.metadata[0].name
                key  = "db_username"
              }
            }
          }
          env {
            name = "DB_PASSWORD"
            value_from {
              secret_key_ref {
                name = kubernetes_secret.rds_credentials.metadata[0].name
                key  = "db_password"
              }
            }
          }
          env {
            name = "DB_ENDPOINT"
            value_from {
              secret_key_ref {
                name = kubernetes_secret.rds_credentials.metadata[0].name
                key  = "db_endpoint"
              }
            }
          }
          env {
            name = "DB_NAME"
            value_from {
              secret_key_ref {
                name = kubernetes_secret.rds_credentials.metadata[0].name
                key  = "db_name"
              }
            }
          }
          env {
            name = "DB_PORT"
            value_from {
              secret_key_ref {
                name = kubernetes_secret.rds_credentials.metadata[0].name
                key  = "db_port"
              }
            }
          }

        }
      }
    }
  }
}

