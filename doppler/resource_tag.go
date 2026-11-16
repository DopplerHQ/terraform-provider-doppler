package doppler

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTag() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTagCreate,
		ReadContext:   resourceTagRead,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		UpdateContext: resourceTagUpdate,
		DeleteContext: resourceTagDelete,
		Schema: map[string]*schema.Schema{
			"slug": {
				Description: "The slug of the tag. Generated from the name when omitted",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
			},
			"name": {
				Description: "The name of the tag",
				Type:        schema.TypeString,
				Required:    true,
			},
			"color": {
				Description: "The color of the tag. One of `gray` (default), `purple`, `blue`, `green`, `yellow`, `orange`, `red`, or `pink`",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "gray",
			},
		},
	}
}

func resourceTagCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)

	var diags diag.Diagnostics
	name := d.Get("name").(string)
	color := d.Get("color").(string)
	slug := d.Get("slug").(string)

	tag, err := client.CreateTag(ctx, name, color, slug)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(tag.Slug)

	err = updateTagState(d, tag)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceTagUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)

	var diags diag.Diagnostics
	slug := d.Id()
	name := d.Get("name").(string)
	color := d.Get("color").(string)

	tag, err := client.UpdateTag(ctx, slug, name, color)
	if err != nil {
		return diag.FromErr(err)
	}

	err = updateTagState(d, tag)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceTagRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)

	var diags diag.Diagnostics
	slug := d.Id()

	tag, err := client.GetTag(ctx, slug)
	if err != nil {
		return handleNotFoundError(err, d)
	}

	err = updateTagState(d, tag)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func updateTagState(d *schema.ResourceData, tag *Tag) error {
	if err := d.Set("slug", tag.Slug); err != nil {
		return err
	}

	if err := d.Set("name", tag.Name); err != nil {
		return err
	}

	if err := d.Set("color", tag.Color); err != nil {
		return err
	}
	return nil
}

func resourceTagDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)

	var diags diag.Diagnostics
	slug := d.Id()

	if err := client.DeleteTag(ctx, slug); err != nil {
		return diag.FromErr(err)
	}

	return diags
}
