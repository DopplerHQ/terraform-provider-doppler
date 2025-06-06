---
page_title: "doppler_service_account Resource - terraform-provider-doppler"
subcategory: "Service Accounts"
description: |-
	Manage a Doppler service account.
---

# doppler_service_account (Resource)

Manage a Doppler service account.

## Example Usage

```terraform
resource "doppler_service_account" "ci" {
  name = "ci"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) The name of the service account

### Optional

- `workplace_permissions` (List of String) A list of the workplace permissions for the service account (or use `workplace_role`)
- `workplace_role` (String) The identifier of the workplace role for the service account (or use `workplace_permissions`)

### Read-Only

- `id` (String) The ID of this resource.
- `slug` (String) The slug of the service account

## Import

Import is supported using the following syntax:

```shell
# import using the service account slug from the URL:
# https://dashboard.doppler.com/workplace/[workplace-slug]/team/service_accounts/[service-account-slug]
terraform import doppler_service_account.default <service-account-slug>
```
