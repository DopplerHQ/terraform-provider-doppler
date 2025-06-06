---
page_title: "doppler_project_member_user Resource - terraform-provider-doppler"
subcategory: "Project Structure"
description: |-
	Manage a Doppler project user member.
---

# doppler_project_member_user (Resource)

Manage a Doppler project user member.

## Example Usage

```terraform
data "doppler_user" "brian" {
  email = "brian@doppler.com"
}

resource "doppler_project_member_user" "backend_brian" {
  project      = "backend"
  user_slug    = data.doppler_user.brian.slug
  role         = "collaborator"
  environments = ["dev", "stg"]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `project` (String) The name of the Doppler project where the access is applied
- `role` (String) The project role identifier for the access
- `user_slug` (String) The slug of the Doppler workplace user

### Optional

- `environments` (Set of String) The environments in the project where this access will apply (null or omitted for roles with access to all environments)

### Read-Only

- `id` (String) The ID of this resource.
