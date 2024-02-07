### Basic Usage

data "doppler_secrets" "this" {}

# Access individual secrets
output "stripe_key" {
  # nonsensitive used for demo purposes only
  value = nonsensitive(data.doppler_secrets.this.map.STRIPE_KEY)
}

# Use `tonumber` and `tobool` to parse string values into Terraform primatives
output "max_workers" {
  value = nonsensitive(tonumber(data.doppler_secrets.this.map.MAX_WORKERS))
}

# JSON values can be decoded direcly in Terraform
# e.g. FEATURE_FLAGS = `{ "AUTOPILOT": true, "TOP_SPEED": 130 }`
output "json_parsing_values" {
  value = nonsensitive(jsondecode(data.doppler_secrets.this.map.FEATURE_FLAGS)["TOP_SPEED"])
}

### Referencing secrets from multiple projects

variable "doppler_token_dev" {
  type = string
  description = "A token to authenticate with Doppler for the dev config"
}

variable "doppler_token_prd" {
  type = string
  description = "A token to authenticate with Doppler for the prd config"
}

provider "doppler" {
  doppler_token = var.doppler_token_dev
  alias = "dev"
}

provider "doppler" {
  doppler_token = var.doppler_token_prd
  alias = "prd"
}

data "doppler_secrets" "dev" {
  provider = doppler.dev
}

data "doppler_secrets" "prd" {
  provider = doppler.prd
}

output "port-dev" {
  value = nonsensitive(data.doppler_secrets.dev.map.PORT)
}

output "port-prd" {
  value = nonsensitive(data.doppler_secrets.prd.map.PORT)
}
