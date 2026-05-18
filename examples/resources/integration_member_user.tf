data "doppler_user" "brian" {
  email = "brian@doppler.com"
}

resource "doppler_integration_member_user" "aws_brian" {
  integration = doppler_integration_aws_secrets_manager.prod.slug
  user_slug   = data.doppler_user.brian.slug
  role        = "consumer"
}
