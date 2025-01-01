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

# resource "aws_security_group_rule" "allow_dms_to_rds" {
#   type                     = "ingress"
#   from_port                = 5432
#   to_port                  = 5432
#   protocol                 = "tcp"
#   security_group_id        = aws_security_group.rds_sg.id
#   source_security_group_id = aws_security_group.dms_sg.id # DMS Security Group
#   description              = "Allow DMS to connect to RDS"
# }
#
#
resource "aws_security_group_rule" "dms_to_rds" {
  type                     = "egress"
  from_port                = 5432
  to_port                  = 5432
  protocol                 = "tcp"
  source_security_group_id = aws_security_group.rds_sg.id
  security_group_id        = aws_security_group.dms_sg.id
}

# ingress?
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
  ssl_mode      = "require"
}

resource "aws_dms_endpoint" "opensearch_endpoint" {
  endpoint_id         = "opensearch-endpoint"
  endpoint_type       = "target"
  engine_name         = "opensearch"
  service_access_role = aws_iam_role.dms_access_role.arn
  elasticsearch_settings {
    endpoint_uri            = "https://${aws_opensearch_domain.opensearch.endpoint}"
    service_access_role_arn = aws_iam_role.dms_access_role.arn
    use_new_mapping_type    = true
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
          "ec2:*",
          "opensearch:*",
          "es:ESHttpPost",
          "es:ESHttpPut",
          "es:ESHttpGet",
          "es:ESHttpHead",
          "es:ESHttpDelete",
        ]
        Resource = [
          "${aws_opensearch_domain.opensearch.arn}",
          "${aws_opensearch_domain.opensearch.arn}/*",
          "*"
        ]
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
  migration_type           = "full-load-and-cdc"
  replication_task_settings = jsonencode({
    TargetMetadata = {
      TargetSchema       = ""
      SupportLobs        = false
      FullLobMode        = false
      LobChunkSize       = 64
      LimitedSizeLobMode = true
      LobMaxSize         = 32
    }
    FullLoadSettings = {
      TargetTablePrepMode             = "DO_NOTHING"
      CreatePkAfterFullLoad           = false
      StopTaskCachedChangesApplied    = false
      StopTaskCachedChangesNotApplied = false
      MaxFullLoadSubTasks             = 8
      TransactionConsistencyTimeout   = 600
      CommitRate                      = 10000
    }
    Logging = {
      EnableLogging = true
      LogComponents = [
        {
          Id       = "TRANSFORMATION"
          Severity = "LOGGER_SEVERITY_DEFAULT"
        },
        {
          Id       = "SOURCE_UNLOAD"
          Severity = "LOGGER_SEVERITY_DEFAULT"
        },
        {
          Id       = "IO"
          Severity = "LOGGER_SEVERITY_DEFAULT"
        },
        {
          Id       = "TARGET_LOAD"
          Severity = "LOGGER_SEVERITY_DEFAULT"
        },
        {
          Id       = "PERFORMANCE"
          Severity = "LOGGER_SEVERITY_DEFAULT"
        }
      ]
    }
    ControlTablesSettings = {
      ControlSchema               = ""
      HistoryTimeslotInMinutes    = 5
      HistoryTableEnabled         = false
      SuspendedTablesTableEnabled = false
      StatusTableEnabled          = false
    }
    StreamBufferSettings = {
      StreamBufferCount        = 3
      StreamBufferSizeInMB     = 8
      CtrlStreamBufferSizeInMB = 5
    }
    ChangeProcessingDdlHandlingPolicy = {
      HandleSourceTableDropped   = true
      HandleSourceTableTruncated = true
      HandleSourceTableAltered   = true
    }
    ErrorBehavior = {
      DataErrorPolicy               = "LOG_ERROR"
      DataTruncationErrorPolicy     = "LOG_ERROR"
      DataErrorEscalationPolicy     = "SUSPEND_TABLE"
      DataErrorEscalationCount      = 0
      TableErrorPolicy              = "SUSPEND_TABLE"
      TableErrorEscalationPolicy    = "STOP_TASK"
      TableErrorEscalationCount     = 0
      RecoverableErrorCount         = -1
      RecoverableErrorInterval      = 5
      RecoverableErrorThrottling    = true
      RecoverableErrorThrottlingMax = 1800
      ApplyErrorDeletePolicy        = "IGNORE_RECORD"
      ApplyErrorInsertPolicy        = "LOG_ERROR"
      ApplyErrorUpdatePolicy        = "LOG_ERROR"
      ApplyErrorEscalationPolicy    = "LOG_ERROR"
      ApplyErrorEscalationCount     = 0
      FullLoadIgnoreConflicts       = true
    }
    ChangeProcessingTuning = {
      BatchApplyPreserveTransaction = true
      BatchApplyTimeoutMin          = 1
      BatchApplyTimeoutMax          = 30
      BatchApplyMemoryLimit         = 500
      BatchSplitSize                = 0
      MinTransactionSize            = 1000
      CommitTimeout                 = 1
      MemoryLimitTotal              = 1024
      MemoryKeepTime                = 60
      StatementCacheSize            = 50
    }
  })
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
        "rule-type" = "selection"
        "rule-id"   = "2"
        "rule-name" = "include-test-table"
        "object-locator" = {
          "schema-name" = "public"
          "table-name"  = "test"
        }
        "rule-action" = "include"
      },
      {
        "rule-type"   = "transformation"
        "rule-id"     = "4"
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
        "rule-id"     = "5"
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
        "rule-id"     = "6"
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
        "rule-id"     = "7"
        "rule-name"   = "exclude-partner-genders"
        "rule-action" = "remove-column"
        "rule-target" = "column"
        "object-locator" = {
          "schema-name" = "public"
          "table-name"  = "users"
          "column-name" = "partner_genders"
        }
      },
      {
        "rule-type"   = "transformation"
        "rule-id"     = "8"
        "rule-name"   = "exclude-instagram"
        "rule-action" = "remove-column"
        "rule-target" = "column"
        "object-locator" = {
          "schema-name" = "public"
          "table-name"  = "users"
          "column-name" = "instagram"
        }
      },
      {
        "rule-type"   = "transformation"
        "rule-id"     = "9"
        "rule-name"   = "exclude-snapchat"
        "rule-action" = "remove-column"
        "rule-target" = "column"
        "object-locator" = {
          "schema-name" = "public"
          "table-name"  = "users"
          "column-name" = "snapchat"
        }
      },
      {
        "rule-type"   = "transformation"
        "rule-id"     = "10"
        "rule-name"   = "exclude-phone"
        "rule-action" = "remove-column"
        "rule-target" = "column"
        "object-locator" = {
          "schema-name" = "public"
          "table-name"  = "users"
          "column-name" = "phone_number"
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



