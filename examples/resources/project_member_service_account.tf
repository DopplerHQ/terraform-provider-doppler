resource "doppler_project_member_service_account" "backend_ci" {
  project              = "backend"
  service_account_slug = doppler_service_account.ci.slug
  role                 = "viewer"
  environments         = ["dev", "stg", "prd"]

}
