---
page_title: "doppler_integration_terraform_cloud Resource - terraform-provider-doppler"
subcategory: "Integrations"
description: |-
	Manage a Terraform Cloud Doppler integration.
---

# doppler_integration_terraform_cloud (Resource)

Manage a Terraform Cloud Doppler integration.

## Example Usage

```terraform
data "tfe_workspace" "prod" {
  name         = "my-workspace-name"
  organization = "my-org-name"
}

resource "doppler_integration_terraform_cloud" "prod" {
  name    = "Production"
  api_key = "my_api_key"
}

resource "doppler_secrets_sync_terraform_cloud" "backend_prod" {
  integration = doppler_integration_terraform_cloud.prod.id
  project     = "backend"
  config      = "prd"

  sync_target        = "workspace"
  workspace_id       = data.tfe_workspace.prod.id
  variable_sync_type = "terraform"
  name_transform     = "lowercase"

  delete_behavior = "leave_in_target"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `api_key` (String, Sensitive) A Terraform Cloud API key.
- `name` (String) The name of the integration

### Read-Only

- `id` (String) The ID of this resource.
