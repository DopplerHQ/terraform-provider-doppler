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

func resourceIntegrationGCPSecretManager() *schema.Resource {
	builder := ResourceIntegrationBuilder{
		Type: "gcp_secret_manager",
		DataSchema: map[string]*schema.Schema{
			"gcp_key": {
				Description: "The IAM Service Account JSON key. See https://docs.doppler.com/docs/gcp-secret-manager for details.",
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
			},
			"gcp_secret_prefix": {
				Description: "The prefix added to any secret created by this integration in GCP. See https://docs.doppler.com/docs/gcp-secret-manager for details.",
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   false,
			},
		},
		DataBuilder: func(d *schema.ResourceData) IntegrationData {
			return map[string]interface{}{
				"gcp_key":           d.Get("gcp_key"),
				"gcp_secret_prefix": d.Get("gcp_secret_prefix"),
			}
		},
	}
	return builder.Build()
}

func resourceIntegrationTwilio() *schema.Resource {
	builder := ResourceIntegrationBuilder{
		Type: "twilio",
		DataSchema: map[string]*schema.Schema{
			"account_sid": {
				Description: "The Account SID",
				Type:        schema.TypeString,
				Required:    true,
			},
			"key_sid": {
				Description: "The Key SID (cannot equal accountSID)",
				Type:        schema.TypeString,
				Required:    true,
			},
			"key_secret": {
				Description: "The Key Secret",
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
			},
		},
		DataBuilder: func(d *schema.ResourceData) IntegrationData {
			return map[string]interface{}{
				"accountSID": d.Get("account_sid"),
				"keySID":     d.Get("key_sid"),
				"keySecret":  d.Get("key_secret"),
			}
		},
	}
	return builder.Build()
}

func resourceIntegrationCloudflareTokens() *schema.Resource {
	builder := ResourceIntegrationBuilder{
		Type: "cloudflare_tokens",
		DataSchema: map[string]*schema.Schema{
			"api_token": {
				Description: "The API Token",
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
			},
		},
		DataBuilder: func(d *schema.ResourceData) IntegrationData {
			return map[string]interface{}{
				"apiToken": d.Get("api_token"),
			}
		},
	}
	return builder.Build()
}

func resourceIntegrationMongoDBAtlas() *schema.Resource {
	builder := ResourceIntegrationBuilder{
		Type: "mongodb_atlas",
		DataSchema: map[string]*schema.Schema{
			"public_key": {
				Description: "The Public Key",
				Type:        schema.TypeString,
				Required:    true,
			},
			"private_key": {
				Description: "The Private Key",
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
			},
		},
		DataBuilder: func(d *schema.ResourceData) IntegrationData {
			return map[string]interface{}{
				"publicKey":  d.Get("public_key"),
				"privateKey": d.Get("private_key"),
			}
		},
	}
	return builder.Build()
}

func resourceIntegrationSendGrid() *schema.Resource {
	builder := ResourceIntegrationBuilder{
		Type: "sendgrid",
		DataSchema: map[string]*schema.Schema{
			"api_key": {
				Description: "The API Key",
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
			},
		},
		DataBuilder: func(d *schema.ResourceData) IntegrationData {
			return map[string]interface{}{
				"apiKey": d.Get("api_key"),
			}
		},
	}
	return builder.Build()
}

func resourceIntegrationGCPServiceAccountKeys() *schema.Resource {
	builder := ResourceIntegrationBuilder{
		Type: "gcp_service_account_keys",
		DataSchema: map[string]*schema.Schema{
			"impersonated_service_account": {
				Description: "The service account email of the account to be impersonated",
				Type:        schema.TypeString,
				Required:    true,
			},
			"external_id": {
				Description: "The Doppler-generated external id (placed in the service account description field)",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
		DataBuilder: func(d *schema.ResourceData) IntegrationData {
			return map[string]interface{}{
				"impersonatedServiceAccount": d.Get("impersonated_service_account"),
				"externalId":                 d.Get("external_id"),
			}
		},
	}
	return builder.Build()
}

func resourceIntegrationGCPCloudSQLMySQL() *schema.Resource {
	builder := ResourceIntegrationBuilder{
		Type: "gcp_cloudsql_mysql",
		DataSchema: map[string]*schema.Schema{
			"gcp_key": {
				Description: "The GCP service account key",
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
			},
		},
		DataBuilder: func(d *schema.ResourceData) IntegrationData {
			return map[string]interface{}{
				"gcpKey": d.Get("gcp_key"),
			}
		},
	}
	return builder.Build()
}

func resourceIntegrationGCPCloudSQLPostgres() *schema.Resource {
	builder := ResourceIntegrationBuilder{
		Type: "gcp_cloudsql_postgres",
		DataSchema: map[string]*schema.Schema{
			"gcp_key": {
				Description: "The GCP service account key",
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
			},
		},
		DataBuilder: func(d *schema.ResourceData) IntegrationData {
			return map[string]interface{}{
				"gcpKey": d.Get("gcp_key"),
			}
		},
	}
	return builder.Build()
}

func resourceIntegrationGCPCloudSQLSQLServer() *schema.Resource {
	builder := ResourceIntegrationBuilder{
		Type: "gcp_cloudsql_sqlserver",
		DataSchema: map[string]*schema.Schema{
			"gcp_key": {
				Description: "The GCP service account key",
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
			},
		},
		DataBuilder: func(d *schema.ResourceData) IntegrationData {
			return map[string]interface{}{
				"gcpKey": d.Get("gcp_key"),
			}
		},
	}
	return builder.Build()
}

func resourceIntegrationAWSIAMUserKeys() *schema.Resource {
	builder := ResourceIntegrationBuilder{
		Type: "aws_iam_user_keys",
		DataSchema: map[string]*schema.Schema{
			"assume_role_arn": {
				Description: "IAM Role ARN for role assumption",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
		DataBuilder: func(d *schema.ResourceData) IntegrationData {
			return map[string]interface{}{
				"aws_assume_role_arn": d.Get("assume_role_arn"),
			}
		},
	}
	return builder.Build()
}

func resourceIntegrationAWSMySQL() *schema.Resource {
	builder := ResourceIntegrationBuilder{
		Type: "aws_mysql",
		DataSchema: map[string]*schema.Schema{
			"assume_role_arn": {
				Description: "IAM Role ARN for role assumption",
				Type:        schema.TypeString,
				Required:    true,
			},
			"lambda_arn": {
				Description: "The Lambda ARN",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
		DataBuilder: func(d *schema.ResourceData) IntegrationData {
			return map[string]interface{}{
				"roleARN":   d.Get("assume_role_arn"),
				"lambdaARN": d.Get("lambda_arn"),
			}
		},
	}
	return builder.Build()
}

func resourceIntegrationAWSPostgres() *schema.Resource {
	builder := ResourceIntegrationBuilder{
		Type: "aws_postgres",
		DataSchema: map[string]*schema.Schema{
			"assume_role_arn": {
				Description: "IAM Role ARN for role assumption",
				Type:        schema.TypeString,
				Required:    true,
			},
			"lambda_arn": {
				Description: "The Lambda ARN",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
		DataBuilder: func(d *schema.ResourceData) IntegrationData {
			return map[string]interface{}{
				"roleARN":   d.Get("assume_role_arn"),
				"lambdaARN": d.Get("lambda_arn"),
			}
		},
	}
	return builder.Build()
}

func resourceIntegrationAWSMSSQLServer() *schema.Resource {
	builder := ResourceIntegrationBuilder{
		Type: "aws_mssql",
		DataSchema: map[string]*schema.Schema{
			"assume_role_arn": {
				Description: "IAM Role ARN for role assumption",
				Type:        schema.TypeString,
				Required:    true,
			},
			"lambda_arn": {
				Description: "The Lambda ARN",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
		DataBuilder: func(d *schema.ResourceData) IntegrationData {
			return map[string]interface{}{
				"roleARN":   d.Get("assume_role_arn"),
				"lambdaARN": d.Get("lambda_arn"),
			}
		},
	}
	return builder.Build()
}
