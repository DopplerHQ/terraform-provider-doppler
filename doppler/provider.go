package doppler

import (
	"context"
	"fmt"
	"os"
	"strings"

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
				Description: "A Doppler token, either a personal or service token. This can also be set via the DOPPLER_TOKEN environment variable. Only one of `doppler_token` or OIDC authentication (`oidc_identity` + `oidc_token`/`oidc_token_file`) may be specified.",
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("DOPPLER_TOKEN", nil),
			},
			"oidc_identity": {
				Description: "The identity ID or slug of the Doppler service account identity for OIDC authentication. This can also be set via the DOPPLER_OIDC_IDENTITY environment variable.",
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("DOPPLER_OIDC_IDENTITY", nil),
			},
			"oidc_token": {
				Description: "A JWT token to use for OIDC authentication. Only one of `oidc_token` or `oidc_token_file` may be set. This can also be set via the DOPPLER_OIDC_TOKEN environment variable.",
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("DOPPLER_OIDC_TOKEN", nil),
			},
			"oidc_token_file": {
				Description: "A path to a file containing a JWT token for OIDC authentication (e.g. a Kubernetes projected service account token). Only one of `oidc_token` or `oidc_token_file` may be set. This can also be set via the DOPPLER_OIDC_TOKEN_FILE environment variable.",
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("DOPPLER_OIDC_TOKEN_FILE", nil),
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

			"doppler_service_account":          resourceServiceAccount(),
			"doppler_service_account_token":    resourceServiceAccountToken(),
			"doppler_service_account_identity": resourceServiceAccountIdentity(),

			"doppler_group":         resourceGroup(),
			"doppler_group_member":  resourceGroupMemberWorkplaceUser(),
			"doppler_group_members": resourceGroupMembers(),

			"doppler_webhook": resourceWebhook(),

			"doppler_change_request_policy": resourceChangeRequestPolicy(),

			"doppler_project_member_group":           resourceProjectMemberGroup(),
			"doppler_project_member_service_account": resourceProjectMemberServiceAccount(),
			"doppler_project_member_user":            resourceProjectMemberUser(),

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

			"doppler_integration_twilio":                   resourceIntegrationTwilio(),
			"doppler_integration_cloudflare_tokens":        resourceIntegrationCloudflareTokens(),
			"doppler_integration_mongodb_atlas":            resourceIntegrationMongoDBAtlas(),
			"doppler_integration_sendgrid":                 resourceIntegrationSendGrid(),
			"doppler_integration_gcp_service_account_keys": resourceIntegrationGCPServiceAccountKeys(),
			"doppler_integration_gcp_cloudsql_mysql":       resourceIntegrationGCPCloudSQLMySQL(),
			"doppler_integration_gcp_cloudsql_postgres":    resourceIntegrationGCPCloudSQLPostgres(),
			"doppler_integration_gcp_cloudsql_sqlserver":   resourceIntegrationGCPCloudSQLSQLServer(),
			"doppler_integration_aws_iam_user_keys":        resourceIntegrationAWSIAMUserKeys(),
			"doppler_integration_aws_mysql":                resourceIntegrationAWSMySQL(),
			"doppler_integration_aws_mssql":                resourceIntegrationAWSMSSQLServer(),
			"doppler_integration_aws_postgres":             resourceIntegrationAWSPostgres(),

			"doppler_integration_external_id": resourceExternalId(),

			"doppler_rotated_secret_twilio":                   resourceRotatedSecretTwilio(),
			"doppler_rotated_secret_cloudflare_tokens":        resourceRotatedSecretCloudflareTokens(),
			"doppler_rotated_secret_mongodb_atlas":            resourceRotatedSecretMongoDBAtlas(),
			"doppler_rotated_secret_sendgrid":                 resourceRotatedSecretSendGrid(),
			"doppler_rotated_secret_gcp_cloudsql":             resourceRotatedSecretGCPCloudSQL(),
			"doppler_rotated_secret_aws_iam_user_keys":        resourceRotatedSecretAWSIAMUserKeys(),
			"doppler_rotated_secret_aws_mysql":                resourceRotatedSecretAWSMySQL(),
			"doppler_rotated_secret_aws_mssql":                resourceRotatedSecretAWSMSSQLServer(),
			"doppler_rotated_secret_aws_postgres":             resourceRotatedSecretAWSPostgres(),
			"doppler_rotated_secret_gcp_service_account_keys": resourceRotatedSecretGCPServiceAccountKeys(),

			// creating Azure Vault oauth integrations is not currently supported
			// "doppler_integration_azure_vault":  resourceIntegrationAzureVault(),
			"doppler_integration_azure_vault_service_principal": resourceIntegrationAzureVaultServicePrincipal(),
			"doppler_secrets_sync_azure_vault":                  resourceSyncAzureVault(),

			"doppler_integration_gcp_secret_manager":  resourceIntegrationGCPSecretManager(),
			"doppler_secrets_sync_gcp_secret_manager": resourceSyncGCPSecretManager(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"doppler_secrets":      dataSourceSecrets(),
			"doppler_user":         dataSourceUser(),
			"doppler_group":        dataSourceGroup(),
			"doppler_environments": dataSourceEnvironments(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	host := d.Get("host").(string)
	verifyTLS := d.Get("verify_tls").(bool)
	token := d.Get("doppler_token").(string)

	oidcIdentity := d.Get("oidc_identity").(string)
	oidcToken := d.Get("oidc_token").(string)
	oidcTokenFile := d.Get("oidc_token_file").(string)

	var diags diag.Diagnostics

	hasToken := token != ""
	hasOIDC := oidcIdentity != "" || oidcToken != "" || oidcTokenFile != ""

	if hasToken && hasOIDC {
		return nil, diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Conflicting authentication configuration",
				Detail:   "Only one of `doppler_token` or OIDC authentication (`oidc_identity` + `oidc_token`/`oidc_token_file`) may be specified.",
			},
		}
	}

	if !hasToken && !hasOIDC {
		return nil, diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Missing authentication configuration",
				Detail:   "Either `doppler_token` or OIDC authentication (`oidc_identity` + `oidc_token`/`oidc_token_file`) must be specified.",
			},
		}
	}

	if hasOIDC {
		if oidcIdentity == "" {
			return nil, diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Missing OIDC identity",
					Detail:   "`oidc_identity` must be specified when using OIDC authentication.",
				},
			}
		}

		if oidcToken != "" && oidcTokenFile != "" {
			return nil, diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Conflicting OIDC token configuration",
					Detail:   "Only one of `oidc_token` or `oidc_token_file` may be specified, not both.",
				},
			}
		}

		if oidcToken == "" && oidcTokenFile == "" {
			return nil, diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Missing OIDC token",
					Detail:   "One of `oidc_token` or `oidc_token_file` must be specified when using OIDC authentication.",
				},
			}
		}

		jwt := oidcToken
		if oidcTokenFile != "" {
			contents, err := os.ReadFile(oidcTokenFile)
			if err != nil {
				return nil, diag.Diagnostics{
					diag.Diagnostic{
						Severity: diag.Error,
						Summary:  "Unable to read OIDC token file",
						Detail:   fmt.Sprintf("Failed to read OIDC token from file %q: %s", oidcTokenFile, err),
					},
				}
			}
			jwt = strings.TrimSpace(string(contents))
			if jwt == "" {
				return nil, diag.Diagnostics{
					diag.Diagnostic{
						Severity: diag.Error,
						Summary:  "Empty OIDC token file",
						Detail:   fmt.Sprintf("OIDC token file %q exists but is empty.", oidcTokenFile),
					},
				}
			}
		}

		apiToken, err := exchangeOIDCToken(host, oidcIdentity, jwt, verifyTLS)
		if err != nil {
			return nil, diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "OIDC token exchange failed",
					Detail:   fmt.Sprintf("Failed to exchange OIDC token with Doppler: %s", err),
				},
			}
		}
		token = apiToken
	}

	return APIClient{Host: host, APIKey: token, VerifyTLS: verifyTLS}, diags
}
