package doppler

import (
	"context"
	"fmt"
	"reflect"
	"strings"

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
			"descriptor": {
				Description: "The descriptor (project.config) of the Doppler config",
				Type:        schema.TypeString,
				Required:    false,
				Computed:    true,
			},
			"inheritable": {
				Description: "Whether or not the Doppler config can be inherited by other configs",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"inherits": {
				Description: "A list of other Doppler config descriptors that this config inherits from. Descriptors match the format \"project.config\" (e.g. backend.stg), which is most easily retrieved as the computed descriptor of a doppler_config resource (e.g. doppler_config.backend_stg.descriptor)",
				Optional:    true,
				Type:        schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func inheritsArgToDescriptors(inherits []interface{}) ([]ConfigDescriptor, error) {
	var descriptors []ConfigDescriptor

	for _, descriptor := range inherits {
		split := strings.Split(descriptor.(string), ".")
		if len(split) != 2 {
			return nil, fmt.Errorf("Unable to parse [%s] as descriptor", descriptor)
		}
		descriptors = append(descriptors, ConfigDescriptor{Project: split[0], Config: split[1]})
	}

	return descriptors, nil
}

func resourceConfigCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)

	var diags diag.Diagnostics
	project := d.Get("project").(string)
	environment := d.Get("environment").(string)
	name := d.Get("name").(string)
	inheritable := d.Get("inheritable").(bool)
	inherits := d.Get("inherits").([]interface{})

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

	if err = d.Set("descriptor", fmt.Sprintf("%s.%s", config.Project, config.Name)); err != nil {
		return diag.FromErr(err)
	}

	if config.Inheritable != inheritable {
		// Configs are always created as not inheritable, and inheritability cannot be specified during the creation request.
		config, err = client.UpdateConfigInheritable(ctx, project, name, inheritable)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	descriptors, nil := inheritsArgToDescriptors(inherits)
	if err != nil {
		return diag.FromErr(err)
	}
	if !reflect.DeepEqual(config.Inherits, descriptors) {
		config, err = client.UpdateConfigInherits(ctx, project, name, descriptors)
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

	if d.HasChange("name") {
		config, err := client.RenameConfig(ctx, project, currentName, newName)
		if err != nil {
			return diag.FromErr(err)
		}
		d.SetId(config.getResourceId())
		if err = d.Set("descriptor", fmt.Sprintf("%s.%s", config.Project, config.Name)); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("inheritable") {
		_, err = client.UpdateConfigInheritable(ctx, project, newName, d.Get("inheritable").(bool))
		if err != nil {
			oldValue, _ := d.GetChange("inheritable")
			err2 := d.Set("inheritable", oldValue)
			if err2 != nil {
				return diag.FromErr(err2)
			}
			return diag.FromErr(err)
		}
	}

	if d.HasChange("inherits") {
		inherits := d.Get("inherits").([]interface{})

		descriptors, nil := inheritsArgToDescriptors(inherits)
		if err != nil {
			return diag.FromErr(err)
		}

		_, err = client.UpdateConfigInherits(ctx, project, newName, descriptors)
		if err != nil {
			oldValue, _ := d.GetChange("inherits")
			err2 := d.Set("inherits", oldValue)
			if err2 != nil {
				return diag.FromErr(err2)
			}
			return diag.FromErr(err)
		}
	}

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

	if err = d.Set("descriptor", fmt.Sprintf("%s.%s", config.Project, config.Name)); err != nil {
		return diag.FromErr(err)
	}

	if err = d.Set("inheritable", config.Inheritable); err != nil {
		return diag.FromErr(err)
	}

	var descriptorsStrs []string

	for _, descriptor := range config.Inherits {
		descriptorsStrs = append(descriptorsStrs, fmt.Sprintf("%s.%s", descriptor.Project, descriptor.Config))
	}

	if err = d.Set("inherits", descriptorsStrs); err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceConfigDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)

	var diags diag.Diagnostics
	project, env, name, err := parseConfigResourceId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
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
