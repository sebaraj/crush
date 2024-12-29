module "eks" {
  source                         = "terraform-aws-modules/eks/aws"
  cluster_name                   = "cluster"
  cluster_version                = "1.30"
  vpc_id                         = module.vpc.vpc_id
  subnet_ids                     = concat(module.vpc.public_subnets, module.vpc.private_subnets)
  cluster_endpoint_public_access = true

  eks_managed_node_groups = {
    nodes = {
      desired_size   = 1
      max_size       = 3
      min_size       = 1
      instance_types = ["t2.medium"]
    }

  }
}

data "aws_eks_cluster" "cluster" {
  name       = module.eks.cluster_name
  depends_on = [module.eks]
}

data "aws_eks_cluster_auth" "cluster" {
  name       = module.eks.cluster_name
  depends_on = [module.eks]
}

provider "kubernetes" {
  host                   = data.aws_eks_cluster.cluster.endpoint
  token                  = data.aws_eks_cluster_auth.cluster.token
  cluster_ca_certificate = base64decode(data.aws_eks_cluster.cluster.certificate_authority[0].data)
  config_path            = "~/.kube/config"

  exec {
    api_version = "client.authentication.k8s.io/v1beta1"
    command     = "aws"
    args = [
      "eks",
      "get-token",
      "--cluster-name",
      module.eks.cluster_name
    ]
  }

}

resource "kubernetes_cluster_role_binding" "admin" {
  metadata {
    name = "cluster-admin-binding"
  }

  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "ClusterRole"
    name      = "cluster-admin"
  }

  subject {
    kind      = "User"
    name      = "system:master"
    api_group = "rbac.authorization.k8s.io"
  }


}



