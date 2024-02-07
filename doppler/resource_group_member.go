package doppler

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type ResourceGroupMemberGetMemberSlugFunc = func(ctx context.Context, d *schema.ResourceData, m interface{}) (*string, error)

// Builds a member-type-specific resource schema using the configuration params.
type ResourceGroupMemberBuilder struct {
	// The Doppler member type
	MemberType string

	// Any additional schema fields for the resource
	DataSchema map[string]*schema.Schema

	// A function which uses the resource data to return a member slug
	GetMemberSlugFunc ResourceGroupMemberGetMemberSlugFunc
}

func (builder ResourceGroupMemberBuilder) Build() *schema.Resource {
	resourceSchema := map[string]*schema.Schema{
		"group_slug": {
			Description: "The slug of the Doppler group",
			Type:        schema.TypeString,
			Required:    true,
			// Members cannot be moved directly from one group to another, they must be re-created
			ForceNew: true,
		},
	}

	for name, subschema := range builder.DataSchema {
		s := *subschema
		resourceSchema[name] = &s
	}

	return &schema.Resource{
		CreateContext: builder.CreateContextFunc(),
		ReadContext:   builder.ReadContextFunc(),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		DeleteContext: builder.DeleteContextFunc(),
		Schema:        resourceSchema,
	}
}

func (builder ResourceGroupMemberBuilder) CreateContextFunc() schema.CreateContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		client := m.(APIClient)

		var diags diag.Diagnostics
		group := d.Get("group_slug").(string)

		memberSlug, err := builder.GetMemberSlugFunc(ctx, d, m)
		if err != nil {
			return diag.FromErr(err)
		}

		err = client.CreateGroupMember(ctx, group, builder.MemberType, *memberSlug)
		if err != nil {
			return diag.FromErr(err)
		}

		d.SetId(getGroupMemberId(group, builder.MemberType, *memberSlug))

		return diags
	}
}

func (builder ResourceGroupMemberBuilder) ReadContextFunc() schema.ReadContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		client := m.(APIClient)

		var diags diag.Diagnostics

		group, memberType, memberSlug, err := parseGroupMemberId(d.Id())
		if err != nil {
			return diag.FromErr(err)
		}

		err = client.GetGroupMember(ctx, group, memberType, memberSlug)
		if err != nil {
			return handleNotFoundError(err, d)
		}

		return diags
	}
}

func (builder ResourceGroupMemberBuilder) DeleteContextFunc() schema.DeleteContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		client := m.(APIClient)

		var diags diag.Diagnostics

		group, memberType, memberSlug, err := parseGroupMemberId(d.Id())
		if err != nil {
			return diag.FromErr(err)
		}

		if err := client.DeleteGroupMember(ctx, group, memberType, memberSlug); err != nil {
			return diag.FromErr(err)
		}

		return diags
	}
}
