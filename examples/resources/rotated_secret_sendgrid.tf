resource "doppler_integration_sendgrid" "i_sendgrid" {
  name    = "TF SendGrid"
  api_key = "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
}

resource "doppler_rotated_secret_sendgrid" "rs_sendgrid" {
  integration         = doppler_integration_sendgrid.i_sendgrid.id
  project             = "backend"
  config              = "dev"
  name                = "SENDGRID"
  rotation_period_sec = 2592000
  scopes              = ["alerts.create", "alerts.read"]
}
