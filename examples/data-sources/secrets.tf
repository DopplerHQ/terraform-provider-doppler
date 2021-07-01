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
