package doppler

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceProjectMemberGroup() *schema.Resource {
	builder := ResourceProjectMemberBuilder{
		MemberType: "group",
		DataSchema: map[string]*schema.Schema{
			"group_slug": {
				Description: "The slug of the Doppler group",
				Type:        schema.TypeString,
				Required:    true,
				// Access cannot be moved directly from one group to another, it must be re-created
				ForceNew: true,
			},
		},
		GetMemberSlugFunc: func(ctx context.Context, d *schema.ResourceData, m interface{}) (*string, error) {
			groupSlug := d.Get("group_slug").(string)
			return &groupSlug, nil
		},
	}
	return builder.Build()
}

func resourceProjectMemberServiceAccount() *schema.Resource {
	builder := ResourceProjectMemberBuilder{
		MemberType: "service_account",
		DataSchema: map[string]*schema.Schema{
			"service_account_slug": {
				Description: "The slug of the Doppler service account",
				Type:        schema.TypeString,
				Required:    true,
				// Access cannot be moved directly from one service account to another, it must be re-created
				ForceNew: true,
			},
		},
		GetMemberSlugFunc: func(ctx context.Context, d *schema.ResourceData, m interface{}) (*string, error) {
			serviceAccountSlug := d.Get("service_account_slug").(string)
			return &serviceAccountSlug, nil
		},
	}
	return builder.Build()
}
