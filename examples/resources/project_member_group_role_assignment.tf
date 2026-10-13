data "doppler_group" "engineering" {
  slug = "engineering"
}

resource "doppler_project_member_group_role_assignment" "backend_engineering_viewer" {
  project      = "backend"
  group_slug   = data.doppler_group.engineering.slug
  role         = "viewer"
  environments = ["dev", "stg"]
}

resource "doppler_project_member_group_role_assignment" "backend_engineering_collaborator" {
  project      = "backend"
  group_slug   = data.doppler_group.engineering.slug
  role         = "collaborator"
  environments = ["prd"]
}
