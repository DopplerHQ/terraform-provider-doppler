# The initial password for the first db user. This will be rotated, so we provide a default and ignore changes below.
# Provide this with -var-file: https://developer.hashicorp.com/terraform/language/values/variables#variable-definitions-tfvars-files 
variable "db_password_1" {
  type = string
  default = ""
  # Consider using ephemeral instead if your client supports it: https://developer.hashicorp.com/terraform/language/values/variables#exclude-values-from-state
  sensitive = true
}
variable "db_password_2" {
  type = string
  default = ""
  sensitive = true
}

resource "doppler_integration_aws_postgres" "i_aws_postgres" {
  name            = "TF AWS Postgres"
  assume_role_arn = "arn:aws:iam::xxxxxxxxxxxx:role/xxxxxxxxxxxxxxxx"
  lambda_arn      = "arn:aws:lambda:xxxxxxxxx:xxxxxxxxxxxx:function:xxxxxxxxxxxxxxxxxxxx"
}

resource "doppler_rotated_secret_aws_postgres" "rs_aws_postgres" {
  integration            = doppler_integration_aws_postgres.i_aws_postgres.id
  project                = "backend"
  config                 = "dev"
  name                   = "DB"
  rotation_period_sec    = 2592000
  host                   = "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
  port                   = 5432
  database               = "xxxxxxxx"
  managing_user_username = "xxxxxxxx"
  managing_user_password = "xxxxxxxxxxxxxxxxxxxx"
  credentials {
    username = "xxxxxxxxxxx"
    password = var.db_password_1
  }
  credentials {
    username = "xxxxxxxxxxx"
    password = var.db_password_2
  }
  lifecycle {
    # The credentials are rotated regularly by Doppler, and cannot be updated via TF after initialization, so skip checking the credentials against state.
    ignore_changes = [credentials]
  }
}


