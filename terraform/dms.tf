resource "aws_dms_replication_subnet_group" "dms_subnet_group" {
  replication_subnet_group_id          = "dms-subnet-group"
  replication_subnet_group_description = "DMS subnet group"
  subnet_ids                           = module.vpc.private_subnets
}

resource "aws_security_group" "dms_sg" {
  name   = "dms-sg"
  vpc_id = module.vpc.vpc_id

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_security_group_rule" "dms_to_rds" {
  type                     = "egress"
  from_port                = 5432
  to_port                  = 5432
  protocol                 = "tcp"
  source_security_group_id = aws_security_group.rds_sg.id
  security_group_id        = aws_security_group.dms_sg.id
}

resource "aws_security_group_rule" "dms_to_opensearch" {
  type                     = "egress"
  from_port                = 443
  to_port                  = 443
  protocol                 = "tcp"
  source_security_group_id = aws_security_group.opensearch_sg.id
  security_group_id        = aws_security_group.dms_sg.id
}

resource "aws_dms_endpoint" "rds_endpoint" {
  endpoint_id   = "rds-endpoint"
  endpoint_type = "source"
  engine_name   = "postgres"
  username      = var.db_username
  password      = var.db_password
  server_name   = aws_db_instance.postgres_db.address
  database_name = aws_db_instance.postgres_db.db_name
  port          = 5432
}

resource "aws_dms_endpoint" "opensearch_endpoint" {
  endpoint_id         = "opensearch-endpoint"
  endpoint_type       = "target"
  engine_name         = "elasticsearch"
  service_access_role = aws_iam_role.dms_access_role.arn
  elasticsearch_settings {
    endpoint_uri            = "https://${aws_opensearch_domain.opensearch.endpoint}"
    service_access_role_arn = aws_iam_role.dms_access_role.arn
  }
}

resource "aws_iam_role" "dms_access_role" {
  name = "dms-access-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Principal = {
        Service = "dms.amazonaws.com"
      }
      Action = "sts:AssumeRole"
    }]
  })
}

resource "aws_iam_policy" "dms_access_policy" {
  name        = "dms-access-policy"
  description = "Policy for DMS to access RDS and OpenSearch"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "rds:*",
          "es:*",
          "s3:*",
          "ec2:*"
        ]
        Resource = "*"
      }
    ]
  })
}

resource "aws_iam_role" "dms_vpc_role" {
  name = "dms-vpc-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Effect = "Allow",
        Principal = {
          Service = "dms.amazonaws.com"
        },
        Action = "sts:AssumeRole"
      }
    ]
  })
}

resource "aws_iam_policy" "dms_vpc_policy" {
  name        = "dms-vpc-policy"
  description = "Policy for DMS to access VPC components"

  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Effect = "Allow",
        Action = [
          "ec2:Describe*",
          "ec2:CreateNetworkInterface",
          "ec2:DeleteNetworkInterface",
          "ec2:AttachNetworkInterface",
          "ec2:DetachNetworkInterface"
        ],
        Resource = "*"
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "dms_vpc_role_attachment" {
  role       = aws_iam_role.dms_vpc_role.name
  policy_arn = aws_iam_policy.dms_vpc_policy.arn
}

resource "aws_iam_role_policy_attachment" "dms_access_role_attachment" {
  role       = aws_iam_role.dms_access_role.name
  policy_arn = aws_iam_policy.dms_access_policy.arn
}

resource "aws_dms_replication_task" "cdc_task" {
  replication_task_id      = "cdc-users-to-opensearch"
  replication_instance_arn = aws_dms_replication_instance.replication_instance.replication_instance_arn
  source_endpoint_arn      = aws_dms_endpoint.rds_endpoint.endpoint_arn
  target_endpoint_arn      = aws_dms_endpoint.opensearch_endpoint.endpoint_arn
  migration_type           = "cdc"
  table_mappings = jsonencode({
    rules = [
      {
        "rule-type" = "selection"
        "rule-id"   = "1"
        "rule-name" = "include-users-table"
        "object-locator" = {
          "schema-name" = "public"
          "table-name"  = "users"
        }
        "rule-action" = "include"
      },
      {
        "rule-type"   = "transformation"
        "rule-id"     = "2"
        "rule-name"   = "exclude-created-at"
        "rule-action" = "remove-column"
        "rule-target" = "column"
        "object-locator" = {
          "schema-name" = "public"
          "table-name"  = "users"
          "column-name" = "created_at"
        }
      },
      {
        "rule-type"   = "transformation"
        "rule-id"     = "3"
        "rule-name"   = "exclude-updated-at"
        "rule-action" = "remove-column"
        "rule-target" = "column"
        "object-locator" = {
          "schema-name" = "public"
          "table-name"  = "users"
          "column-name" = "updated_at"
        }
      },
      {
        "rule-type"   = "transformation"
        "rule-id"     = "4"
        "rule-name"   = "exclude-gender"
        "rule-action" = "remove-column"
        "rule-target" = "column"
        "object-locator" = {
          "schema-name" = "public"
          "table-name"  = "users"
          "column-name" = "gender"
        }
      },
      {
        "rule-type"   = "transformation"
        "rule-id"     = "5"
        "rule-name"   = "exclude-partner-genders"
        "rule-action" = "remove-column"
        "rule-target" = "column"
        "object-locator" = {
          "schema-name" = "public"
          "table-name"  = "users"
          "column-name" = "partner_genders"
        }
      }
    ]
  })
  cdc_start_position = "4AF/B00000D0"
}

resource "aws_dms_replication_instance" "replication_instance" {
  replication_instance_id     = "dms-replication-instance"
  replication_instance_class  = "dms.t2.micro"
  allocated_storage           = 5
  apply_immediately           = true
  availability_zone           = "us-east-2a"
  multi_az                    = false
  vpc_security_group_ids      = [aws_security_group.dms_sg.id]
  replication_subnet_group_id = aws_dms_replication_subnet_group.dms_subnet_group.id
}



