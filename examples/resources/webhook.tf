resource "doppler_webhook" "ci" {
  project = doppler_project.test_proj.name
  url = "https://localhost/webhook"
  name = "My Webhook"
  secret = "my signing secret-2"
  enabled = true
  enabled_configs = [doppler_config.ci_github.name]
  authentication {
    type = "Bearer"
    token = "my bearer token"
  }
  payload = jsonencode({
    myKey = "my value"
  })
}