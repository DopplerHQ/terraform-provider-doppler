package doppler

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func awsAssumeRoleDataSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"assume_role_arn": {
			Description: "The ARN of the AWS role for Doppler to assume",
			Type:        schema.TypeString,
			Required:    true,
		},
	}
}

func awsAssumeRoleDataBuilder(d *schema.ResourceData) IntegrationData {
	return IntegrationData{
		"aws_assume_role_arn": d.Get("assume_role_arn"),
	}
}

func resourceIntegrationAWSAssumeRoleIntegration(integrationType string) *schema.Resource {
	builder := ResourceIntegrationBuilder{
		Type:        integrationType,
		DataSchema:  awsAssumeRoleDataSchema(),
		DataBuilder: awsAssumeRoleDataBuilder,
	}
	return builder.Build()
}

func resourceIntegrationCircleCi() *schema.Resource {
	builder := ResourceIntegrationBuilder{
		Type: "circleci",
		DataSchema: map[string]*schema.Schema{
			"api_token": {
				Description: "A CircleCI API token. See https://docs.doppler.com/docs/circleci for details.",
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
			},
		},
		DataBuilder: func(d *schema.ResourceData) IntegrationData {
			return map[string]interface{}{
				"api_token": d.Get("api_token"),
			}
		},
	}
	return builder.Build()
}

func resourceIntegrationTerraformCloud() *schema.Resource {
	builder := ResourceIntegrationBuilder{
		Type: "terraform_cloud",
		DataSchema: map[string]*schema.Schema{
			"api_key": {
				Description: "A Terraform Cloud API key.",
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
			},
		},
		DataBuilder: func(d *schema.ResourceData) IntegrationData {
			return map[string]interface{}{
				"api_key": d.Get("api_key"),
			}
		},
	}
	return builder.Build()
}

func resourceIntegrationFlyio() *schema.Resource {
	builder := ResourceIntegrationBuilder{
		Type: "flyio",
		DataSchema: map[string]*schema.Schema{
			"api_key": {
				Description: "A Fly.io API key.",
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
			},
		},
		DataBuilder: func(d *schema.ResourceData) IntegrationData {
			return map[string]interface{}{
				"api_key": d.Get("api_key"),
			}
		},
	}
	return builder.Build()
}

func resourceIntegrationAzureVaultServicePrincipal() *schema.Resource {
	builder := ResourceIntegrationBuilder{
		Type: "azure_vault_service_principal",
		DataSchema: map[string]*schema.Schema{
			"client_id": {
				Description: "The Service Principal Client ID. See https://docs.doppler.com/docs/azure-key-vault#custom-service-principal for details.",
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   false,
			},
			"client_secret": {
				Description: "The Service Principal Client Secret. See https://docs.doppler.com/docs/azure-key-vault#custom-service-principal for details.",
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
			},
			"tenant_id": {
				Description: "The Service Principal Tenant ID. See https://docs.doppler.com/docs/azure-key-vault#custom-service-principal for details.",
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   false,
			},
		},
		DataBuilder: func(d *schema.ResourceData) IntegrationData {
			return map[string]interface{}{
				"clientId":     d.Get("client_id"),
				"clientSecret": d.Get("client_secret"),
				"tenantId":     d.Get("tenant_id"),
			}
		},
	}
	return builder.Build()
}
