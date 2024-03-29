package doppler

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(APIClient)

	result, err := client.GetWorkplaceUser(ctx, d.Get("email").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(result.Slug)
	if err := d.Set("slug", result.Slug); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func dataSourceUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceUserRead,
		Schema: map[string]*schema.Schema{
			"slug": {
				Description: "The slug of the Doppler user",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"email": {
				Description: "The email address of the Doppler user",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
