resource "doppler_service_account_token" "builder_ci_token" {
  service_account = "builder"
  name            = "Builder CI Token"
  expires_at      = "2024-05-30T11:00:00.000Z"
}

# Service token key available as `doppler_service_account_token.builder_ci_token.api_key`
