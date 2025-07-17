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
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
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
			"personal_configs": {
				Description: "Whether or not personal configs are enabled for the environment",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
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
	personalConfigs := d.Get("personal_configs").(bool)

	environment, err := client.CreateEnvironment(ctx, project, slug, name, personalConfigs)
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
	newPersonalConfigs := d.Get("personal_configs").(bool)

	environment, err := client.UpdateEnvironment(ctx, project, currentSlug, newSlug, newName, newPersonalConfigs)
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
		return handleNotFoundError(err, d)
	}

	if err = d.Set("slug", environment.Id); err != nil {
		return diag.FromErr(err)
	}

	if err = d.Set("name", environment.Name); err != nil {
		return diag.FromErr(err)
	}

	if err = d.Set("project", environment.Project); err != nil {
		return diag.FromErr(err)
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

	if err = client.DeleteEnvironment(ctx, project, slug); err != nil {
		return diag.FromErr(err)
	}

	return diags
}
