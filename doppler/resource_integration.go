package doppler

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type IntegrationDataBuilderFunc = func(d *schema.ResourceData) IntegrationData

type ResourceIntegrationBuilder struct {
	Type        string
	DataSchema  map[string]*schema.Schema
	DataBuilder IntegrationDataBuilderFunc
}

// resourceIntegration returns a schema resource object for the integration model.
func (builder ResourceIntegrationBuilder) Build() *schema.Resource {
	resourceSchema := map[string]*schema.Schema{
		"name": {
			Description: "The name of the integration",
			Type:        schema.TypeString,
			Required:    true,
		},
	}

	for name, subschema := range builder.DataSchema {
		s := *subschema
		resourceSchema[name] = &s
	}

	return &schema.Resource{
		CreateContext: builder.CreateContextFunc(),
		ReadContext:   builder.ReadContextFunc(),
		UpdateContext: builder.UpdateContextFunc(),
		DeleteContext: builder.DeleteContextFunc(),
		Schema:        resourceSchema,
	}
}

func (builder ResourceIntegrationBuilder) CreateContextFunc() schema.CreateContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		client := m.(APIClient)

		var diags diag.Diagnostics
		name := d.Get("name").(string)
		integData := builder.DataBuilder(d)

		integ, err := client.CreateIntegration(ctx, integData, name, builder.Type)
		if err != nil {
			return diag.FromErr(err)
		}

		d.SetId(integ.Slug)

		return diags
	}
}

func (builder ResourceIntegrationBuilder) UpdateContextFunc() schema.UpdateContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		client := m.(APIClient)

		var diags diag.Diagnostics
		slug := d.Id()
		name := ""
		if d.HasChange("name") {
			name = d.Get("name").(string)
		}

		var data IntegrationData = nil
		hasAnyDataFieldChanged := false

		for key := range builder.DataSchema {
			if d.HasChange(key) {
				hasAnyDataFieldChanged = true
			}
		}

		if hasAnyDataFieldChanged {
			data = builder.DataBuilder(d)
		}

		_, err := client.UpdateIntegration(ctx, slug, name, data)
		if err != nil {
			return diag.FromErr(err)
		}
		return diags
	}
}

func (builder ResourceIntegrationBuilder) ReadContextFunc() schema.ReadContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		client := m.(APIClient)

		var diags diag.Diagnostics

		slug := d.Id()

		integ, err := client.GetIntegration(ctx, slug)
		if err != nil {
			return handleNotFoundError(err, d)
		}

		if err = d.Set("name", integ.Name); err != nil {
			return diag.FromErr(err)
		}

		return diags
	}
}

func (builder ResourceIntegrationBuilder) DeleteContextFunc() schema.DeleteContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		client := m.(APIClient)

		var diags diag.Diagnostics

		slug := d.Id()
		if err := client.DeleteIntegration(ctx, slug); err != nil {
			return diag.FromErr(err)
		}

		return diags
	}
}
