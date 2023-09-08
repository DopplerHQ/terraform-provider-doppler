package doppler

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceServiceAccount() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceServiceAccountCreate,
		ReadContext:   resourceServiceAccountRead,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		UpdateContext: resourceServiceAccountUpdate,
		DeleteContext: resourceServiceAccountDelete,
		Schema: map[string]*schema.Schema{
			"slug": {
				Description: "The slug of the service account",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				Description: "The name of the service account",
				Type:        schema.TypeString,
				Required:    true,
			},
			"workplace_role": {
				Description:  "The identifier of the workplace role for the service account (or use `workplace_permissions`)",
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "",
				ExactlyOneOf: []string{"workplace_role", "workplace_permissions"},
			},
			"workplace_permissions": {
				Description: "A list of the workplace permissions for the service account (or use `workplace_role`)",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceServiceAccountCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)

	var diags diag.Diagnostics
	name := d.Get("name").(string)

	workplaceRole := ""
	rawWorkplaceRole, workplaceRoleSet := d.Get("workplace_role").(string)
	if workplaceRoleSet {
		workplaceRole = rawWorkplaceRole
	}

	var workplacePermissions []string
	rawPermissions := d.Get("workplace_permissions").([]interface{})
	if rawPermissions != nil {
		workplacePermissions = []string{}
		for _, v := range rawPermissions {
			workplacePermissions = append(workplacePermissions, v.(string))
		}
	}

	serviceAccount, err := client.CreateServiceAccount(ctx, name, workplaceRole, workplacePermissions)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(serviceAccount.Slug)

	err = updateServiceAccountState(d, serviceAccount)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceServiceAccountUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)

	var diags diag.Diagnostics

	slug := d.Id()

	name := ""
	if d.HasChange("name") {
		name = d.Get("name").(string)
	}

	workplaceRole := ""
	if d.HasChange("workplace_role") {
		newWorkplaceRole, newWorkplaceRoleSet := d.Get("workplace_role").(string)
		if newWorkplaceRoleSet {
			workplaceRole = newWorkplaceRole
		}
	}

	var workplacePermissions []string
	if d.HasChange("workplace_permissions") {
		rawPermissions := d.Get("workplace_permissions").([]interface{})
		if rawPermissions != nil {
			workplacePermissions = []string{}
			for _, v := range rawPermissions {
				workplacePermissions = append(workplacePermissions, v.(string))
			}
		}
	}

	serviceAccount, err := client.UpdateServiceAccount(ctx, slug, name, workplaceRole, workplacePermissions)
	if err != nil {
		return diag.FromErr(err)
	}

	err = updateServiceAccountState(d, serviceAccount)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceServiceAccountRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)

	var diags diag.Diagnostics

	slug := d.Id()

	serviceAccount, err := client.GetServiceAccount(ctx, slug)
	if err != nil {
		return handleNotFoundError(err, d)
	}

	err = updateServiceAccountState(d, serviceAccount)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func updateServiceAccountState(d *schema.ResourceData, serviceAccount *ServiceAccount) error {
	if err := d.Set("slug", serviceAccount.Slug); err != nil {
		return err
	}

	if err := d.Set("name", serviceAccount.Name); err != nil {
		return err
	}

	if !serviceAccount.WorkplaceRole.IsInlineRole {
		if err := d.Set("workplace_role", serviceAccount.WorkplaceRole.Identifier); err != nil {
			return err
		}
		if err := d.Set("workplace_permissions", nil); err != nil {
			return err
		}
	} else {
		if err := d.Set("workplace_role", ""); err != nil {
			return err
		}
		if err := d.Set("workplace_permissions", serviceAccount.WorkplaceRole.Permissions); err != nil {
			return err
		}
	}
	return nil
}

func resourceServiceAccountDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)

	var diags diag.Diagnostics
	slug := d.Id()

	if err := client.DeleteServiceAccount(ctx, slug); err != nil {
		return diag.FromErr(err)
	}

	return diags
}
