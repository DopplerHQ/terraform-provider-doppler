package doppler

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceGroupCreate,
		ReadContext:   resourceGroupRead,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		UpdateContext: resourceGroupUpdate,
		DeleteContext: resourceGroupDelete,
		Schema: map[string]*schema.Schema{
			"slug": {
				Description: "The slug of the group",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				Description: "The name of the group",
				Type:        schema.TypeString,
				Required:    true,
			},
			"default_project_role": {
				Description: "The default project role assigned to the group when added to a Doppler project",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
}

func resourceGroupCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)

	var diags diag.Diagnostics
	name := d.Get("name").(string)

	defaultProjectRole := ""
	rawDefaultProjectRole, defaultProjectRoleSet := d.Get("default_project_role").(string)
	if defaultProjectRoleSet {
		defaultProjectRole = rawDefaultProjectRole
	}

	group, err := client.CreateGroup(ctx, name, defaultProjectRole)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(group.Slug)

	err = updateGroupState(d, group)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceGroupUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)

	var diags diag.Diagnostics

	slug := d.Id()

	name := ""
	if d.HasChange("name") {
		name = d.Get("name").(string)
	}

	var defaultProjectRole *string
	if d.HasChange("default_project_role") {
		newDefaultProjectRole, newDefaultProjectRoleSet := d.Get("default_project_role").(string)
		if newDefaultProjectRoleSet {
			defaultProjectRole = &newDefaultProjectRole
		} else {
			// Empty string unsets the default project role
			empty := ""
			defaultProjectRole = &empty
		}
	}

	group, err := client.UpdateGroup(ctx, slug, name, defaultProjectRole)
	if err != nil {
		return diag.FromErr(err)
	}

	err = updateGroupState(d, group)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)

	var diags diag.Diagnostics

	slug := d.Id()

	group, err := client.GetGroup(ctx, slug)
	if err != nil {
		return handleNotFoundError(err, d)
	}

	err = updateGroupState(d, group)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func updateGroupState(d *schema.ResourceData, group *Group) error {
	if err := d.Set("slug", group.Slug); err != nil {
		return err
	}

	if err := d.Set("name", group.Name); err != nil {
		return err
	}

	if err := d.Set("default_project_role", group.DefaultProjectRole.Identifier); err != nil {
		return err
	}
	return nil
}

func resourceGroupDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)

	var diags diag.Diagnostics
	slug := d.Id()

	if err := client.DeleteGroup(ctx, slug); err != nil {
		return diag.FromErr(err)
	}

	return diags
}
