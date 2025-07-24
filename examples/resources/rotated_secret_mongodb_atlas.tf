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

resource "doppler_integration_mongodb_atlas" "i_mongodb_atlas" {
  name        = "TF MongoDB Atlas"
  public_key  = "xxxxxxxx"
  private_key = "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
}

resource "doppler_rotated_secret_mongodb_atlas" "rs_mongodb_atlas" {
  integration         = doppler_integration_mongodb_atlas.i_mongodb_atlas.id
  project             = "backend"
  config              = "dev"
  name                = "MONGODB"
  rotation_period_sec = 2592000
  project_id          = "xxxxxxxxxxxxxxxxx"
  credentials {
    username = "xxxxxxxxx"
    password = var.db_password_1
  }
  credentials {
    username = "xxxxxxxxx"
    password = var.db_password_2
  }
  lifecycle {
    # The credentials are rotated regularly by Doppler, and cannot be updated via TF after initialization, so skip checking the credentials against state.
    ignore_changes = [credentials]
  }
}

