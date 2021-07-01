# Terraform Provider Doppler

The Doppler Terraform Provider allows you to interact with your [Doppler](https://doppler.com) secrets and configuration.

# Usage

```hcl
terraform {
  required_providers {
    doppler = {
      # version = <latest version>
      source = "DopplerHQ/doppler"
    }
  }
}

provider "doppler" {
  doppler_token = "<YOUR DOPPLER TOKEN>"
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

resource "random_password" "db_password" {
  length = 32
  special = true
}

# Set secrets in Doppler
resource "doppler_secret" "db_password" {
  project = "backend"
  config = "dev"
  name = "DB_PASSWORD"
  value = random_password.db_password.result
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
