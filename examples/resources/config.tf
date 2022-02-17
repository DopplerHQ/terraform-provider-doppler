resource "doppler_config" "backend_ci_github" {
  project = "backend"
  environment = "ci"
  name = "ci_github"
}
