package doppler

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
	}

	for name, subschema := range builder.DataSchema {
		s := *subschema
		resourceSchema[name] = &s
	}

	return &schema.Resource{
		CreateContext: builder.CreateContextFunc(),
		ReadContext:   builder.ReadContextFunc(),
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

func (builder ResourceSyncBuilder) DeleteContextFunc() schema.DeleteContextFunc {

	return func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		client := m.(APIClient)

		var diags diag.Diagnostics

		slug := d.Id()
		config := d.Get("config").(string)
		project := d.Get("project").(string)
		// In the future, we can support this as a param on the sync
		deleteFromTarget := false
		if err := client.DeleteSync(ctx, slug, deleteFromTarget, config, project); err != nil {
			return diag.FromErr(err)
		}

		return diags
	}
}
