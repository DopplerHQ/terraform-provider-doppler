resource "doppler_integration_member_service_account" "aws_ci" {
  integration          = doppler_integration_aws_secrets_manager.prod.slug
  service_account_slug = doppler_service_account.ci.slug
  role                 = "consumer"
}
