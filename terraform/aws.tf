provider "aws" {
  region  = "us-east-2"
  profile = "default"
}

module "vpc" {
  source                  = "terraform-aws-modules/vpc/aws"
  name                    = "eks-vpc"
  cidr                    = "10.0.0.0/16"
  azs                     = ["us-east-2a", "us-east-2b", "us-east-2c"]
  public_subnets          = ["10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"]
  map_public_ip_on_launch = true
  public_subnet_tags = {
    "kubernetes.io/cluster/cluster" = "shared"
    "kubernetes.io/role/elb"        = "1"
  }
  private_subnet_tags = {
    "kubernetes.io/role/internal-elb" = 1
    "kubernetes.io/cluster/cluster"   = "shared"
  }

}

