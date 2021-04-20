---
page_title: "doppler_secrets Data Source - terraform-provider-doppler"
subcategory: "Secrets"
description: |-
  Retrieves all secrets in the config.
---

# doppler_secrets (Data Source)

Retrieves all secrets in the config.

## Example Usage

```hcl
data "doppler_secrets" "this" {}

# Access individual secrets
output "stripe_key" {
  value = data.doppler_secrets.this.map.STRIPE_KEY
}

# Use `tonumber` and `tobool` to parse string values into Terraform primatives
output "max_workers" {
  value = tonumber(data.doppler_secrets.this.map.MAX_WORKERS)
}

# JSON values can be decoded direcly in Terraform
# e.g. FEATURE_FLAGS = `{ "AUTOPILOT": true, "TOP_SPEED": 130 }`
output "json_parsing_values" {
  value = jsondecode(data.doppler_secrets.this.map.FEATURE_FLAGS)["TOP_SPEED"]
}
```

## Attribute Reference

- **map** (Map of String, Sensitive) A map of secret names to computed secret values
