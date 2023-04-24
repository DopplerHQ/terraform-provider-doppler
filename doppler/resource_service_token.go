package doppler

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceServiceToken() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceServiceTokenCreate,
		ReadContext:   resourceServiceTokenRead,
		DeleteContext: resourceServiceTokenDelete,
		// ForceNew is specified for all user-specified fields
		// Service tokens cannot be moved, renamed, or edited to change their access
		Schema: map[string]*schema.Schema{
			"project": {
				Description: "The name of the Doppler project where the service token is located",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"config": {
				Description: "The name of the Doppler config where the service token is located",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Description: "The name of the Doppler service token",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"access": {
				Description:  "The access level (read or read/write)",
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "read",
				ValidateFunc: validation.StringInSlice([]string{"read", "read/write"}, false),
				ForceNew:     true,
			},
			"key": {
				Description: "The key for the Doppler service token",
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
			},
		},
	}
}

func resourceServiceTokenCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)

	var diags diag.Diagnostics
	project := d.Get("project").(string)
	config := d.Get("config").(string)
	access := d.Get("access").(string)
	name := d.Get("name").(string)

	token, err := client.CreateServiceToken(ctx, project, config, access, name)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(token.getResourceId())

	if err = d.Set("key", token.Key); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceServiceTokenRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)

	var diags diag.Diagnostics
	project, config, slug, err := parseServiceTokenResourceId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	tokens, err := client.GetServiceTokens(ctx, project, config)
	if err != nil {
		return handleNotFoundError(err, d)
	}

	var token *ServiceToken
	for _, searchToken := range tokens {
		if searchToken.Slug == slug {
			tokenRef := searchToken
			token = &tokenRef
		}
	}

	if token == nil {
		err := &CustomNotFoundError{Message: "Could not find requested service token"}
		return handleNotFoundError(err, d)
	}

	if err = d.Set("project", token.Project); err != nil {
		return diag.FromErr(err)
	}

	if err = d.Set("config", token.Config); err != nil {
		return diag.FromErr(err)
	}

	if err = d.Set("access", token.Access); err != nil {
		return diag.FromErr(err)
	}

	// `key` cannot be read after initial creation

	return diags
}

func resourceServiceTokenDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)

	var diags diag.Diagnostics
	project, config, slug, err := parseServiceTokenResourceId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if err = client.DeleteServiceToken(ctx, project, config, slug); err != nil {
		return diag.FromErr(err)
	}

	return diags
}
