package doppler

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Builds a member-type-specific resource schema for a single role assignment using the configuration params.
type ResourceProjectMemberRoleBuilder struct {
	// The Doppler member type
	MemberType string

	// Any additional schema fields for the resource
	DataSchema map[string]*schema.Schema

	// A function which uses the resource data to return a member slug
	GetMemberSlugFunc ResourceProjectMemberGetMemberSlugFunc
}

func (builder ResourceProjectMemberRoleBuilder) Build() *schema.Resource {
	resourceSchema := map[string]*schema.Schema{
		"project": {
			Description: "The name of the Doppler project where the access is applied",
			Type:        schema.TypeString,
			Required:    true,
			// Access cannot be moved directly from one project to another, it must be re-created
			ForceNew: true,
		},
		"role": {
			Description: "The project role identifier for this assignment. Must use either one of the built-in project role slugs (`admin`, `collaborator`, `viewer`, or `no_access`), or the slug for a custom project role. A member holds each role at most once per project.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"environments": {
			Description: "The environments in the project where this role applies (null or omitted for roles with access to all environments)",
			Type:        schema.TypeSet,
			Optional:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
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

func (builder ResourceProjectMemberRoleBuilder) CreateContextFunc() schema.CreateContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		client := m.(APIClient)

		var diags diag.Diagnostics
		project := d.Get("project").(string)
		role := d.Get("role").(string)
		environments := getProjectMemberRoleEnvironments(d)

		memberSlug, err := builder.GetMemberSlugFunc(ctx, d, m)
		if err != nil {
			return diag.FromErr(err)
		}

		assignment, err := client.CreateProjectMemberRole(ctx, project, builder.MemberType, *memberSlug, role, environments)
		if err != nil {
			return diag.FromErr(err)
		}

		if err = updateProjectMemberRoleState(d, assignment); err != nil {
			return diag.FromErr(err)
		}

		d.SetId(getProjectMemberRoleId(project, builder.MemberType, *memberSlug, role))

		return diags
	}
}

func (builder ResourceProjectMemberRoleBuilder) UpdateContextFunc() schema.UpdateContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		client := m.(APIClient)

		var diags diag.Diagnostics

		project, memberType, memberSlug, role, err := parseProjectMemberRoleId(d.Id())
		if err != nil {
			return diag.FromErr(err)
		}

		assignment, err := client.UpdateProjectMemberRole(ctx, project, memberType, memberSlug, role, getProjectMemberRoleEnvironments(d))
		if err != nil {
			return diag.FromErr(err)
		}

		if err = updateProjectMemberRoleState(d, assignment); err != nil {
			return diag.FromErr(err)
		}

		return diags
	}
}

func (builder ResourceProjectMemberRoleBuilder) ReadContextFunc() schema.ReadContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		client := m.(APIClient)

		var diags diag.Diagnostics

		project, memberType, existingMemberSlug, role, err := parseProjectMemberRoleId(d.Id())
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
			d.SetId(getProjectMemberRoleId(project, builder.MemberType, memberSlug, role))
		}

		assignment, err := client.GetProjectMemberRole(ctx, project, memberType, memberSlug, role)
		if err != nil {
			return handleNotFoundError(err, d)
		}

		if err = updateProjectMemberRoleState(d, assignment); err != nil {
			return diag.FromErr(err)
		}

		return diags
	}
}

func (builder ResourceProjectMemberRoleBuilder) DeleteContextFunc() schema.DeleteContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		client := m.(APIClient)

		var diags diag.Diagnostics

		project, memberType, memberSlug, role, err := parseProjectMemberRoleId(d.Id())
		if err != nil {
			return diag.FromErr(err)
		}

		if err := client.DeleteProjectMemberRole(ctx, project, memberType, memberSlug, role); err != nil {
			return diag.FromErr(err)
		}

		return diags
	}
}

func getProjectMemberRoleEnvironments(d *schema.ResourceData) []string {
	rawEnvironments := d.Get("environments").(*schema.Set).List()
	environments := make([]string, len(rawEnvironments))
	for i, v := range rawEnvironments {
		environments[i] = v.(string)
	}
	return environments
}

func updateProjectMemberRoleState(d *schema.ResourceData, assignment *ProjectMemberRoleAssignment) error {
	return d.Set("environments", assignment.Environments)
}
