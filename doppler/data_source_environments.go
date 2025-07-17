package doppler

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceEnvironmentsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(APIClient)

	d.SetId(client.GetId())
	project := d.Get("project").(string)

	environments, err := client.ListEnvironments(ctx, project)
	if err != nil {
		return diag.FromErr(err)
	}

	// Convert environments to a list of maps for Terraform
	var environmentsList []map[string]interface{}
	for _, env := range environments {
		environmentMap := map[string]interface{}{
			"slug":       env.Id,
			"name":       env.Name,
			"project":    env.Project,
			"created_at": env.CreatedAt,
		}
		environmentsList = append(environmentsList, environmentMap)
	}

	if err := d.Set("list", environmentsList); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func dataSourceEnvironments() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceEnvironmentsRead,
		Schema: map[string]*schema.Schema{
			"project": {
				Description: "The project to list environments for",
				Type:        schema.TypeString,
				Required:    true,
			},
			"list": {
				Description: "List of environments in the project",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"slug": {
							Description: "The slug of the environment",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"name": {
							Description: "The name of the environment",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"project": {
							Description: "The project the environment belongs to",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"created_at": {
							Description: "When the environment was created",
							Type:        schema.TypeString,
							Computed:    true,
						},
					},
				},
			},
		},
	}
}
