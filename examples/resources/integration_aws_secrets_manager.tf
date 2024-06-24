resource "aws_iam_role" "doppler_secrets_manager" {
  name = "doppler_secrets_manager"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = "sts:AssumeRole"
        Principal = {
          AWS = "arn:aws:iam::299900769157:user/doppler-integration-operator"
        },
        Condition = {
          StringEquals = {
            "sts:ExternalId" = "<YOUR_WORKPLACE_SLUG>"
          }
        }
      },
    ]
  })

  inline_policy {
    name = "doppler_secret_manager"

    policy = jsonencode({
      Version = "2012-10-17"
      Statement = [
        {
          Action = [

            "secretsmanager:GetSecretValue",
            "secretsmanager:DescribeSecret",
            "secretsmanager:PutSecretValue",
            "secretsmanager:CreateSecret",
            "secretsmanager:DeleteSecret",
            "secretsmanager:TagResource",
            "secretsmanager:UpdateSecret"
          ]
          Effect   = "Allow"
          Resource = "*"
          # Limit Doppler to only access certain secret names
        },
      ]
    })
  }
}

resource "doppler_integration_aws_secrets_manager" "prod" {
  name            = "Production"
  assume_role_arn = aws_iam_role.doppler_secrets_manager.arn
}

resource "doppler_secrets_sync_aws_secrets_manager" "backend_prod" {
  integration = doppler_integration_aws_secrets_manager.prod.id
  project     = "backend"
  config      = "prd"

  region = "us-east-1"
  path   = "/backend/"
  tags   = { myTag = "enabled" }

  delete_behavior = "leave_in_target"
}

