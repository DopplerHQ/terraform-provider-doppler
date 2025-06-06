---
page_title: "doppler_secrets_sync_azure_vault Resource - terraform-provider-doppler"
subcategory: "Integrations"
description: |-
	Manage an Azure Vault Doppler sync.
---

# doppler_secrets_sync_azure_vault (Resource)

Manage an Azure Vault Doppler sync.

## Example Usage

```terraform
# Integration
resource "doppler_integration_azure_vault_service_principal" "prod" {
  name          = "prod"
  client_id     = "77ed9112-b7f5-4d1d-a60f-198375e7f265"
  client_secret = "kXOQb4juoe~jEqQVQdeXEFhsgIvkmoE0Qz3jBwPo"
  tenant_id     = "c77d1b3d-6350-4696-b59f-90dae3e0b41e"
}

# Single-Secret
resource "doppler_secrets_sync_azure_vault" "backend_prod" {
  integration = doppler_integration_azure_vault_service_principal.prod.id
  project     = "backend"
  config      = "prd"

  sync_strategy      = "single-secret"
  vault_uri          = "https://backend-prd.vault.azure.net/"
  single_secret_name = "DopplerSecrets"

  delete_behavior = "delete_from_target"
}

# Multi-Secret
resource "doppler_secrets_sync_azure_vault" "backend_prod" {
  integration = doppler_integration_azure_vault_service_principal.prod.id
  project     = "backend"
  config      = "prd"

  sync_strategy = "multi-secret"
  vault_uri     = "https://backend-prd.vault.azure.net/"

  delete_behavior = "delete_from_target"
}

# Non-Service Principal Integration
resource "doppler_secrets_sync_azure_vault" "backend_prod" {
  # Since you can't create oauth Azure Vault integrations via API you need to
  # manually supply the integration UUID, which can be found via API or from
  # the dashboard URL on the sync creation page
  integration = "c38716b1-2b3b-41e3-bc40-c95e55bc339b"
  project     = "backend"
  config      = "prd"

  sync_strategy      = "single-secret"
  vault_uri          = "https://backend-prd.vault.azure.net/"
  single_secret_name = "DopplerSecrets"

  delete_behavior = "delete_from_target"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `config` (String) The name of the Doppler config
- `integration` (String) The slug of the integration to use for this sync
- `project` (String) The name of the Doppler project
- `sync_strategy` (String) Determines whether secrets are synced to a single secret (`single-secret`) as a JSON object or multiple discrete secrets (`multi-secret`).
- `vault_uri` (String) The Azure Vault URI for the vault secrets will be synced to.

### Optional

- `delete_behavior` (String) The behavior to be performed on the secrets in the sync target when this resource is deleted or recreated. Either `leave_in_target` (default) or `delete_from_target`.
- `single_secret_name` (String) The name of the secret being synced to when using the "single-secret" sync strategy. Required when using "single-secret" sync strategy.

### Read-Only

- `id` (String) The ID of this resource.
