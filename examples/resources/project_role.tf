resource "doppler_project_role" "log_viewer" {
  name        = "Log Viewer"
  permissions = ["enclave_config_logs"]
}
