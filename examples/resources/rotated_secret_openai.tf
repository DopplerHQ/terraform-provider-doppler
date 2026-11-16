resource "doppler_integration_openai" "i_openai" {
  name      = "TF OpenAI"
  admin_key = "sk-admin-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
}

resource "doppler_rotated_secret_openai" "rs_openai" {
  integration         = doppler_integration_openai.i_openai.id
  project             = "backend"
  config              = "dev"
  name                = "OPENAI"
  rotation_period_sec = 2592000
  project_id          = "proj_xxxxxxxxxxxxxxxxxxxxxxxx"
}
