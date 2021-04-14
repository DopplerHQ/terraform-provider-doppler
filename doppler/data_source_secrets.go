package doppler

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceSecretsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	apiContext := m.(APIContext)

	format := d.Get("format").(string)

	result, err := GetSecrets(apiContext)
	if err != nil {
		return diag.FromErr(err)
	}

	secrets := make(map[string]string)

	for _, computedSecret := range result {
		var transformedValue string
		if format == "raw" {
			transformedValue = computedSecret.RawValue
		} else {
			transformedValue = computedSecret.ComputedValue
		}

		secrets[computedSecret.Name] = transformedValue
	}

	if err := d.Set("map", secrets); err != nil {
		return diag.FromErr(err)
	}

	// always run
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}

func dataSourceSecrets() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSecretsRead,
		Schema: map[string]*schema.Schema{
			"format": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default: "computed",
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := val.(string)
					if v != "raw" && v != "computed" {
						errs = append(errs, fmt.Errorf("%s must be either 'raw' or 'computed'"))
					}
					return
				},
			},
			"map": &schema.Schema{
				Type:      schema.TypeMap,
				Computed:  true,
				Sensitive: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}
