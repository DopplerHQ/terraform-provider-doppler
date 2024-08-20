resource "doppler_integration_circleci" "prod" {
  name      = "Production"
  api_token = "my_api_token"
}

resource "doppler_secrets_sync_circleci" "backend_prod" {
  integration = doppler_integration_circleci.prod.id
  project     = "backend"
  config      = "prd"

  resource_type     = "project"
  resource_id       = "github/myorg/myproject"
  organization_slug = "myorg"

  delete_behavior = "leave_in_target"
}
