package doppler

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTrustedIP() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTrustedIPCreate,
		ReadContext:   resourceTrustedIPRead,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		DeleteContext: resourceTrustedIPDelete,
		Schema: map[string]*schema.Schema{
			"project": {
				Description: "The name of the Doppler project where the config is located",
				Type:        schema.TypeString,
				Required:    true,
				// Trusted IPs cannot be moved directly from one project to another, they must be re-created
				ForceNew: true,
			},
			"config": {
				Description: "The name of the Doppler config",
				Type:        schema.TypeString,
				Required:    true,
				// Trusted IPs cannot be moved directly from one config to another, they must be re-created
				ForceNew: true,
			},
			"ip": {
				Description: "The IP address or CIDR range to trust",
				Type:        schema.TypeString,
				Required:    true,
				// Trusted IP values cannot be updated, they must be re-created
				ForceNew: true,
			},
		},
	}
}

func resourceTrustedIPCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)

	var diags diag.Diagnostics
	project := d.Get("project").(string)
	config := d.Get("config").(string)
	ip := d.Get("ip").(string)

	added, err := client.AddTrustedIP(ctx, project, config, ip)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(added.getResourceId())

	return diags
}

func resourceTrustedIPRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)

	var diags diag.Diagnostics
	project, config, ip, err := parseTrustedIPResourceId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	ips, err := client.GetTrustedIPs(ctx, project, config)
	if err != nil {
		return handleNotFoundError(err, d)
	}

	var found *TrustedIP
	for _, i := range ips {
		if i.IP == ip {
			found = &i
			break
		}
	}

	if found == nil {
		return handleNotFoundError(err, d)
	}

	if err = d.Set("project", found.Project); err != nil {
		return diag.FromErr(err)
	}

	if err = d.Set("config", found.Config); err != nil {
		return diag.FromErr(err)
	}

	if err = d.Set("ip", found.IP); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceTrustedIPDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)

	var diags diag.Diagnostics
	project, config, ip, err := parseTrustedIPResourceId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if err = client.DeleteTrustedIP(ctx, project, config, ip); err != nil {
		return diag.FromErr(err)
	}

	return diags
}
