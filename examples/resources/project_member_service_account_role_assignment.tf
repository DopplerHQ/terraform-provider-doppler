resource "doppler_project_member_service_account_role_assignment" "backend_ci_viewer" {
  project              = "backend"
  service_account_slug = doppler_service_account.ci.slug
  role                 = "viewer"
  environments         = ["dev", "stg"]
}

resource "doppler_project_member_service_account_role_assignment" "backend_ci_collaborator" {
  project              = "backend"
  service_account_slug = doppler_service_account.ci.slug
  role                 = "collaborator"
  environments         = ["prd"]
}
