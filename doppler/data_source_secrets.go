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
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"config": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"map": {
				Type:      schema.TypeMap,
				Computed:  true,
				Sensitive: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}
