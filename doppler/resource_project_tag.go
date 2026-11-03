package doppler

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceProjectTag() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceProjectTagCreate,
		ReadContext:   resourceProjectTagRead,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		DeleteContext: resourceProjectTagDelete,
		Schema: map[string]*schema.Schema{
			"project": {
				Description: "The name of the Doppler project the tag is assigned to",
				Type:        schema.TypeString,
				Required:    true,
				// Tag assignments cannot be moved directly from one project to another, they must be re-created
				ForceNew: true,
			},
			"tag_slug": {
				Description: "The slug of the Doppler tag",
				Type:        schema.TypeString,
				Required:    true,
				// Tag assignments cannot be moved directly from one tag to another, they must be re-created
				ForceNew: true,
			},
		},
	}
}

func resourceProjectTagCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)

	var diags diag.Diagnostics
	project := d.Get("project").(string)
	tagSlug := d.Get("tag_slug").(string)

	if err := client.CreateProjectTag(ctx, project, tagSlug); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(getProjectTagId(project, tagSlug))

	return diags
}

func resourceProjectTagRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)

	var diags diag.Diagnostics
	project, tagSlug, err := parseProjectTagId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if err = client.GetProjectTag(ctx, project, tagSlug); err != nil {
		return handleNotFoundError(err, d)
	}

	if err = d.Set("project", project); err != nil {
		return diag.FromErr(err)
	}

	if err = d.Set("tag_slug", tagSlug); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceProjectTagDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)

	var diags diag.Diagnostics
	project, tagSlug, err := parseProjectTagId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if err = client.DeleteProjectTag(ctx, project, tagSlug); err != nil {
		return diag.FromErr(err)
	}

	return diags
}
