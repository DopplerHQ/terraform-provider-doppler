resource "doppler_environment" "backend_ci" {
  project = "backend"
  slug = "ci"
  name = "Continuous Integration"
}
