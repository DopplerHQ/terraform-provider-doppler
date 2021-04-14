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
  token = "<YOUR DOPPLER TOKEN>"
}

data "doppler_secrets" "this" {}

output "stripe_key" {
  value = data.doppler_secrets.this.map.STRIPE_KEY
}
```

[More examples](examples/main.tf)

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
