data "tfe_workspace" "prod" {
  name         = "my-workspace-name"
  organization = "my-org-name"
}

resource "doppler_integration_terraform_cloud" "prod" {
  name    = "Production"
  api_key = "my_api_key"
}

resource "doppler_secrets_sync_terraform_cloud" "backend_prod" {
  integration = doppler_integration_terraform_cloud.prod.id
  project     = "backend"
  config      = "prd"

  sync_target        = "workplace"
  workplace_id       = data.tfe_workspace.prod.id
  variable_sync_type = "terraform"
  name_transform     = "lowercase"
}
