package doppler

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceConfig() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceConfigCreate,
		ReadContext:   resourceConfigRead,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		UpdateContext: resourceConfigUpdate,
		DeleteContext: resourceConfigDelete,
		Schema: map[string]*schema.Schema{
			"project": {
				Description: "The name of the Doppler project where the config is located",
				Type:        schema.TypeString,
				Required:    true,
				// Configs cannot be moved directly from one project to another, they must be re-created
				ForceNew: true,
			},
			"environment": {
				Description: "The name of the Doppler environment where the config is located",
				Type:        schema.TypeString,
				Required:    true,
				// Configs cannot be moved directly from one environment to another, they must be re-created
				ForceNew: true,
			},
			"name": {
				Description: "The name of the Doppler config",
				Type:        schema.TypeString,
				Required:    true,
			},
			"inheritable": {
				Description: "Whether or not the Doppler config can be inherited by other configs",
				Type:        schema.TypeBool,
				Optional:    true,
			},
		},
	}
}

func resourceConfigCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)

	var diags diag.Diagnostics
	project := d.Get("project").(string)
	environment := d.Get("environment").(string)
	name := d.Get("name").(string)
	inheritable := d.Get("inheritable").(bool)

	var config *Config
	var err error

	if name == environment {
		// By definition, root configs share the same name as their environment. If the user attempted to define
		// a resource for the root config (which would have required an environment to already be created), we
		// should just fetch the root config instead of attempting to create it, which would fail.
		config, err = client.GetConfig(ctx, project, name)
	} else {
		config, err = client.CreateConfig(ctx, project, environment, name)
	}

	if err != nil {
		return diag.FromErr(err)
	}

	if inheritable {
		// Configs are always created as not inheritable, and inheritability cannot be specified during the creation request.
		config, err = client.UpdateConfigInheritable(ctx, project, name, inheritable)

		if err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId(config.getResourceId())

	return diags
}

func resourceConfigUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)

	var diags diag.Diagnostics
	project, _, currentName, err := parseConfigResourceId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	newName := d.Get("name").(string)

	config, err := client.RenameConfig(ctx, project, currentName, newName)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(config.getResourceId())
	return diags
}

func resourceConfigRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)

	var diags diag.Diagnostics
	project, _, name, err := parseConfigResourceId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	config, err := client.GetConfig(ctx, project, name)
	if err != nil {
		return handleNotFoundError(err, d)
	}

	if err = d.Set("project", config.Project); err != nil {
		return diag.FromErr(err)
	}

	if err = d.Set("environment", config.Environment); err != nil {
		return diag.FromErr(err)
	}

	if err = d.Set("name", config.Name); err != nil {
		return diag.FromErr(err)
	}


	if err = d.Set("inheritable", config.Inheritable); err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceConfigDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)

	var diags diag.Diagnostics
	project, env, name, err := parseConfigResourceId(d.Id())
	if env == name {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "Root configs do not need to be manually deleted",
				Detail:   `Root configs are implicitly created/deleted along with their environments and cannot be manually deleted. Deleting the environment that contains this root config will result in the root config being deleted.`,
			},
		}
	}

	if err = client.DeleteConfig(ctx, project, name); err != nil {
		return diag.FromErr(err)
	}

	return diags
}
