module "lb_role" {
  source = "terraform-aws-modules/iam/aws//modules/iam-role-for-service-accounts-eks"

  role_name                              = "eks-alb-ingress"
  attach_load_balancer_controller_policy = true

  oidc_providers = {
    main = {
      provider_arn               = module.eks.oidc_provider_arn
      namespace_service_accounts = ["kube-system:aws-load-balancer-controller"]
    }
  }
}

provider "helm" {
  kubernetes {
    host                   = data.aws_eks_cluster.cluster.endpoint
    token                  = data.aws_eks_cluster_auth.cluster.token
    cluster_ca_certificate = base64decode(data.aws_eks_cluster.cluster.certificate_authority.0.data)
  }
}

resource "helm_release" "alb_ingress_controller" {
  depends_on = [
    module.eks,
    data.aws_eks_cluster.cluster,
    data.aws_eks_cluster_auth.cluster,
  ]
  name       = "aws-alb-ingress-controller"
  namespace  = "kube-system"
  repository = "https://aws.github.io/eks-charts"
  chart      = "aws-load-balancer-controller"

  set {
    name  = "clusterName"
    value = module.eks.cluster_name
  }

  set {
    name  = "serviceAccount.create"
    value = "true"
  }

  set {
    name  = "serviceAccount.name"
    value = "aws-load-balancer-controller"
  }

  set {
    name  = "serviceAccount.annotations.eks\\.amazonaws\\.com/role-arn"
    value = module.lb_role.iam_role_arn
  }
  set {
    name  = "region"
    value = "us-east-2"
  }

  timeout = 300
}

resource "kubernetes_ingress_v1" "frontend_ingress" {
  metadata {
    name      = "frontend-ingress"
    namespace = "default"
    annotations = {
      "kubernetes.io/ingress.class"                       = "alb"
      "alb.ingress.kubernetes.io/scheme"                  = "internet-facing"
      "alb.ingress.kubernetes.io/listen-ports"            = "[{\"HTTPS\":443}]"
      "alb.ingress.kubernetes.io/certificate-arn"         = aws_acm_certificate_validation.certificate_validation.certificate_arn
      "alb.ingress.kubernetes.io/ssl-redirect"            = "443"
      "alb.ingress.kubernetes.io/backend-protocol"        = "HTTP"
      "alb.ingress.kubernetes.io/target-type"             = "instance"
      "alb.ingress.kubernetes.io/target-group-attributes" = "deregistration_delay.timeout_seconds=30,stickiness.enabled=true,stickiness.lb_cookie.duration_seconds=300"
      "alb.ingress.kubernetes.io/healthcheck-path"        = "/health"
      "alb.ingress.kubernetes.io/healthcheck-port"        = "traffic-port"
    }
  }

  spec {
    rule {
      host = "api.yalecrush.com"

      http {
        path {
          path      = "/auth"
          path_type = "Prefix"

          backend {
            service {
              name = "auth-service"
              port {
                number = 80
              }
            }
          }
        }

        path {
          path      = "/v1/user"
          path_type = "Prefix"

          backend {
            service {
              name = "user-service"
              port {
                number = 80
              }
            }
          }
        }

        path {
          path      = "/v1/match"
          path_type = "Prefix"

          backend {
            service {
              name = "match-service"
              port {
                number = 80
              }
            }
          }
        }
      }
    }
  }
}




