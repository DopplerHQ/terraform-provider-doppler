resource "aws_iam_role" "doppler_parameter_store" {
  name = "doppler_parameter_store"

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

            "ssm:PutParameter",
            "ssm:LabelParameterVersion",
            "ssm:DeleteParameter",
            "ssm:RemoveTagsFromResource",
            "ssm:GetParameterHistory",
            "ssm:AddTagsToResource",
            "ssm:GetParametersByPath",
            "ssm:GetParameters",
            "ssm:GetParameter",
            "ssm:DeleteParameters"
          ]
          Effect   = "Allow"
          Resource = "*"
          # Limit Doppler to only access certain names
        },
      ]
    })
  }
}


resource "doppler_integration_aws_parameter_store" "prod" {
  name            = "Production"
  assume_role_arn = aws_iam_role.doppler_parameter_store.arn
}

resource "doppler_secrets_sync_aws_parameter_store" "backend_prod" {
  integration = doppler_integration_aws_parameter_store.prod.id
  project     = "backend"
  config      = "prd"

  region        = "us-east-1"
  path          = "/backend/"
  secure_string = true
  tags          = { myTag = "enabled" }

  delete_behavior = "leave_in_target"
}

