resource "google_iam_workload_identity_pool" "doppler" {
  workload_identity_pool_id = "doppler-pool"
}

resource "google_iam_workload_identity_pool_provider" "doppler" {
  workload_identity_pool_id          = google_iam_workload_identity_pool.doppler.workload_identity_pool_id
  workload_identity_pool_provider_id = "doppler"

  oidc {
    issuer_uri = "https://api.doppler.com"
  }

  attribute_mapping = {
    "google.subject" = "assertion.sub"
  }

  attribute_condition = "assertion.workplace == '<YOUR_WORKPLACE_SLUG>'"
}

resource "doppler_integration_gcp_secret_manager_oidc" "prod" {
  name                           = "Production"
  gcp_workload_identity_provider = "//iam.googleapis.com/${google_iam_workload_identity_pool_provider.doppler.name}"
  gcp_project_id                 = "your-gcp-project-id"
  gcp_secret_prefix              = "doppler-"
}

resource "google_project_iam_member" "doppler" {
  project = "your-gcp-project-id"
  role    = "roles/secretmanager.admin"
  member  = doppler_integration_gcp_secret_manager_oidc.prod.federation_principal

  condition {
    title      = "doppler-prefix-only"
    expression = "resource.name.startsWith(\"projects/<YOUR_PROJECT_NUMBER>/secrets/doppler-\")"
  }
}
