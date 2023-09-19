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
				Description: "The Doppler API host (i.e. https://api.doppler.com)",
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("DOPPLER_API_HOST", defaultAPIHost),
			},
			"verify_tls": {
				Description: "Whether or not to verify TLS",
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("DOPPLER_VERIFY_TLS", true),
			},
			"doppler_token": {
				Description: "A Doppler token, either a personal or service token",
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
			"doppler_trusted_ip":    resourceTrustedIP(),
			"doppler_service_token": resourceServiceToken(),

			"doppler_service_account": resourceServiceAccount(),

			"doppler_group": resourceGroup(),

			"doppler_project_member_group":           resourceProjectMemberGroup(),
			"doppler_project_member_service_account": resourceProjectMemberServiceAccount(),

			"doppler_integration_aws_secrets_manager":  resourceIntegrationAWSAssumeRoleIntegration("aws_secrets_manager"),
			"doppler_secrets_sync_aws_secrets_manager": resourceSyncAWSSecretsManager(),

			"doppler_integration_aws_parameter_store":  resourceIntegrationAWSAssumeRoleIntegration("aws_parameter_store"),
			"doppler_secrets_sync_aws_parameter_store": resourceSyncAWSParameterStore(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"doppler_secrets": dataSourceSecrets(),
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
