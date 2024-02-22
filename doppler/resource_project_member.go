package doppler

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type ResourceProjectMemberGetMemberSlugFunc = func(ctx context.Context, d *schema.ResourceData, m interface{}) (*string, error)

// Builds a member-type-specific resource schema using the configuration params.
type ResourceProjectMemberBuilder struct {
	// The Doppler member type
	MemberType string

	// Any additional schema fields for the resource
	DataSchema map[string]*schema.Schema

	// A function which uses the resource data to return a member slug
	GetMemberSlugFunc ResourceProjectMemberGetMemberSlugFunc
}

func (builder ResourceProjectMemberBuilder) Build() *schema.Resource {
	resourceSchema := map[string]*schema.Schema{
		"project": {
			Description: "The name of the Doppler project where the access is applied",
			Type:        schema.TypeString,
			Required:    true,
			// Access cannot be moved directly from one project to another, it must be re-created
			ForceNew: true,
		},
		"role": {
			Description: "The project role identifier for the access",
			Type:        schema.TypeString,
			Required:    true,
		},
		"environments": {
			Description: "The environments in the project where this access will apply (null or omitted for roles with access to all environments)",
			Type:        schema.TypeSet,
			Optional:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
	}

	v0ResourceSchema := map[string]*schema.Schema{
		"project": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"role": {
			Type:     schema.TypeString,
			Required: true,
		},
		"environments": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
	}

	for name, subschema := range builder.DataSchema {
		s := *subschema
		resourceSchema[name] = &s
		v0ResourceSchema[name] = &s
	}

	v0Resource := &schema.Resource{Schema: v0ResourceSchema}

	return &schema.Resource{
		CreateContext: builder.CreateContextFunc(),
		ReadContext:   builder.ReadContextFunc(),
		UpdateContext: builder.UpdateContextFunc(),
		DeleteContext: builder.DeleteContextFunc(),
		Schema:        resourceSchema,
		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    v0Resource.CoreConfigSchema().ImpliedType(),
				Upgrade: resourceUpgradeV0,
				Version: 0,
			},
		},
	}
}

func resourceUpgradeV0(ctx context.Context, rawState map[string]any, meta any) (map[string]any, error) {
	// migrate from TypeList to TypeSet so that order does not matter
	rawState["environments"] = rawState["environments"].([]interface{})

	return rawState, nil
}

func (builder ResourceProjectMemberBuilder) CreateContextFunc() schema.CreateContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		client := m.(APIClient)

		var diags diag.Diagnostics
		project := d.Get("project").(string)
		role := d.Get("role").(string)

		rawEnvironments := d.Get("environments").(*schema.Set).List()
		environments := make([]string, len(rawEnvironments))
		for i, v := range rawEnvironments {
			environments[i] = v.(string)
		}

		memberSlug, err := builder.GetMemberSlugFunc(ctx, d, m)
		if err != nil {
			return diag.FromErr(err)
		}

		member, err := client.CreateProjectMember(ctx, project, builder.MemberType, *memberSlug, role, environments)
		if err != nil {
			return diag.FromErr(err)
		}

		err = updateProjectMemberState(d, member)
		if err != nil {
			return diag.FromErr(err)
		}

		d.SetId(getProjectMemberId(project, builder.MemberType, member.Slug))

		return diags
	}
}

func (builder ResourceProjectMemberBuilder) UpdateContextFunc() schema.UpdateContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		client := m.(APIClient)

		var diags diag.Diagnostics

		project, memberType, memberSlug, err := parseProjectMemberId(d.Id())
		if err != nil {
			return diag.FromErr(err)
		}

		var role *string
		if d.HasChange("role") {
			newRole := d.Get("role").(string)
			role = &newRole
		}

		var environments []string
		if d.HasChange("environments") {
			rawEnvironments := d.Get("environments").(*schema.Set).List()
			environments = make([]string, len(rawEnvironments))
			for i, v := range rawEnvironments {
				environments[i] = v.(string)
			}
		}

		member, err := client.UpdateProjectMember(ctx, project, memberType, memberSlug, role, environments)
		if err != nil {
			return diag.FromErr(err)
		}

		err = updateProjectMemberState(d, member)
		if err != nil {
			return diag.FromErr(err)
		}

		return diags
	}
}

func (builder ResourceProjectMemberBuilder) ReadContextFunc() schema.ReadContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		client := m.(APIClient)

		var diags diag.Diagnostics

		project, memberType, existingMemberSlug, err := parseProjectMemberId(d.Id())
		if err != nil {
			return diag.FromErr(err)
		}

		checkedMemberSlug, err := builder.GetMemberSlugFunc(ctx, d, m)
		if err != nil {
			return diag.FromErr(err)
		}
		memberSlug := *checkedMemberSlug

		if memberSlug != existingMemberSlug {
			// Because resources can identify members using mutable fields (e.g. email),
			// we need to recheck the member to make sure that it still exists and it's still the same underlying member.
			d.SetId(getProjectMemberId(project, builder.MemberType, memberSlug))
		}

		member, err := client.GetProjectMember(ctx, project, memberType, memberSlug)
		if err != nil {
			return handleNotFoundError(err, d)
		}

		err = updateProjectMemberState(d, member)
		if err != nil {
			return diag.FromErr(err)
		}

		return diags
	}
}

func updateProjectMemberState(d *schema.ResourceData, projectMember *ProjectMember) error {
	if err := d.Set("role", projectMember.Role.Identifier); err != nil {
		return err
	}

	if err := d.Set("environments", projectMember.Environments); err != nil {
		return err
	}

	return nil
}

func (builder ResourceProjectMemberBuilder) DeleteContextFunc() schema.DeleteContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		client := m.(APIClient)

		var diags diag.Diagnostics

		project, memberType, memberSlug, err := parseProjectMemberId(d.Id())
		if err != nil {
			return diag.FromErr(err)
		}

		if err := client.DeleteProjectMember(ctx, project, memberType, memberSlug); err != nil {
			return diag.FromErr(err)
		}

		return diags
	}
}
