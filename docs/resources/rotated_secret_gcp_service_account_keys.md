---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "doppler_rotated_secret_gcp_service_account_keys Resource - terraform-provider-doppler"
subcategory: ""
description: |-
  
---

# doppler_rotated_secret_gcp_service_account_keys (Resource)





<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `config` (String) The name of the Doppler config
- `integration` (String) The slug of the integration to use for this rotated secret
- `name` (String) The name of the rotated secret
- `project` (String) The name of the Doppler project
- `rotation_period_sec` (Number) How frequently to rotate the secret
- `service_account` (String) The Service Account Email whose keys should be rotated

### Read-Only

- `id` (String) The ID of this resource.
