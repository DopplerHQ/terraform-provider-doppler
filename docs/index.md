---
page_title: "doppler Provider"
description: |-
  The Doppler provider is used to interact with resources provided by Doppler. The provider must be configured with a Doppler token for authentication.
---

# Doppler Provider

The [Doppler](https://doppler.com) provider is used to interact with resources provided by Doppler. The provider must be configured with a Doppler token for authentication.

## Example Usage

```hcl
provider "doppler" {
  token = "<YOUR DOPPLER TOKEN>"
}
```

## Argument Reference

### Required

- **doppler_token** (String) A Doppler service token

### Optional

- **host** (String) The Doppler API host (i.e. https://api.doppler.com)
- **verify_tls** (Boolean) Whether or not to verify TLS

## Getting Help

If you need help, customers on our Standard, Pro, or Enterprise plan can reach out via our in-product support, or post a message in our [community forum](https://community.doppler.com/c/need-help/6).

If you'd like to report an issue or request an enhancement, please [create a new issue on the repository](https://github.com/DopplerHQ/terraform-provider-doppler/issues).
