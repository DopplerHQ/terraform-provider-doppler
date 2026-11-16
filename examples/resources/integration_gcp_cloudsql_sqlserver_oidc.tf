resource "doppler_integration_gcp_cloudsql_sqlserver_oidc" "prod" {
  name                           = "Production"
  gcp_workload_identity_provider = "//iam.googleapis.com/projects/123456789/locations/global/workloadIdentityPools/doppler-pool/providers/doppler"
  gcp_project_id                 = "your-gcp-project-id"
}

resource "google_project_iam_member" "doppler" {
  project = "your-gcp-project-id"
  role    = "roles/cloudsql.admin"
  member  = doppler_integration_gcp_cloudsql_sqlserver_oidc.prod.federation_principal
}
