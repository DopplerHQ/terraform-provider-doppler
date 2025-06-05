data "doppler_user" "brian" {
  email = "brian@doppler.com"
}

resource "doppler_project_member_user" "backend_brian" {
  project      = "backend"
  user_slug    = data.doppler_user.brian.slug
  role         = "collaborator"
  environments = ["dev", "stg"]
}
