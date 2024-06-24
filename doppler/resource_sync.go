package doppler

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

type SyncDataBuilderFunc = func(d *schema.ResourceData) SyncData

type ResourceSyncBuilder struct {
	DataSchema  map[string]*schema.Schema
	DataBuilder IntegrationDataBuilderFunc
}

// resourceSync returns a schema resource object for the Sync model.
func (builder ResourceSyncBuilder) Build() *schema.Resource {
	resourceSchema := map[string]*schema.Schema{
		"integration": {
			Description: "The slug of the integration to use for this sync",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"project": {
			Description: "The name of the Doppler project",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"config": {
			Description: "The name of the Doppler config",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"delete_behavior": {
			Description: "The behavior to be performed on the secrets in the sync target when this resource is deleted or recreated. Either `leave_in_target` (default) or `delete_from_target`.",
			Type:        schema.TypeString,
			Optional:    true,
			// Implicitly defaults to "leave_in_target" but not defined here to avoid state migration
			ValidateFunc: validation.StringInSlice([]string{"leave_in_target", "delete_from_target"}, false),
		},
	}

	for name, subschema := range builder.DataSchema {
		s := *subschema
		resourceSchema[name] = &s
	}

	return &schema.Resource{
		CreateContext: builder.CreateContextFunc(),
		ReadContext:   builder.ReadContextFunc(),
		UpdateContext: resourceSyncUpdate,
		DeleteContext: builder.DeleteContextFunc(),
		Schema:        resourceSchema,
	}
}

func (builder ResourceSyncBuilder) CreateContextFunc() schema.CreateContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		client := m.(APIClient)

		var diags diag.Diagnostics
		integ := d.Get("integration").(string)
		config := d.Get("config").(string)
		project := d.Get("project").(string)
		syncData := builder.DataBuilder(d)

		sync, err := client.CreateSync(ctx, syncData, config, project, integ)
		if err != nil {
			return diag.FromErr(err)
		}

		d.SetId(sync.Slug)

		return diags
	}
}

func (builder ResourceSyncBuilder) ReadContextFunc() schema.ReadContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		client := m.(APIClient)

		var diags diag.Diagnostics

		name := d.Id()
		config := d.Get("config").(string)
		project := d.Get("project").(string)

		sync, err := client.GetSync(ctx, config, project, name)
		if err != nil {
			return handleNotFoundError(err, d)
		}

		if err = d.Set("integration", sync.Integration); err != nil {
			return diag.FromErr(err)
		}

		if err = d.Set("project", sync.Project); err != nil {
			return diag.FromErr(err)
		}

		if err = d.Set("config", sync.Config); err != nil {
			return diag.FromErr(err)
		}

		return diags
	}
}

func resourceSyncUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	// This function must be specified in order to update `delete_behavior` but no API operations are required.
	// All other fields require `ForceNew`.
	return diags
}

func (builder ResourceSyncBuilder) DeleteContextFunc() schema.DeleteContextFunc {

	return func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		client := m.(APIClient)

		var diags diag.Diagnostics

		slug := d.Id()
		config := d.Get("config").(string)
		project := d.Get("project").(string)
		// NOTE: `delete_behavior` might be null, this logic will treat that as `leave_in_target`
		deleteFromTarget := d.Get("delete_behavior").(string) == "delete_from_target"
		if err := client.DeleteSync(ctx, slug, deleteFromTarget, config, project); err != nil {
			return diag.FromErr(err)
		}

		return diags
	}
}
