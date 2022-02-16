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
			"doppler_secret":      resourceSecret(),
			"doppler_project":     resourceProject(),
			"doppler_environment": resourceEnvironment(),
			"doppler_config":      resourceConfig(),
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
