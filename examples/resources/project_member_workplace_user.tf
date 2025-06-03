resource "doppler_project_member_workplace_user" "backend_jane_doe" {
  project      = "backend"
  workplace_user_slug   = doppler_workplace_user.jane_doe.slug
  role         = "collaborator"
  environments = ["dev", "stg"]

}
