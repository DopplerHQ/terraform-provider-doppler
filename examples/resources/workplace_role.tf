resource "doppler_workplace_role" "team_manager" {
  name        = "Team Manager"
  permissions = ["team_manage"]
}
