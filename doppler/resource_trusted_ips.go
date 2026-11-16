package doppler

import (
	"context"
	"fmt"
	"net"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func validateCIDR(val interface{}, key string) (warns []string, errs []error) {
	v := val.(string)
	if _, _, err := net.ParseCIDR(v); err != nil {
		errs = append(errs, fmt.Errorf("%q is not a valid CIDR range", v))
	}
	return
}

func resourceTrustedIPs() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTrustedIPsCreate,
		ReadContext:   resourceTrustedIPsRead,
		UpdateContext: resourceTrustedIPsUpdate,
		DeleteContext: resourceTrustedIPsDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"project": {
				Description: "The name of the Doppler project where the config is located",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"config": {
				Description: "The name of the Doppler config",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"trusted_ips": {
				Description: "List of trusted IP ranges in CIDR notation (e.g. 203.0.113.0/24, 1.2.3.4/32). Use 0.0.0.0/0 to allow all traffic.",
				Type:        schema.TypeSet,
				Required:    true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateCIDR,
				},
			},
		},
	}
}

func resourceTrustedIPsCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)
	var diags diag.Diagnostics

	project := d.Get("project").(string)
	config := d.Get("config").(string)

	currentIPs, err := client.GetTrustedIPs(ctx, project, config)
	if err != nil {
		return diag.FromErr(err)
	}

	isDefaultOnly := len(currentIPs) == 1 && currentIPs[0] == "0.0.0.0/0"
	if len(currentIPs) > 0 && !isDefaultOnly {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "This config has existing trusted IPs",
			Detail:   "This config has existing trusted IP entries. They will be overwritten by this resource.",
		})
	}

	d.SetId(getTrustedIPsResourceId(project, config))

	currentIPMap := make(map[string]bool)
	for _, ip := range currentIPs {
		currentIPMap[ip] = true
	}

	desiredIPs := d.Get("trusted_ips").(*schema.Set).List()
	desiredIPMap := make(map[string]bool)
	for _, ip := range desiredIPs {
		desiredIPMap[ip.(string)] = true
	}

	for ip := range desiredIPMap {
		if !currentIPMap[ip] {
			if err := client.AddTrustedIP(ctx, project, config, ip); err != nil {
				return append(diags, diag.FromErr(err)...)
			}
		}
	}

	for ip := range currentIPMap {
		if !desiredIPMap[ip] {
			if err := client.DeleteTrustedIP(ctx, project, config, ip); err != nil {
				return append(diags, diag.FromErr(err)...)
			}
		}
	}

	return diags
}

func resourceTrustedIPsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)
	var diags diag.Diagnostics

	project, config, err := parseTrustedIPsResourceId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	ips, err := client.GetTrustedIPs(ctx, project, config)
	if err != nil {
		return handleNotFoundError(err, d)
	}

	if err = d.Set("project", project); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("config", config); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("trusted_ips", ips); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceTrustedIPsUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)

	project, config, err := parseTrustedIPsResourceId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("trusted_ips") {
		oldRaw, newRaw := d.GetChange("trusted_ips")
		oldIPList := oldRaw.(*schema.Set).List()
		newIPList := newRaw.(*schema.Set).List()

		oldIPMap := make(map[string]bool)
		for _, ip := range oldIPList {
			oldIPMap[ip.(string)] = true
		}

		newIPMap := make(map[string]bool)
		for _, ip := range newIPList {
			newIPMap[ip.(string)] = true
		}

		for ip := range newIPMap {
			if !oldIPMap[ip] {
				if err := client.AddTrustedIP(ctx, project, config, ip); err != nil {
					return diag.FromErr(err)
				}
			}
		}

		for ip := range oldIPMap {
			if !newIPMap[ip] {
				if err := client.DeleteTrustedIP(ctx, project, config, ip); err != nil {
					return diag.FromErr(err)
				}
			}
		}
	}

	return nil
}

func resourceTrustedIPsDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)
	var diags diag.Diagnostics

	project, config, err := parseTrustedIPsResourceId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	currentIPs, err := client.GetTrustedIPs(ctx, project, config)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := client.AddTrustedIP(ctx, project, config, "0.0.0.0/0"); err != nil {
		return diag.FromErr(err)
	}

	for _, ip := range currentIPs {
		if ip == "0.0.0.0/0" {
			continue
		}
		if err := client.DeleteTrustedIP(ctx, project, config, ip); err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}
