package doppler

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceSecretsObjectsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	apiContext := m.(APIContext)

	result, err := GetSecrets(apiContext)
	if err != nil {
		return diag.FromErr(err)
	}

	secrets := make([]map[string]string, 0)

	for _, computedSecret := range result {
		secrets = append(secrets, map[string]string{
			"name":     computedSecret.Name,
			"raw":      computedSecret.RawValue,
			"computed": computedSecret.ComputedValue,
		})
	}
	if err := d.Set("secrets", secrets); err != nil {
		return diag.FromErr(err)
	}

	// always run
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}

func dataSourceSecretsObjects() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSecretsObjectsRead,
		Schema: map[string]*schema.Schema{
			"secrets": &schema.Schema{
				Type:      schema.TypeList,
				Computed:  true,
				Sensitive: true,
				Elem: &schema.Schema{
					Type: schema.TypeMap,
					Elem: schema.TypeString,
				},
			},
		},
	}
}
