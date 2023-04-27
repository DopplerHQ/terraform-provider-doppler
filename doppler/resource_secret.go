package doppler

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceSecret() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSecretUpdate,
		ReadContext:   resourceSecretRead,
		UpdateContext: resourceSecretUpdate,
		DeleteContext: resourceSecretDelete,
		Schema: map[string]*schema.Schema{
			"project": {
				Description: "The name of the Doppler project",
				Type:        schema.TypeString,
				Required:    true,
			},
			"config": {
				Description: "The name of the Doppler config",
				Type:        schema.TypeString,
				Required:    true,
			},
			"name": {
				Description: "The name of the Doppler secret",
				Type:        schema.TypeString,
				Required:    true,
			},
			"value": {
				Description: "The raw secret value",
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
			},
			"computed": {
				Description: "The computed secret value, after resolving secret references",
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
			},
		},
		CustomizeDiff: customdiff.ComputedIf("computed", func(ctx context.Context, d *schema.ResourceDiff, meta interface{}) bool {
			return d.HasChange("value")
		}),
	}
}

func resourceSecretUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)

	var diags diag.Diagnostics
	project := d.Get("project").(string)
	config := d.Get("config").(string)
	name := d.Get("name").(string)
	value := d.Get("value").(string)

	secrets := []RawSecret{{Name: name, Value: &value}}
	if !d.IsNewResource() && d.HasChange("name") {
		previousName, _ := d.GetChange("name")
		secrets = append(secrets, RawSecret{Name: previousName.(string), Value: nil})
	}

	if err := client.UpdateSecrets(ctx, project, config, secrets); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(getSecretId(project, config, name))

	readDiags := resourceSecretRead(ctx, d, m)
	diags = append(diags, readDiags...)
	return diags
}

func resourceSecretRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)

	var diags diag.Diagnostics

	secretId := d.Id()
	project, config, name, err := parseSecretId(secretId)
	if err != nil {
		return diag.FromErr(err)
	}

	secret, err := client.GetSecret(ctx, project, config, name)
	if err != nil {
		return handleNotFoundError(err, d)
	}

	if err = d.Set("value", secret.Value.Raw); err != nil {
		return diag.FromErr(err)
	}

	if err = d.Set("computed", secret.Value.Computed); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceSecretDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)

	var diags diag.Diagnostics

	secretId := d.Id()
	tokens := strings.Split(secretId, ".")
	if len(tokens) != 3 {
		return diag.Errorf("Invalid secretId")
	}

	project := tokens[0]
	config := tokens[1]
	name := tokens[2]

	newSecret := RawSecret{Name: name, Value: nil}
	if err := client.UpdateSecrets(ctx, project, config, []RawSecret{newSecret}); err != nil {
		return diag.FromErr(err)
	}

	return diags
}
