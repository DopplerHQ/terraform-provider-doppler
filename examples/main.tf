terraform {
  required_providers {
    doppler = {
      # version = <latest version>
      source = "DopplerHQ/doppler"
    }
  }
}

###
### Setup the Doppler provider
###

variable "doppler_token" {
  type = string
}

# The provider must always be specified with authentication
provider "doppler" {
  # Your Doppler token, either a personal or service token
  doppler_token = var.doppler_token
  # The token can be provided with the environment variable `DOPPLER_TOKEN` instead
}

###
### Read Doppler secrets with the doppler_secrets data provider
###

# Mapped access to secrets
data "doppler_secrets" "this" {
  # Project and config are required if you are using a personal token
  project = "backend"
  config = "dev"
}

output "all_secrets" {
  # nonsensitive used for demo purposes only
  value = nonsensitive(data.doppler_secrets.this.map)
}

output "stripe_key" {
  # Individual keys can be accessed directly by name
  value = nonsensitive(data.doppler_secrets.this.map.STRIPE_KEY)
}

output "string_parsing" {
  # Use `tonumber` and `tobool` to parse string values into Terraform primatives
  value = nonsensitive(tonumber(data.doppler_secrets.this.map.MAX_WORKERS))
}

output "json_parsing_values" {
  # JSON values can be decoded direcly in Terraform
  # e.g. FEATURE_FLAGS = `{ "AUTOPILOT": true, "TOP_SPEED": 130 }`
  value = nonsensitive(jsondecode(data.doppler_secrets.this.map.FEATURE_FLAGS)["TOP_SPEED"])
}

###
### Create and modify Doppler secrets with the `doppler_secret` resource
###

resource "random_password" "db_password" {
  length = 32
  special = true
}

resource "doppler_secret" "db_password" {
  project = "backend"
  config = "dev"
  name = "DB_PASSWORD"
  value = random_password.db_password.result
}

output "resource_value" {
  # Access the raw secret value
  value = nonsensitive(doppler_secret.db_password.value)
}

output "resource_computed" {
  # Access the computed secret value (if using Doppler secrets referencing)
  value = nonsensitive(doppler_secret.db_password.computed)
}

###
### Create and modify Doppler projects, environments, configs, and service tokens
###

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

output "token_key" {
  # Access the service token key
  value = nonsensitive(doppler_service_token.ci_github_token.key)
}
