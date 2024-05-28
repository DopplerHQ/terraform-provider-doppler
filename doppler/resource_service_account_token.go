package doppler

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceServiceAccountToken() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceServiceAccountTokenCreate,
		ReadContext:   resourceServiceAccountTokenRead,
		DeleteContext: resourceServiceAccountTokenDelete,
		// ForceNew is specified for all user-specified fields
		// Service account tokens cannot be moved, renamed, or edited to change their access
		Schema: map[string]*schema.Schema{
			"service_account_slug": {
				Description: "Slug of the service account",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Description: "The display name of the API token",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"slug": {
				Description: "Slug of the service account token",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"expires_at": {
				Description: "The datetime at which the API token should expire. " +
					"If not provided, the API token will remain valid indefinitely unless manually revoked",
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"created_at": {
				Description: "The datetime that the token was created.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"api_key": {
				Description: "The api key used to authenticate the service account",
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
			},
		},
	}
}

func resourceServiceAccountTokenCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)

	var diags diag.Diagnostics
	serviceAccount := d.Get("service_account_slug").(string)
	name := d.Get("name").(string)
	expiresAt, ok := d.Get("expires_at").(string)
	if !ok {
		expiresAt = ""
	}

	token, err := client.CreateServiceAccountToken(ctx, serviceAccount, name, expiresAt)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(token.ServiceAccountToken.getResourceId())

	if err = d.Set("expires_at", token.ServiceAccountToken.ExpiresAt); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("created_at", token.ServiceAccountToken.CreatedAt); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("api_key", token.ApiKey); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("slug", token.ServiceAccountToken.Slug); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceServiceAccountTokenRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)

	var diags diag.Diagnostics
	serviceAccount := d.Get("service_account_slug").(string)
	slug := d.Id()

	token, err := client.GetServiceAccountToken(ctx, serviceAccount, slug)
	if err != nil {
		return diag.FromErr(err)
	}

	if err = d.Set("name", token.ServiceAccountToken.Name); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("expires_at", token.ServiceAccountToken.ExpiresAt); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("created_at", token.ServiceAccountToken.CreatedAt); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("slug", token.ServiceAccountToken.Slug); err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceServiceAccountTokenDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)

	var diags diag.Diagnostics
	serviceAccount := d.Get("service_account_slug").(string)
	slug := d.Id()

	if err := client.DeleteServiceAccountToken(ctx, serviceAccount, slug); err != nil {
		return diag.FromErr(err)
	}
	return diags
}
