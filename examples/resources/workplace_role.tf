resource "doppler_workplace_role" "log_viewer" {
  name        = "Team Manager"
  permissions = ["team_manage"]
}
