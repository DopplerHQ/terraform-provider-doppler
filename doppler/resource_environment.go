package doppler

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceEnvironment() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceEnvironmentCreate,
		ReadContext:   resourceEnvironmentRead,
		UpdateContext: resourceEnvironmentUpdate,
		DeleteContext: resourceEnvironmentDelete,
		Schema: map[string]*schema.Schema{
			"project": {
				Description: "The name of the Doppler project where the environment is located",
				Type:        schema.TypeString,
				Required:    true,
				// Environments cannot be moved directly from one project to another, they must be re-created
				ForceNew: true,
			},
			"slug": {
				Description: "The slug of the Doppler environment",
				Type:        schema.TypeString,
				Required:    true,
			},
			"name": {
				Description: "The name of the Doppler environment",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func resourceEnvironmentCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)

	var diags diag.Diagnostics
	project := d.Get("project").(string)
	slug := d.Get("slug").(string)
	name := d.Get("name").(string)

	environment, err := client.CreateEnvironment(ctx, project, slug, name)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(environment.getResourceId())

	return diags
}

func resourceEnvironmentUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)

	var diags diag.Diagnostics
	project, currentSlug, err := parseEnvironmentResourceId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	newName := d.Get("name").(string)
	newSlug := d.Get("slug").(string)

	environment, err := client.RenameEnvironment(ctx, project, currentSlug, newSlug, newName)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(environment.getResourceId())
	return diags
}

func resourceEnvironmentRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)

	var diags diag.Diagnostics
	project, slug, err := parseEnvironmentResourceId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	environment, err := client.GetEnvironment(ctx, project, slug)
	if err != nil {
		return diag.FromErr(err)
	}

	setErr := d.Set("slug", environment.Slug)
	if setErr != nil {
		return diag.FromErr(setErr)
	}

	setErr = d.Set("name", environment.Name)
	if setErr != nil {
		return diag.FromErr(setErr)
	}

	return diags
}

func resourceEnvironmentDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)

	var diags diag.Diagnostics
	project, slug, err := parseEnvironmentResourceId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.DeleteEnvironment(ctx, project, slug)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}
