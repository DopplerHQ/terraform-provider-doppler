resource "doppler_trusted_ip" "backend_ci_github" {
  project = "backend"
  config = "ci_github"
  ip = "127.0.0.1/32"
}
