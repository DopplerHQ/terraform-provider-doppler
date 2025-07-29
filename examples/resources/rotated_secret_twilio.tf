resource "doppler_integration_twilio" "i_twilio" {
  name        = "TF Twilio"
  account_sid = "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
  key_sid     = "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
  key_secret  = "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
}

resource "doppler_rotated_secret_twilio" "rs_twilio" {
  integration         = doppler_integration_twilio.i_twilio.id
  project             = "backend"
  config              = "dev"
  name                = "TWILIO"
  rotation_period_sec = 2592000
}

