package doppler

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const defaultAPIHost = "https://api.doppler.com"

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"host": {
				Description: "The Doppler API host (i.e. https://api.doppler.com). This can also be set via the DOPPLER_API_HOST environment variable.",
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("DOPPLER_API_HOST", defaultAPIHost),
			},
			"verify_tls": {
				Description: "Whether or not to verify TLS. This can also be set via the DOPPLER_VERIFY_TLS environment variable.",
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("DOPPLER_VERIFY_TLS", true),
			},
			"doppler_token": {
				Description: "A Doppler token, either a personal or service token. This can also be set via the DOPPLER_TOKEN environment variable.",
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("DOPPLER_TOKEN", nil),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"doppler_secret":        resourceSecret(),
			"doppler_project":       resourceProject(),
			"doppler_environment":   resourceEnvironment(),
			"doppler_config":        resourceConfig(),
			"doppler_service_token": resourceServiceToken(),

			"doppler_project_role": resourceProjectRole(),

			"doppler_workplace_role": resourceWorkplaceRole(),

			"doppler_service_account":       resourceServiceAccount(),
			"doppler_service_account_token": resourceServiceAccountToken(),

			"doppler_group":         resourceGroup(),
			"doppler_group_member":  resourceGroupMemberWorkplaceUser(),
			"doppler_group_members": resourceGroupMembers(),

			"doppler_webhook": resourceWebhook(),

			"doppler_change_request_policy": resourceChangeRequestPolicy(),

			"doppler_project_member_group":           resourceProjectMemberGroup(),
			"doppler_project_member_service_account": resourceProjectMemberServiceAccount(),
			"doppler_project_member_workplace_user":  resourceProjectMemberWorkplaceUser(),

			"doppler_integration_aws_secrets_manager":  resourceIntegrationAWSAssumeRoleIntegration("aws_secrets_manager"),
			"doppler_secrets_sync_aws_secrets_manager": resourceSyncAWSSecretsManager(),

			"doppler_integration_aws_parameter_store":  resourceIntegrationAWSAssumeRoleIntegration("aws_parameter_store"),
			"doppler_secrets_sync_aws_parameter_store": resourceSyncAWSParameterStore(),

			"doppler_integration_circleci":  resourceIntegrationCircleCi(),
			"doppler_secrets_sync_circleci": resourceSyncCircleCi(),

			"doppler_integration_terraform_cloud":  resourceIntegrationTerraformCloud(),
			"doppler_secrets_sync_terraform_cloud": resourceSyncTerraformCloud(),

			// creating integrations is not currently supported for GitHub syncs
			// "doppler_integration_github":  resourceIntegrationGitHub(),
			"doppler_secrets_sync_github_actions":    resourceSyncGitHubActions(),
			"doppler_secrets_sync_github_codespaces": resourceSyncGitHubCodespaces(),
			"doppler_secrets_sync_github_dependabot": resourceSyncGitHubDependabot(),

			"doppler_integration_flyio":  resourceIntegrationFlyio(),
			"doppler_secrets_sync_flyio": resourceSyncFlyio(),

			// creating Azure Vault oauth integrations is not currently supported
			// "doppler_integration_azure_vault":  resourceIntegrationAzureVault(),
			"doppler_integration_azure_vault_service_principal": resourceIntegrationAzureVaultServicePrincipal(),
			"doppler_secrets_sync_azure_vault":                  resourceSyncAzureVault(),

			"doppler_integration_gcp_secret_manager": resourceIntegrationGCPSecretManager(),
			"doppler_secrets_sync_gcp_secret_manager": resourceSyncGCPSecretManager(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"doppler_secrets": dataSourceSecrets(),
			"doppler_user":    dataSourceUser(),
			"doppler_group":   dataSourceGroup(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	host := d.Get("host").(string)
	verifyTLS := d.Get("verify_tls").(bool)
	token := d.Get("doppler_token").(string)

	var diags diag.Diagnostics

	return APIClient{Host: host, APIKey: token, VerifyTLS: verifyTLS}, diags
}
