resource "doppler_integration_flyio" "prod" {
  name    = "TF Fly.io"
  api_key = "fo1_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
}

resource "doppler_secrets_sync_flyio" "backend_prod" {
  integration = doppler_integration_flyio.prod.id
  project     = "backend"
  config      = "prd"

  app_id           = "my-app"
  restart_machines = true

  delete_behavior = "leave_in_target"
}
