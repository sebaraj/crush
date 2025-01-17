resource "aws_db_parameter_group" "postgres_params" {
  name        = "rds-postgres-params"
  family      = "postgres15"
  description = "Parameter group for logical replication"

  parameter {
    apply_method = "pending-reboot"
    name         = "max_wal_senders"
    value        = "10"
  }

  parameter {
    apply_method = "pending-reboot"
    name         = "max_replication_slots"
    value        = "10"
  }

  parameter {
    apply_method = "pending-reboot"
    name         = "rds.logical_replication"
    value        = "1"
  }
}


resource "aws_db_subnet_group" "rds_subnet_group" {
  name        = "rds-postgres-subnet-group"
  description = "Subnet group for RDS Postgres"
  subnet_ids  = module.vpc.private_subnets
}

resource "aws_security_group" "rds_sg" {
  name   = "rds-postgres-sg"
  vpc_id = module.vpc.vpc_id

  ingress {
    from_port       = 5432
    to_port         = 5432
    protocol        = "tcp"
    security_groups = [module.eks.node_security_group_id]
    description     = "Allow Postgres access from EKS worker nodes"
  }

  ingress {
    from_port       = 5432
    to_port         = 5432
    protocol        = "tcp"
    security_groups = [aws_security_group.dms_sg.id]
    description     = "Allow Postgres access from DMS"
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

variable "db_username" {
  description = "username for psql database"
  type        = string
  sensitive   = true
}

variable "db_password" {
  description = "password for psql database"
  type        = string
  sensitive   = true
}

resource "aws_db_instance" "postgres_db" {
  identifier             = "my-postgres-db"
  db_name                = "mydb"
  engine                 = "postgres"
  engine_version         = "15"
  instance_class         = "db.t3.micro"
  allocated_storage      = 20
  storage_encrypted      = true
  username               = var.db_username
  password               = var.db_password
  db_subnet_group_name   = aws_db_subnet_group.rds_subnet_group.name
  vpc_security_group_ids = [aws_security_group.rds_sg.id]
  publicly_accessible    = false
  parameter_group_name   = aws_db_parameter_group.postgres_params.name

  backup_window = "03:00-04:00"
}

resource "kubernetes_secret" "rds_credentials" {
  metadata {
    name      = "rds-credentials"
    namespace = "default"
  }

  data = {
    db_username = var.db_username
    db_password = var.db_password
    db_endpoint = aws_db_instance.postgres_db.endpoint
    db_name     = "mydb"
    db_port     = "5432"
  }

  type = "Opaque"
}

