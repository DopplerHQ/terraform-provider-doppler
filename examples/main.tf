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
  # Providing a host is optional, helpful for testing
  # host = "https://staging-api.doppler.com"

  # Your Doppler API key / service token
  api_key = "<DOPPLER API KEY>"

  # Host can be provided with the environment variable `DOPPLER_API_HOST`
  # API key can be provided with the environment variable `DOPPLER_TOKEN`
}

# Mapped access to computed secrets
data "doppler_secrets" "computed" {
  format = "computed"
}
output "all_secrets_computed" {
  value = data.doppler_secrets.computed.secrets
  # e.g `data.doppler_secrets.computed.secrets.STRIPE_KEY`
}

# Mapped access to raw secrets
data "doppler_secrets" "raw" {
  format = "raw"
}
output "all_secrets_raw" {
  value = data.doppler_secrets.raw.secrets
  # e.g. `data.doppler_secrets.raw.secrets.STRIPE_KEY`
}

# Both raw and computed formats can also be used with `lowercase`,
# which converts all secret names to lowercase (a Terraform naming convention).
data "doppler_secrets" "lower" {
  format = "computed"
  lowercase = true
}
output "all_secrets_lower" {
  value = data.doppler_secrets.lower.secrets
  # e.g `data.doppler_secrets.lower.secrets.stripe_key`
}

# Full access to the secrets objects, mostly for advanced use cases
data "doppler_secrets_objects" "objects" {}
output "all_secrets_objects" {
  value = data.doppler_secrets_objects.objects
}
