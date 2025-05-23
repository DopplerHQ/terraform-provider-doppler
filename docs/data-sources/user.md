---
page_title: "doppler_user Data Source - terraform-provider-doppler"
subcategory: "Users"
description: |-
  Retrieve an existing Doppler user.
---

# doppler_user (Data Source)

Retrieve an existing Doppler user.

## Example Usage

```terraform
data "doppler_user" "nic" {
  email = "nic@doppler.com"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `email` (String) The email address of the Doppler user

### Read-Only

- `id` (String) The ID of this resource.
- `slug` (String) The slug of the Doppler user
