package doppler

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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

func resourceIntegrationOpenAI() *schema.Resource {
	builder := ResourceIntegrationBuilder{
		Type: "openai_service_account",
		DataSchema: map[string]*schema.Schema{
			"admin_key": {
				Description: "The OpenAI admin key",
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
			},
		},
		DataBuilder: func(d *schema.ResourceData) IntegrationData {
			return map[string]interface{}{
				"adminKey": d.Get("admin_key"),
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

func gcpOIDCDataSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"gcp_workload_identity_provider": {
			Description: "The full resource name of the workload identity pool provider Doppler federates with, e.g. `//iam.googleapis.com/projects/123456789/locations/global/workloadIdentityPools/my-pool/providers/my-provider`. Note that it uses the numeric project number, not the project ID.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"gcp_project_id": {
			Description: "The ID of the GCP project the integration operates on",
			Type:        schema.TypeString,
			Required:    true,
		},
		"federation_principal": {
			Description: "The IAM principal identifier to grant roles to in GCP for this integration",
			Type:        schema.TypeString,
			Computed:    true,
		},
	}
}

func gcpFederationComputedFields(d *schema.ResourceData, integ *Integration) error {
	f := integ.Federation
	if f == nil || f.Principal == "" {
		return fmt.Errorf("the Doppler API returned no federation setup values for this integration; verify the connection still uses keyless (OIDC) auth and that this Doppler version supports it")
	}
	return d.Set("federation_principal", f.Principal)
}

func azureFederationComputedFields(d *schema.ResourceData, integ *Integration) error {
	f := integ.Federation
	if f == nil || f.Issuer == "" || f.Subject == "" || f.Audience == "" {
		return fmt.Errorf("the Doppler API returned no federation setup values for this integration; verify the connection still uses keyless (OIDC) auth and that this Doppler version supports it")
	}
	if err := d.Set("federation_issuer", f.Issuer); err != nil {
		return err
	}
	if err := d.Set("federation_subject", f.Subject); err != nil {
		return err
	}
	return d.Set("federation_audience", f.Audience)
}

func resourceIntegrationGCPSecretManagerOIDC() *schema.Resource {
	dataSchema := gcpOIDCDataSchema()
	dataSchema["gcp_secret_prefix"] = &schema.Schema{
		Description: "The prefix added to any secret created by this integration in GCP. Cannot be changed after the integration is created.",
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
	}
	builder := ResourceIntegrationBuilder{
		Type:       "gcp_secret_manager",
		DataSchema: dataSchema,
		DataBuilder: func(d *schema.ResourceData) IntegrationData {
			data := map[string]interface{}{
				"authMethod":                     "oidc",
				"gcp_workload_identity_provider": d.Get("gcp_workload_identity_provider"),
				"gcp_project_id":                 d.Get("gcp_project_id"),
			}
			if prefix := d.Get("gcp_secret_prefix").(string); prefix != "" {
				data["gcp_secret_prefix"] = prefix
			}
			return data
		},
		ComputedFieldsFunc: gcpFederationComputedFields,
	}
	return builder.Build()
}

func resourceIntegrationGCPCloudSQLOIDC(integrationType string) *schema.Resource {
	builder := ResourceIntegrationBuilder{
		Type:       integrationType,
		DataSchema: gcpOIDCDataSchema(),
		DataBuilder: func(d *schema.ResourceData) IntegrationData {
			return map[string]interface{}{
				"authMethod":                     "oidc",
				"gcp_workload_identity_provider": d.Get("gcp_workload_identity_provider"),
				"gcp_project_id":                 d.Get("gcp_project_id"),
			}
		},
		ComputedFieldsFunc: gcpFederationComputedFields,
	}
	return builder.Build()
}

func resourceIntegrationAzureServicePrincipal(integrationType string) *schema.Resource {
	builder := ResourceIntegrationBuilder{
		Type: integrationType,
		DataSchema: map[string]*schema.Schema{
			"tenant_id": {
				Description:  "The Azure directory (tenant) ID",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsUUID,
			},
			"client_id": {
				Description:  "The application (client) ID of the managing service principal",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsUUID,
			},
			"client_secret": {
				Description: "The client secret of the managing service principal",
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
			},
		},
		DataBuilder: func(d *schema.ResourceData) IntegrationData {
			return map[string]interface{}{
				"tenantId":     d.Get("tenant_id"),
				"clientId":     d.Get("client_id"),
				"clientSecret": d.Get("client_secret"),
				"authMethod":   "clientSecret",
			}
		},
	}
	return builder.Build()
}

func resourceIntegrationAzureServicePrincipalOIDC(integrationType string) *schema.Resource {
	builder := ResourceIntegrationBuilder{
		Type: integrationType,
		DataSchema: map[string]*schema.Schema{
			"tenant_id": {
				Description:  "The Azure directory (tenant) ID",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsUUID,
			},
			"client_id": {
				Description:  "The application (client) ID of the managing service principal",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsUUID,
			},
			"federation_issuer": {
				Description: "The issuer of the OIDC tokens Doppler issues for this integration",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"federation_subject": {
				Description: "The subject of the OIDC tokens Doppler issues for this integration",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"federation_audience": {
				Description: "The audience to configure on the federated identity credential",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
		DataBuilder: func(d *schema.ResourceData) IntegrationData {
			return map[string]interface{}{
				"tenantId":   d.Get("tenant_id"),
				"clientId":   d.Get("client_id"),
				"authMethod": "oidc",
			}
		},
		ComputedFieldsFunc: azureFederationComputedFields,
	}
	return builder.Build()
}
