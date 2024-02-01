package doppler

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceGroupMemberWorkplaceUser() *schema.Resource {
	builder := ResourceGroupMemberBuilder{
		MemberType: "workplace_user",
		DataSchema: map[string]*schema.Schema{
			"user": {
				Description: "The ID of the Doppler workplace user",
				Type:        schema.TypeString,
				Required:    true,
				// Members cannot be moved directly from one group to another, they must be re-created
				ForceNew: true,
			},
		},
		GetMemberSlugFunc: func(ctx context.Context, d *schema.ResourceData, m interface{}) (*string, error) {
			workplaceUserSlug := d.Get("user").(string)
			return &workplaceUserSlug, nil
		},
	}
	return builder.Build()
}
