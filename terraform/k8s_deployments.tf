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

    type = "ClusterIP"
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

    type = "ClusterIP"
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

    type = "ClusterIP"
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
        container {
          name  = "auth"
          image = "hashicorp/http-echo:0.2.3"
          args = [
            "-text=Hello from Auth Dummy Server!"
          ]

          port {
            container_port = 5678
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
        container {
          name  = "user"
          image = "hashicorp/http-echo:0.2.3"
          args = [
            "-text=Hello from User Dummy Server!"
          ]

          port {
            container_port = 5678
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
        container {
          name  = "match"
          image = "hashicorp/http-echo:0.2.3"
          args = [
            "-text=Hello from Match Dummy Server!"
          ]

          port {
            container_port = 5678
          }
        }
      }
    }
  }
}

