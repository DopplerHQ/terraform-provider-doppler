# The default is specified here specifically to indicate what kind of value is
# expected. In production use, don't hard-code the key! Instead pass that
# in using `TF_VAR_gcp_key`.
variable "gcp_key" {
  type    = string
  default = <<-EOT
  {
    "type": "service_account",
    "project_id": "your-gcp-project-id",
    "private_key_id": "...",
    "private_key": "-----BEGIN PRIVATE KEY-----\n ... \n-----END PRIVATE KEY-----\n",
    "client_email": "doppler-secret-manager@your-gcp-project-id.iam.gserviceaccount.com",
    "client_id": "12345678901234567890",
    "auth_uri": "https://accounts.google.com/o/oauth2/auth",
    "token_uri": "https://oauth2.googleapis.com/token",
    "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
    "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/doppler-secret-manager%40your-gcp-project-id.iam.gserviceaccount.com",
    "universe_domain": "googleapis.com"
  }
  EOT
}

resource "doppler_integration_gcp_secret_manager" "prod" {
  name              = "prod"
  gcp_key           = var.gcp_key
  gcp_secret_prefix = "doppler-"
}
