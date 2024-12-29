provider "aws" {
  region  = "us-east-2"
  profile = "default"
}

module "vpc" {
  source                  = "terraform-aws-modules/vpc/aws"
  name                    = "eks-vpc"
  cidr                    = "10.0.0.0/16"
  azs                     = ["us-east-2a", "us-east-2b", "us-east-2c"]
  enable_dns_hostnames    = true
  enable_dns_support      = true
  public_subnets          = ["10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"]
  map_public_ip_on_launch = true
  private_subnets         = ["10.0.11.0/24", "10.0.12.0/24", "10.0.13.0/24"]
  enable_nat_gateway      = true
  single_nat_gateway      = true
  public_subnet_tags = {
    "kubernetes.io/cluster/cluster" = "shared"
    "kubernetes.io/role/elb"        = "1"
  }
  private_subnet_tags = {
    "kubernetes.io/role/internal-elb" = 1
    "kubernetes.io/cluster/cluster"   = "shared"
  }

}

