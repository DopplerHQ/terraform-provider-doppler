package doppler

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(APIClient)

	var err error
	var result *Group
	slug := d.Get("slug")
	if slug != "" {
		result, err = client.GetGroup(ctx, d.Get("slug").(string))
		if err != nil {
			return diag.FromErr(err)
		}
	} else {
		result, err = client.GetGroupByName(ctx, d.Get("name").(string))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId(result.Slug)
	if err := d.Set("slug", result.Slug); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", result.Name); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func dataSourceGroup() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceGroupRead,
		Schema: map[string]*schema.Schema{
			"slug": {
				Description:  "The slug of the Doppler group",
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"slug", "name"},
			},
			"name": {
				Description:  "The name of the Doppler group",
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"slug", "name"},
			},
		},
	}
}
