# The initial key for the first user. This will be rotated, so we provide a default and ignore changes below.
# Provide this with -var-file: https://developer.hashicorp.com/terraform/language/values/variables#variable-definitions-tfvars-files 
variable "key_1" {
  type = string
  default = ""
  # Consider using ephemeral instead if your client supports it: https://developer.hashicorp.com/terraform/language/values/variables#exclude-values-from-state
  sensitive = true
}
variable "key_2" {
  type = string
  default = ""
  sensitive = true
}

resource "doppler_integration_cloudflare_tokens" "i_cf" {
  name      = "TF Cloudflare"
  api_token = "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
}

resource "doppler_rotated_secret_cloudflare_tokens" "rs_cf" {
  integration         = doppler_integration_cloudflare_tokens.i_cf.id
  project             = "backend"
  config              = "dev"
  name                = "CLOUDFLARE"
  rotation_period_sec = 2592000
  credentials {
    value = var.key_1
  }
  credentials {
    value = var.key_2
  }
  lifecycle {
    # The credentials are rotated regularly by Doppler, and cannot be updated via TF after initialization, so skip checking the credentials against state.
    ignore_changes = [credentials]
  }
}

