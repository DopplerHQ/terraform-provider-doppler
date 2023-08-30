resource "doppler_project_member_group" "backend_engineering" {
  project      = "backend"
  group_slug   = doppler_group.engineering.slug
  role         = "collaborator"
  environments = ["dev", "stg"]

}
