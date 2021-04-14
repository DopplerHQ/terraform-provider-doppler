terraform {
  required_providers {
    doppler = {
      version = "0.1"
      source  = "doppler.com/core/doppler"
    }
  }
}

# The provider must always be specified with authentication
provider "doppler" {
  # Your Doppler token
  token = "<YOUR DOPPLER TOKEN>"
  # The token can be provided with the environment variable `DOPPLER_TOKEN` instead
}

# Mapped access to computed secrets
data "doppler_secrets" "this" {
  # `type` is "computed" by default but can be set to "raw"
  # type = "raw"
}

output "all_secrets_computed" {
  value = data.doppler_secrets.this.db
}

output "stripe_key" {
  # Individual keys can be accessed directly by name
  value = data.doppler_secrets.this.db.STRIPE_KEY
}

output "stripe_key_lower" {
  # Secret names are also available in lower case
  value = data.doppler_secrets.this.db.stripe_key
}

output "string_parsing" {
  # Use `tonumber` and `tobool` to parse string values into Terraform primatives
  value = tonumber(data.doppler_secrets.this.db.MAX_WORKERS)
}

output "json_parsing_values" {
  # JSON values can be decoded direcly in Terraform
  # e.g. FEATURE_FLAGS = `{ "AUTOPILOT": true, "TOP_SPEED": 130 }`
  value = jsondecode(data.doppler_secrets.this.db.FEATURE_FLAGS)["TOP_SPEED"]
}

output "json_parsing_to_map" {
  # JSON values can also be parsed into a Terraform map
  # e.g. ID_MAP = `{ "AUTOPILOT": true, "TOP_SPEED": 130 }`
  value = tomap(jsondecode(data.doppler_secrets.this.db.ID_MAP))
}
