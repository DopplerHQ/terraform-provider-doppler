---
page_title: "doppler_group Resource - terraform-provider-doppler"
subcategory: "Groups"
description: |-
	Manage a Doppler group.
---

# doppler_group (Resource)

Manage a Doppler group.

## Example Usage

```terraform
resource "doppler_group" "engineering" {
  name           = "engineering"
  workplace_role = "viewer"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) The name of the group

### Optional

- `default_project_role` (String) The default project role assigned to the group when added to a Doppler project. If set to null, the default project role is inherited from the workplace setting.
- `workplace_role` (String) The workplace role assigned to members of the group. If omitted, state will be tracked in Terraform but not updated in Doppler. Use "no_access" to ensure the group has no workplace permissions

### Read-Only

- `id` (String) The ID of this resource.
- `slug` (String) The slug of the group

## Import

Import is supported using the following syntax:

```shell
# import using the group slug from the URL:
# https://dashboard.doppler.com/workplace/[workplace-slug]/team/groups/[group-slug]
terraform import doppler_group.default <group-slug>
```
