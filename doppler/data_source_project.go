package doppler

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceProject() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceProjectRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the Doppler project",
				Type:        schema.TypeString,
				Required:    true,
			},
			"slug": {
				Description: "The slug of the Doppler project",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"description": {
				Description: "The description of the Doppler project",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"created_at": {
				Description: "When the project was created",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourceProjectRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(APIClient)

	name := d.Get("name").(string)

	project, err := client.GetProject(ctx, name)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(project.Slug)

	if err := d.Set("slug", project.Slug); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("description", project.Description); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("created_at", project.CreatedAt); err != nil {
		return diag.FromErr(err)
	}

	return diags
}
