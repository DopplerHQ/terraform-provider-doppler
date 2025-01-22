# Single-Secret
resource "doppler_secrets_sync_azure_vault" "backend_prod" {
  integration = "bae40485-eca7-478b-abd8-34100c82c679"
  project     = "backend"
  config      = "prd"

  sync_strategy      = "single-secret"
  vault_uri          = "https://backend-prd.vault.azure.net/"
  single_secret_name = "DopplerSecrets"
}

# Multi-Secret
resource "doppler_secrets_sync_azure_vault" "backend_prod" {
  integration = "bae40485-eca7-478b-abd8-34100c82c679"
  project     = "backend"
  config      = "prd"

  sync_strategy = "multi-secret"
  vault_uri     = "https://backend-prd.vault.azure.net/"
}
