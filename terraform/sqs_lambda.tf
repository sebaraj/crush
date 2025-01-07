resource "aws_sqs_queue" "match_queue" {
  name = "match-queue"
}

data "aws_iam_policy_document" "lambda_assume_role_policy" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "lambda_execution_role" {
  name               = "lambda-execution-role"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume_role_policy.json
}

data "aws_iam_policy_document" "lambda_sqs_cw_policy" {
  statement {
    effect = "Allow"
    actions = [
      "sqs:ReceiveMessage",
      "sqs:DeleteMessage",
      "sqs:GetQueueAttributes",
      "sqs:GetQueueUrl"
    ]
    resources = [
      aws_sqs_queue.match_queue.arn
    ]
  }

  statement {
    effect = "Allow"
    actions = [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutLogEvents"
    ]
    resources = ["*"]
  }
}

resource "aws_iam_policy" "lambda_sqs_cw_policy" {
  name   = "lambda-sqs-cw-policy"
  policy = data.aws_iam_policy_document.lambda_sqs_cw_policy.json
}

resource "aws_iam_role_policy_attachment" "lambda_execution_role_sqs_cw_attach" {
  role       = aws_iam_role.lambda_execution_role.name
  policy_arn = aws_iam_policy.lambda_sqs_cw_policy.arn
}

resource "aws_iam_role_policy_attachment" "lambda_vpc_access" {
  role       = aws_iam_role.lambda_execution_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaVPCAccessExecutionRole"
}

resource "aws_security_group" "lambda_sg" {
  name   = "lambda-sg"
  vpc_id = module.vpc.vpc_id

  # Egress allows all outbound
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_security_group_rule" "lambda_to_rds" {
  type                     = "ingress"
  from_port                = 5432
  to_port                  = 5432
  protocol                 = "tcp"
  security_group_id        = aws_security_group.rds_sg.id
  source_security_group_id = aws_security_group.lambda_sg.id
}

resource "aws_lambda_function" "sqs_consumer" {
  function_name = "match-sqs-consumer"
  role          = aws_iam_role.lambda_execution_role.arn
  handler       = "handler"
  runtime       = "go1.23"

  filename         = "${path.module}/lambda_code/lambda.zip"
  source_code_hash = filebase64sha256("${path.module}/lambda_code/lambda.zip")

  vpc_config {
    security_group_ids = [aws_security_group.lambda_sg.id]
    subnet_ids         = module.vpc.private_subnets
  }

  environment {
    variables = {
      DB_ENDPOINT = aws_db_instance.postgres_db.address
      DB_NAME     = "mydb"
      DB_PORT     = "5432"
      DB_USERNAME = var.db_username
      DB_PASSWORD = var.db_password
    }
  }
  timeout = 60
}

resource "aws_lambda_event_source_mapping" "sqs_trigger" {
  event_source_arn = aws_sqs_queue.match_queue.arn
  function_name    = aws_lambda_function.sqs_consumer.arn
  enabled          = true
  batch_size       = 10
}


