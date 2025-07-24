package doppler

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type RotatedSecretParametersBuilderFunc = func(d *schema.ResourceData) RotatedSecretParameters
type RotatedSecretCredentialsBuilderFunc = func(d *schema.ResourceData) RotatedSecretCredentials

type ResourceRotatedSecretBuilder struct {
	ParametersSchema   map[string]*schema.Schema
	ParametersBuilder  RotatedSecretParametersBuilderFunc
	CredentialsSchema  map[string]*schema.Schema
	CredentialsBuilder RotatedSecretCredentialsBuilderFunc
}

// resourceRotatedSecret returns a schema resource object for the RotatedSecret model.
func (builder ResourceRotatedSecretBuilder) Build() *schema.Resource {
	resourceSchema := map[string]*schema.Schema{
		"integration": {
			Description: "The slug of the integration to use for this rotated secret",
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
		"name": {
			Description: "The name of the rotated secret",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    false,
		},
		"rotation_period_sec": {
			Description: "How frequently to rotate the secret",
			Type:        schema.TypeInt,
			Required:    true,
			ForceNew:    false,
		},
	}

	// NOTE: We set ForceNew=true for parameters and credentials because only `name` and `rotation_period_sec` are allowed to be changed.
	// This could have been defined by the caller but it's error-prone to declare it many times when it can be set once here.

	for name, subschema := range builder.ParametersSchema {
		s := *subschema
		s.ForceNew = true
		resourceSchema[name] = &s
	}

	for name, subschema := range builder.CredentialsSchema {
		s := *subschema
		s.ForceNew = true
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

func (builder ResourceRotatedSecretBuilder) CreateContextFunc() schema.CreateContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		client := m.(APIClient)

		var diags diag.Diagnostics
		integ := d.Get("integration").(string)
		config := d.Get("config").(string)
		project := d.Get("project").(string)
		name := d.Get("name").(string)
		rotationPeriodSec := d.Get("rotation_period_sec").(int)
		parameters := map[string]interface{}{}
		credentials := []map[string]interface{}{}

		// Parameters and Credentials can both be optional for a rotated secret - it depends on what the server's expecting
		// for a specific type. We'll handle the Builders being nil to simplify the copypasta needed in implementing a specific one.
		if builder.ParametersBuilder != nil {
			parameters = builder.ParametersBuilder(d)
		}
		if builder.CredentialsBuilder != nil {
			credentials = builder.CredentialsBuilder(d)
		}

		rs, err := client.CreateRotatedSecret(ctx, name, rotationPeriodSec, parameters, credentials, config, project, integ)
		if err != nil {
			return diag.FromErr(err)
		}

		d.SetId(rs.Slug)

		return diags
	}
}

func (builder ResourceRotatedSecretBuilder) ReadContextFunc() schema.ReadContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		client := m.(APIClient)

		var diags diag.Diagnostics

		slug := d.Id()
		config := d.Get("config").(string)
		project := d.Get("project").(string)

		rs, err := client.GetRotatedSecret(ctx, config, project, slug)
		if err != nil {
			return handleNotFoundError(err, d)
		}

		if err = d.Set("integration", rs.Integration.Slug); err != nil {
			return diag.FromErr(err)
		}

		if err = d.Set("project", rs.Project); err != nil {
			return diag.FromErr(err)
		}

		if err = d.Set("config", rs.Config); err != nil {
			return diag.FromErr(err)
		}

		if err = d.Set("name", rs.Name); err != nil {
			return diag.FromErr(err)
		}

		if err = d.Set("rotation_period_sec", rs.RotationPeriodSec); err != nil {
			return diag.FromErr(err)
		}

		return diags
	}
}

func (builder ResourceRotatedSecretBuilder) UpdateContextFunc() schema.UpdateContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		client := m.(APIClient)

		var diags diag.Diagnostics
		slug := d.Id()
		config := d.Get("config").(string)
		project := d.Get("project").(string)
		name := d.Get("name").(string)
		rotationPeriodSec := d.Get("rotation_period_sec").(int)

		if d.HasChangesExcept("name", "rotation_period_sec") {
			// For good measure: This should not be possible because ForceNew is set for all other fields in the schema.
			return diag.Errorf("Only name and rotation_period_sec can be changed for rotated secrets")
		}

		_, err := client.UpdateRotatedSecret(ctx, name, rotationPeriodSec, config, project, slug)
		if err != nil {
			return diag.FromErr(err)
		}

		return diags
	}
}

func (builder ResourceRotatedSecretBuilder) DeleteContextFunc() schema.DeleteContextFunc {

	return func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		client := m.(APIClient)

		var diags diag.Diagnostics

		slug := d.Id()
		config := d.Get("config").(string)
		project := d.Get("project").(string)
		if err := client.DeleteRotatedSecret(ctx, slug, config, project); err != nil {
			return diag.FromErr(err)
		}

		return diags
	}
}
