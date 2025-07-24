package doppler

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceExternalId() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceExternalIdCreate,
		ReadContext:   resourceExternalIdRead,
		DeleteContext: resourceExternalIdDelete,
		Schema: map[string]*schema.Schema{
			"integration_type": {
				Description: "The integration type for the external id",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
		},
	}
}

func resourceExternalIdCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)

	var diags diag.Diagnostics
	integrationType := d.Get("integration_type").(string)

	externalId, err := client.CreateExternalId(ctx, integrationType)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(externalId)

	return diags
}

func resourceExternalIdRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// External IDs are not re-readable - you get one upon creating and are expected to use it
	// in the near future before it expires.
	return diag.Diagnostics{}
}

func resourceExternalIdDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// External IDs are not deletable - they expire on their own
	return diag.Diagnostics{
		diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "External IDs cannot be deleted manually, they automatically expire.",
			Detail:   `External IDs cannot be manually deleted. This operation just removes the external ID from Terraform state. The external ID will expire automatically.`,
		},
	}
}
