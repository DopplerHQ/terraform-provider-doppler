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
			"host": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("DOPPLER_API_HOST", nil),
			},
			"verify_tls": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("DOPPLER_VERIFY_TLS", true),
			},
			"doppler_token": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("DOPPLER_TOKEN", nil),
			},
		},
		ResourcesMap: map[string]*schema.Resource{},
		DataSourcesMap: map[string]*schema.Resource{
			"doppler_secrets":         dataSourceSecrets(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	host := d.Get("host").(string)
	if host == "" {
		host = defaultAPIHost
	}
	verifyTLS := d.Get("verify_tls").(bool)
	token := d.Get("doppler_token").(string)

	var diags diag.Diagnostics

	return APIContext{Host: host, APIKey: token, VerifyTLS: verifyTLS}, diags
}
