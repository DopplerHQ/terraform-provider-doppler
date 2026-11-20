# Requires Terraform 1.11 or later.
ephemeral "random_password" "db_password" {
  length  = 32
  special = true
}

resource "doppler_secret" "db_password" {
  project = "backend"
  config  = "dev"
  name    = "DB_PASSWORD"

  # Never stored in the Terraform plan or state.
  value_wo = ephemeral.random_password.db_password.result
  # Increment to write a new value to Doppler.
  value_wo_version = 1
}
