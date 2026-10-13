data "doppler_user" "brian" {
  email = "brian@doppler.com"
}

resource "doppler_project_member_user_role_assignment" "backend_brian_viewer" {
  project      = "backend"
  user_slug    = data.doppler_user.brian.slug
  role         = "viewer"
  environments = ["dev", "stg"]
}

resource "doppler_project_member_user_role_assignment" "backend_brian_collaborator" {
  project      = "backend"
  user_slug    = data.doppler_user.brian.slug
  role         = "collaborator"
  environments = ["prd"]
}
