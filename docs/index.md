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
