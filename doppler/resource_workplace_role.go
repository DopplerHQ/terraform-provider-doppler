package doppler

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceWorkplaceRole() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWorkplaceRoleCreate,
		ReadContext:   resourceWorkplaceRoleRead,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		UpdateContext: resourceWorkplaceRoleUpdate,
		DeleteContext: resourceWorkplaceRoleDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the Doppler workplace role",
				Type:        schema.TypeString,
				Required:    true,
			},
			"permissions": {
				Description: "A list of [Doppler workplace permissions](https://docs.doppler.com/reference/workplace_roles-create)",
				Type:        schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required: true,
			},
			"identifier": {
				Description: "The role's unique identifier",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"is_custom_role": {
				Description: "Whether or not the role is custom (as opposed to Doppler built-in)",
				Type:        schema.TypeBool,
				Computed:    true,
			},
		},
	}
}

func updateWorkplaceRoleData(d *schema.ResourceData, role *WorkplaceRole) diag.Diagnostics {
	d.SetId(role.Identifier)

	if err := d.Set("name", role.Name); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("permissions", role.Permissions); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("identifier", role.Identifier); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("is_custom_role", role.IsCustomRole); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceWorkplaceRoleCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)

	var diags diag.Diagnostics
	name := d.Get("name").(string)
	permissions := []string{}
	for _, v := range d.Get("permissions").(*schema.Set).List() {
		permissions = append(permissions, v.(string))
	}

	role, err := client.CreateWorkplaceRole(ctx, name, permissions)
	if err != nil {
		return diag.FromErr(err)
	}

	diags = append(diags, updateWorkplaceRoleData(d, role)...)
	return diags
}

func resourceWorkplaceRoleUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)

	var diags diag.Diagnostics
	identifier := d.Id()
	newName := d.Get("name").(string)
	newPermissions := []string{}
	for _, v := range d.Get("permissions").(*schema.Set).List() {
		newPermissions = append(newPermissions, v.(string))
	}

	role, err := client.UpdateWorkplaceRole(ctx, identifier, newName, newPermissions)
	if err != nil {
		return diag.FromErr(err)
	}
	diags = append(diags, updateWorkplaceRoleData(d, role)...)
	return diags
}

func resourceWorkplaceRoleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)

	var diags diag.Diagnostics

	identifier := d.Id()

	role, err := client.GetWorkplaceRole(ctx, identifier)
	if err != nil {
		return handleNotFoundError(err, d)
	}

	diags = append(diags, updateWorkplaceRoleData(d, role)...)
	return diags
}

func resourceWorkplaceRoleDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)

	var diags diag.Diagnostics

	identifier := d.Id()
	if err := client.DeleteWorkplaceRole(ctx, identifier); err != nil {
		return diag.FromErr(err)
	}

	return diags
}
