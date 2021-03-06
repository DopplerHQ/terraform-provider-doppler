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

variable "doppler_token" {
  type = string
}

provider "doppler" {
  doppler_token = var.doppler_token
}

data "doppler_secrets" "this" {
  project = "backend"
  config = "dev"
}

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

# Create and modify Doppler projects, environments, configs, and service tokens

resource "doppler_project" "test_proj" {
  name = "my-test-project"
  description = "This is a test project"
}

resource "doppler_environment" "ci" {
  project = doppler_project.test_proj.name
  slug = "ci"
  name = "CI-CD"
}

resource "doppler_config" "ci_github" {
  project = doppler_project.test_proj.name
  environment = doppler_environment.ci.slug
  name = "ci_github"
}

resource "doppler_service_token" "ci_github_token" {
  project = doppler_project.test_proj.name
  config = doppler_config.ci_github.name
  name = "test token"
  access = "read"
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

Update `examples/main.tf` with the local development provider:

```hcl
terraform {
  required_providers {
    doppler = {
      source  = "doppler.com/core/doppler"
    }
  }
}
```

Then, run the following command to initialize the workspace and apply the sample configuration.

```shell
cd examples
terraform init && terraform apply
```
