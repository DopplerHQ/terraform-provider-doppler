data "azuread_application" "doppler_manager" {
  client_id = "77ed9112-b7f5-4d1d-a60f-198375e7f265"
}

resource "doppler_integration_azure_dynamic_service_principal_oidc" "prod" {
  name      = "Production"
  tenant_id = "c77d1b3d-6350-4696-b59f-90dae3e0b41e"
  client_id = data.azuread_application.doppler_manager.client_id
}

resource "azuread_application_federated_identity_credential" "doppler" {
  application_id = data.azuread_application.doppler_manager.id
  display_name   = "doppler-production"
  issuer         = doppler_integration_azure_dynamic_service_principal_oidc.prod.federation_issuer
  subject        = doppler_integration_azure_dynamic_service_principal_oidc.prod.federation_subject
  audiences      = [doppler_integration_azure_dynamic_service_principal_oidc.prod.federation_audience]
}
