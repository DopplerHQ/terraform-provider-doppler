# Integration
resource "doppler_integration_gcp_secret_manager" "prod" {
  name              = "prod"
  gcp_key           = var.gcp_key
  gcp_secret_prefix = "doppler-"
}

# Single-Secret
resource "doppler_secrets_sync_gcp_secret_manager" "backend_prod" {
  integration = doppler_integration_gcp_secret_manager.prod.id
  project     = "backend"
  config      = "prd"

  sync_strategy = "single-secret"
  name          = "backend-prd-secrets"
  format        = "json"
  regions       = ["automatic"]

  delete_behavior = "delete_from_target"
}

# Multi-Secret
resource "doppler_secrets_sync_gcp_secret_manager" "backend_prod" {
  integration = doppler_integration_gcp_secret_manager.prod.id
  project     = "backend"
  config      = "prd"

  sync_strategy = "multi-secret"
  regions       = ["automatic"]

  delete_behavior = "delete_from_target"
}
