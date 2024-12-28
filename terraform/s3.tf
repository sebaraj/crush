resource "aws_s3_bucket" "images_bucket" {
  bucket = "yale-crush-user-images"
}

resource "aws_s3_bucket_cors_configuration" "images_bucket_cors" {
  bucket = aws_s3_bucket.images_bucket.id

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["GET", "PUT", "POST", "DELETE"]
    allowed_origins = ["https://yalecrush.com"]
    max_age_seconds = 3000
  }
}

resource "aws_s3_bucket_public_access_block" "public_access_block" {
  bucket = aws_s3_bucket.images_bucket.id

  block_public_acls       = false
  block_public_policy     = false
  ignore_public_acls      = false
  restrict_public_buckets = false
}

data "aws_iam_policy_document" "images_bucket_policy" {
  statement {
    sid    = "PublicReadGetObject"
    effect = "Allow"

    principals {
      type        = "AWS"
      identifiers = ["*"]
    }

    actions   = ["s3:GetObject"]
    resources = ["${aws_s3_bucket.images_bucket.arn}/*"]
  }

  statement {
    sid    = "AllowPutObjectForPresignRole"
    effect = "Allow"

    principals {
      type        = "AWS"
      identifiers = [module.presign_role.iam_role_arn]
    }

    actions   = ["s3:PutObject"]
    resources = ["${aws_s3_bucket.images_bucket.arn}/*"]
  }
}

resource "aws_s3_bucket_policy" "images_bucket_policy" {
  bucket = aws_s3_bucket.images_bucket.id

  policy = data.aws_iam_policy_document.images_bucket_policy.json
}

data "aws_iam_policy_document" "s3_presign_policy" {
  statement {
    actions = [
      "s3:PutObject",
      "s3:GetObject",
    ]
    resources = [
      "${aws_s3_bucket.images_bucket.arn}/*"
    ]
  }
}

resource "aws_iam_policy" "s3_presign_policy" {
  name   = "s3-presign-policy"
  policy = data.aws_iam_policy_document.s3_presign_policy.json
}

module "presign_role" {
  source    = "terraform-aws-modules/iam/aws//modules/iam-role-for-service-accounts-eks"
  role_name = "s3-presign-role"
  role_policy_arns = {
    policy = aws_iam_policy.s3_presign_policy.arn
  }
  oidc_providers = {
    main = {
      provider_arn               = module.eks.oidc_provider_arn
      namespace_service_accounts = ["default:presign-serviceaccount"]
    }
  }
}

resource "kubernetes_service_account" "presign_service_account" {
  metadata {
    name      = "presign-serviceaccount"
    namespace = "default"
    annotations = {
      "eks.amazonaws.com/role-arn" = module.presign_role.iam_role_arn
    }
  }
}


