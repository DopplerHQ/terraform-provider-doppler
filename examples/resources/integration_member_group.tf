resource "doppler_integration_member_group" "aws_engineering" {
  integration = doppler_integration_aws_secrets_manager.prod.slug
  group_slug  = doppler_group.engineering.slug
  role        = "consumer"
}
