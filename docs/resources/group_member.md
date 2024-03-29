---
page_title: "doppler_group_member Resource - terraform-provider-doppler"
subcategory: ""
description: |-
	Manage a Doppler user/group membership.
---

# doppler_group_member (Resource)

Manage a Doppler user/group membership.

## Example Usage

```terraform
resource "doppler_group" "engineering" {
  name = "engineering"
}

data "doppler_user" "nic" {
  email = "nic@doppler.com"
}

data "doppler_user" "andre" {
  email = "andre@doppler.com"
}

resource "doppler_group_member" "engineering" {
  for_each   = toset([data.doppler_user.nic.slug, data.doppler_user.andre.slug])
  group_slug = doppler_group.engineering.slug
  user_slug  = each.value
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `group_slug` (String) The slug of the Doppler group
- `user_slug` (String) The slug of the Doppler workplace user

### Read-Only

- `id` (String) The ID of this resource.

## Resource ID Format

Resource IDs are in the format `<group_slug>.workplace_user.<user_slug>`.
