package doppler

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceProject() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceProjectCreate,
		ReadContext:   resourceProjectRead,
		UpdateContext: resourceProjectUpdate,
		DeleteContext: resourceProjectDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the Doppler project",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "The description of the Doppler project",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
}

func resourceProjectCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)

	var diags diag.Diagnostics
	name := d.Get("name").(string)
	description := d.Get("description").(string)

	project, err := client.CreateProject(ctx, name, description)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(project.Slug)

	return diags
}

func resourceProjectUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)

	var diags diag.Diagnostics
	currentName := d.Id()
	newName := d.Get("name").(string)
	description := d.Get("description").(string)

	project, err := client.UpdateProject(ctx, currentName, newName, description)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(project.Slug)
	return diags
}

func resourceProjectRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)

	var diags diag.Diagnostics

	name := d.Id()

	project, err := client.GetProject(ctx, name)
	if err != nil {
		return diag.FromErr(err)
	}

	setErr := d.Set("name", project.Name)
	if setErr != nil {
		return diag.FromErr(setErr)
	}

	setErr = d.Set("description", project.Description)
	if setErr != nil {
		return diag.FromErr(setErr)
	}

	return diags
}

func resourceProjectDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)

	var diags diag.Diagnostics

	name := d.Id()
	err := client.DeleteProject(ctx, name)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}
