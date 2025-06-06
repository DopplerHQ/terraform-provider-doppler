---
page_title: "doppler_integration_circleci Resource - terraform-provider-doppler"
subcategory: "Integrations"
description: |-
	Manage a CircleCI Doppler integration.
---

# doppler_integration_circleci (Resource)

Manage a CircleCI Doppler integration.

## Example Usage

```terraform
resource "doppler_integration_circleci" "prod" {
  name      = "Production"
  api_token = "my_api_token"
}

resource "doppler_secrets_sync_circleci" "backend_prod" {
  integration = doppler_integration_circleci.prod.id
  project     = "backend"
  config      = "prd"

  resource_type     = "project"
  resource_id       = "github/myorg/myproject"
  organization_slug = "myorg"

  delete_behavior = "leave_in_target"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `api_token` (String, Sensitive) A CircleCI API token. See https://docs.doppler.com/docs/circleci for details.
- `name` (String) The name of the integration

### Read-Only

- `id` (String) The ID of this resource.
