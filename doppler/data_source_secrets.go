package doppler

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceSecretsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(APIClient)

	d.SetId(client.GetId())
	project := d.Get("project").(string)
	config := d.Get("config").(string)

	result, err := client.GetComputedSecrets(ctx, project, config)
	if err != nil {
		return diag.FromErr(err)
	}

	secrets := make(map[string]string)

	for _, secret := range result {
		secrets[secret.Name] = secret.Value
	}

	if err := d.Set("map", secrets); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func dataSourceSecrets() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSecretsRead,
		Schema: map[string]*schema.Schema{
			"project": {
				Description: "The name of the Doppler project (required for personal tokens)",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
			},
			"config": {
				Description: "The name of the Doppler config (required for personal tokens)",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
			},
			"map": {
				Description: "A mapping of secret names to computed secret values",
				Type:        schema.TypeMap,
				Computed:    true,
				Sensitive:   true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}
