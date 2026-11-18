resource "doppler_secrets_sync_vercel" "backend_prod" {
  integration = "bae40485-eca7-478b-abd8-34100c82c679"
  project     = "backend"
  config      = "prd"

  team_id       = "team_xxxxxxxxxx"
  project_id    = "prj_xxxxxxxxxx"
  target_id     = "production"
  variable_type = "secret"

  delete_behavior = "leave_in_target"
}

resource "doppler_secrets_sync_vercel" "backend_stg" {
  integration = "bae40485-eca7-478b-abd8-34100c82c679"
  project     = "backend"
  config      = "stg"

  project_id = "prj_xxxxxxxxxx"
  target_id  = "preview"

  delete_behavior = "leave_in_target"
}
