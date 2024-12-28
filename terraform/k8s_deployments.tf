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
      target_port = 5678
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
          name  = "auth"
          image = "hashicorp/http-echo:0.2.3"
          args = [
            "-text=Hello from Auth Dummy Server!"
          ]

          port {
            container_port = 5678
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
          name  = "user"
          image = "hashicorp/http-echo:0.2.3"
          args = [
            "-text=Hello from User Dummy Server!"
          ]

          port {
            container_port = 5678
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

