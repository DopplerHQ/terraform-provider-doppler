resource "doppler_secrets_sync_supabase" "backend_prod" {
  integration = "bae40485-eca7-478b-abd8-34100c82c679"
  project     = "backend"
  config      = "prd"

  project_id = "abcdefghijklmnopqrst"

  delete_behavior = "leave_in_target"
}
