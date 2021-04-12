# Terraform Provider Doppler

# Installation

Currently, this provider does not exist on the Terraform registry and must be manually installed.

If `doppler` is listed in the `required_providers`, the CLI will look for the Doppler provider binary in `~/.terraform.d/plugins/${host_name}/${namespace}/${type}/${version}/${target}` or `%APPDATA%\terraform.d\plugins\${host_name}/${namespace}/${type}/${version}/${target}`.

Using `make install` will automatically build and install the Doppler provider binary to the target location.

# Usage

```
terraform {
  required_providers {
    doppler = {
      version = "0.1"
      source  = "doppler.com/core/doppler"
    }
  }
}

provider "doppler" {
  api_key = "<DOPPLER API KEY>"

  # Host can also be provided with the environment variable `DOPPLER_API_HOST`
  # API key can also be provided with the environment variable `DOPPLER_TOKEN`
}

data "doppler_secrets" "computed" {
  # Load variables as either "computed" or "raw"
  format = "computed"
}

output "all_secrets_computed" {
  value = data.doppler_secrets.computed.secrets
  # Or specific variables with `data.doppler_secrets.computed.secrets.STRIPE_KEY`
}


# Full access to the secrets objects, mostly for advanced use cases
data "doppler_secrets_objects" "objects" {}
output "all_secrets_objects" {
  value = data.doppler_secrets_objects.objects
}
```

# Development

Run the following command to build the provider

```shell
go build -o terraform-provider-doppler
```

## Test sample configuration

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
# remove existing lock file, if present
rm -f .terraform.lock.hcl
# init & apply
terraform init
terraform apply --auto-approve
```
