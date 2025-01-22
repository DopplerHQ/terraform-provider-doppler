# Integration
resource "doppler_integration_azure_vault_service_principal" "prod" {
  name          = "prod"
  client_id     = "77ed9112-b7f5-4d1d-a60f-198375e7f265"
  client_secret = "kXOQb4juoe~jEqQVQdeXEFhsgIvkmoE0Qz3jBwPo"
  tenant_id     = "c77d1b3d-6350-4696-b59f-90dae3e0b41e"
}

# Single-Secret
resource "doppler_secrets_sync_azure_vault_service_principal" "backend_prod" {
  integration = doppler_integration_azure_vault_service_principal.prod.id
  project     = "backend"
  config      = "prd"

  sync_strategy      = "single-secret"
  vault_uri          = "https://backend-prd.vault.azure.net/"
  single_secret_name = "DopplerSecrets"
}

# Multi-Secret
resource "doppler_secrets_sync_azure_vault_service_principal" "backend_prod" {
  integration = doppler_integration_azure_vault_service_principal.prod.id
  project     = "backend"
  config      = "prd"

  sync_strategy = "multi-secret"
  vault_uri     = "https://backend-prd.vault.azure.net/"
}
