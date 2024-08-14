package doppler

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceGroupMembers() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceGroupMembersCreate,
		ReadContext:   resourceGroupMembersRead,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		UpdateContext: resourceGroupMembersUpdate,
		DeleteContext: resourceGroupMembersDelete,
		Schema: map[string]*schema.Schema{
			"group_slug": {
				Description: "The slug of the group",
				Type:        schema.TypeString,
				Required:    true,
				// Members cannot be moved directly from one group to another, they must be re-created
				ForceNew: true,
			},
			"user_slugs": {
				Description: "A list of user slugs in the group",
				Type:        schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required: true,
			},
		},
	}
}

func resourceGroupMembersCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)
	var diags diag.Diagnostics

	groupSlug := d.Get("group_slug").(string)
	// Just fetch one member to see if any exist
	currentMembers, err := client.GetGroupMembers(ctx, groupSlug, PageOptions{Page: 1, PerPage: 1})
	if err != nil {
		return diag.FromErr(err)
	}

	if len(currentMembers) > 0 {
		diags = append(diags,
			diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "This group has existing members",
				Detail:   "This group has existing members. All group memberships have been overwritten by this resource.",
			})
	}

	diags = append(diags, resourceGroupMembersUpdate(ctx, d, m)...)

	d.SetId(groupSlug)

	return diags
}

func resourceGroupMembersUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)

	var diags diag.Diagnostics
	groupSlug := d.Get("group_slug").(string)
	userSlugs := d.Get("user_slugs").(*schema.Set).List()

	members := make([]GroupMember, len(userSlugs))
	for i, v := range userSlugs {
		members[i] = GroupMember{Type: "workplace_user", Slug: v.(string)}
	}

	err := client.ReplaceGroupMembers(ctx, groupSlug, members)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceGroupMembersRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)

	var diags diag.Diagnostics

	groupSlug := d.Id()

	perPage := 1000
	maxPages := 5

	members := []GroupMember{}

	for page := 1; page <= maxPages; page++ {
		pageMembers, err := client.GetGroupMembers(ctx, groupSlug, PageOptions{Page: page, PerPage: perPage})
		if err != nil {
			return handleNotFoundError(err, d)
		}
		members = append(members, pageMembers...)
		if len(pageMembers) < perPage {
			break
		} else if page == maxPages {
			return diag.Errorf("Exceeded max number of group members")
		}
	}

	userSlugs := []string{}
	for _, v := range members {
		if v.Type == "workplace_user" {
			userSlugs = append(userSlugs, v.Slug)
		} else {
			return diag.Errorf("Actor type %s is not supported by this plugin version", v.Type)
		}
	}

	if err := d.Set("group_slug", groupSlug); err != nil {
		return diag.FromErr((err))
	}

	if err := d.Set("user_slugs", userSlugs); err != nil {
		return diag.FromErr((err))
	}

	return diags
}

func resourceGroupMembersDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)

	var diags diag.Diagnostics
	groupSlug := d.Id()

	// Setting the members to an empty list effectively deletes the memberships
	if err := client.ReplaceGroupMembers(ctx, groupSlug, []GroupMember{}); err != nil {
		return diag.FromErr(err)
	}

	return diags
}
