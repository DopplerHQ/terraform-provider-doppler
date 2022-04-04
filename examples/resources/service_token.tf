resource "doppler_service_token" "backend_ci_token" {
  project = "backend"
  config = "ci"
  name = "Builder Token"
  access = "read"
}

# Service token key available as `doppler_service_token.backend_ci_token.key`
