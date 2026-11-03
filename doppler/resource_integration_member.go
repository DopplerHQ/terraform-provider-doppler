package doppler

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type ResourceIntegrationMemberGetMemberSlugFunc = func(ctx context.Context, d *schema.ResourceData, m interface{}) (*string, error)

// Builds a member-type-specific resource schema using the configuration params.
type ResourceIntegrationMemberBuilder struct {
	// The Doppler member type
	MemberType string

	// Any additional schema fields for the resource
	DataSchema map[string]*schema.Schema

	// A function which uses the resource data to return a member slug
	GetMemberSlugFunc ResourceIntegrationMemberGetMemberSlugFunc
}

func (builder ResourceIntegrationMemberBuilder) Build() *schema.Resource {
	resourceSchema := map[string]*schema.Schema{
		"integration": {
			Description: "The slug of the Doppler integration where the access is applied",
			Type:        schema.TypeString,
			Required:    true,
			// Access cannot be moved directly from one integration to another, it must be re-created
			ForceNew: true,
		},
		"role": {
			Description: "The integration role identifier for the access. Must use either one of the built-in integration role slugs (`admin`, `consumer`, `viewer`, or `no_access`), or the slug for a custom integration role.",
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

func (builder ResourceIntegrationMemberBuilder) CreateContextFunc() schema.CreateContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		client := m.(APIClient)

		var diags diag.Diagnostics
		integration := d.Get("integration").(string)
		role := d.Get("role").(string)

		memberSlug, err := builder.GetMemberSlugFunc(ctx, d, m)
		if err != nil {
			return diag.FromErr(err)
		}

		member, err := client.CreateIntegrationMember(ctx, integration, builder.MemberType, *memberSlug, role)
		if err != nil {
			return diag.FromErr(err)
		}

		err = updateIntegrationMemberState(d, member)
		if err != nil {
			return diag.FromErr(err)
		}

		d.SetId(getIntegrationMemberId(integration, builder.MemberType, member.Slug))

		return diags
	}
}

func (builder ResourceIntegrationMemberBuilder) UpdateContextFunc() schema.UpdateContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		client := m.(APIClient)

		var diags diag.Diagnostics

		integration, memberType, memberSlug, err := parseIntegrationMemberId(d.Id())
		if err != nil {
			return diag.FromErr(err)
		}

		role := d.Get("role").(string)

		member, err := client.UpdateIntegrationMember(ctx, integration, memberType, memberSlug, role)
		if err != nil {
			return diag.FromErr(err)
		}

		err = updateIntegrationMemberState(d, member)
		if err != nil {
			return diag.FromErr(err)
		}

		return diags
	}
}

func (builder ResourceIntegrationMemberBuilder) ReadContextFunc() schema.ReadContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		client := m.(APIClient)

		var diags diag.Diagnostics

		integration, memberType, existingMemberSlug, err := parseIntegrationMemberId(d.Id())
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
			d.SetId(getIntegrationMemberId(integration, builder.MemberType, memberSlug))
		}

		member, err := client.GetIntegrationMember(ctx, integration, memberType, memberSlug)
		if err != nil {
			return handleNotFoundError(err, d)
		}

		err = updateIntegrationMemberState(d, member)
		if err != nil {
			return diag.FromErr(err)
		}

		return diags
	}
}

func updateIntegrationMemberState(d *schema.ResourceData, integrationMember *IntegrationMember) error {
	if err := d.Set("role", integrationMember.Role.Identifier); err != nil {
		return err
	}

	return nil
}

func (builder ResourceIntegrationMemberBuilder) DeleteContextFunc() schema.DeleteContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		client := m.(APIClient)

		var diags diag.Diagnostics

		integration, memberType, memberSlug, err := parseIntegrationMemberId(d.Id())
		if err != nil {
			return diag.FromErr(err)
		}

		if err := client.DeleteIntegrationMember(ctx, integration, memberType, memberSlug); err != nil {
			return diag.FromErr(err)
		}

		return diags
	}
}
