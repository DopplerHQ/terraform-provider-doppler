# Terraform Provider Doppler (Pre-Release)

**This software is in pre-release and is not intended to be used in production.**

The Doppler Terraform Provider allows you to interact with your [Doppler](https://doppler.com) secrets and configuration.

# Usage

```
terraform {
  required_providers {
    doppler = {
      version = "0.0.1"
      source = "DopplerHQ/doppler"
    }
  }
}

provider "doppler" {
  token = "<YOUR DOPPLER TOKEN>"
}

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

# Development

Run the following command to build the provider:

```shell
make build
# Outputs terraform-provider-doppler binary
```

## Test Sample Configuration

First, build and install the provider.

```shell
make install
```

Then, run the following command to initialize the workspace and apply the sample configuration.

```shell
terraform init && terraform apply
```

## Running the Example

```shell
# build and install
make
cd examples
# init & apply
terraform init
terraform apply --auto-approve
```
