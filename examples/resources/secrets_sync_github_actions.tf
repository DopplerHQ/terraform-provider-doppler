# Repo
resource "doppler_secrets_sync_github_actions" "backend_prod" {
  integration = "bae40485-eca7-478b-abd8-34100c82c679"
  project     = "backend"
  config      = "prd"

  sync_target = "repo"
  repo_name   = "backend"
}

# Repo + Environment
resource "doppler_secrets_sync_github_actions" "backend_prod" {
  integration = "bae40485-eca7-478b-abd8-34100c82c679"
  project     = "backend"
  config      = "prd"

  sync_target      = "repo"
  repo_name        = "backend"
  environment_name = "production"
}

# Org
resource "doppler_secrets_sync_github_actions" "backend_prod" {
  integration = "bae40485-eca7-478b-abd8-34100c82c679"
  project     = "backend"
  config      = "prd"

  sync_target = "org"
  org_scope   = "private"
}
