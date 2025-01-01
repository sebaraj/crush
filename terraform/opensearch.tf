resource "aws_opensearch_domain" "opensearch" {
  domain_name    = "my-opensearch-domain"
  engine_version = "OpenSearch_2.17"
  vpc_options {
    subnet_ids         = [element(module.vpc.private_subnets, 0)]
    security_group_ids = [aws_security_group.opensearch_sg.id]
  }

  node_to_node_encryption {
    enabled = true
  }

  encrypt_at_rest {
    enabled = true
  }


  ebs_options {
    ebs_enabled = true
    volume_size = 10
    volume_type = "gp2"
  }

  cluster_config {
    instance_type            = "t3.small.search"
    instance_count           = 2
    dedicated_master_enabled = true
    dedicated_master_type    = "t3.small.search"
    dedicated_master_count   = 2
  }

  advanced_options = {
    "rest.action.multi.allow_explicit_index" = "true"
  }
}

# resource "aws_security_group_rule" "opensearch_to_dms" {
#   type                     = "ingress"
#   from_port                = 443
#   to_port                  = 443
#   protocol                 = "tcp"
#   security_group_id        = aws_security_group.opensearch_sg.id
#   source_security_group_id = aws_security_group.dms_sg.id
#   description              = "Allow OpenSearch to respond to DMS"
# }


resource "aws_security_group" "opensearch_sg" {
  name   = "opensearch-sg"
  vpc_id = module.vpc.vpc_id

  ingress {
    from_port       = 443
    to_port         = 443
    protocol        = "tcp"
    security_groups = [module.eks.node_security_group_id]
    description     = "Allow HTTPS access from EKS worker nodes"
  }

  ingress {
    from_port       = 443
    to_port         = 443
    protocol        = "tcp"
    security_groups = [aws_security_group.dms_sg.id]
    description     = "Allow HTTPS access from DMS"
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_opensearch_domain_policy" "main" {
  domain_name = aws_opensearch_domain.opensearch.domain_name

  access_policies = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect    = "Allow",
        Principal = "*",
        Action    = ["es:ESHttpGet", "es:ESHttpPut", "es:ESHttpPost", "es:ESHttpDelete"],
        Resource  = "${aws_opensearch_domain.opensearch.arn}/*"
      },
      {
        Effect = "Allow"
        Principal = {
          AWS = [aws_iam_role.dms_access_role.arn]
        }
        Action = [
          "opensearch:ESHttpGet",
          "opensearch:ESHttpPut",
          "opensearch:ESHttpPost",
          "opensearch:ESHttpDelete",
          "opensearch:ESHttpHead"
        ]
        Resource = "${aws_opensearch_domain.opensearch.arn}/*"
      }
    ]
  })
}

