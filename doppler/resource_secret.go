package doppler

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceSecret() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSecretUpdate,
		ReadContext:   resourceSecretRead,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		UpdateContext: resourceSecretUpdate,
		DeleteContext: resourceSecretDelete,
		Schema: map[string]*schema.Schema{
			"project": {
				Description: "The name of the Doppler project",
				Type:        schema.TypeString,
				Required:    true,
				// Secrets cannot be moved directly from one project to another, they must be re-created
				ForceNew: true,
			},
			"config": {
				Description: "The name of the Doppler config",
				Type:        schema.TypeString,
				Required:    true,
				// Secrets cannot be moved directly from one config to another, they must be re-created
				ForceNew: true,
			},
			"name": {
				Description: "The name of the Doppler secret",
				Type:        schema.TypeString,
				Required:    true,
			},
			"value": {
				Description:  "The raw secret value. This value is stored in the Terraform state. Exactly one of `value` or `value_wo` must be set.",
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				ExactlyOneOf: []string{"value", "value_wo"},
			},
			"value_wo": {
				Description:  "The raw secret value, as a write-only argument that is never stored in the Terraform plan or state. Terraform cannot detect changes to it, so change `value_wo_version` to write a new value. Exactly one of `value` or `value_wo` must be set.",
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				WriteOnly:    true,
				ExactlyOneOf: []string{"value", "value_wo"},
				RequiredWith: []string{"value_wo_version"},
			},
			"value_wo_version": {
				Description:  "A version number for `value_wo`. Change it to write a new `value_wo` to Doppler. Must be 1 or greater. Required when `value_wo` is set.",
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(1),
				RequiredWith: []string{"value_wo"},
			},
			"visibility": {
				Description:  "The visibility of the secret. One of `masked`, `unmasked`, or `restricted`. Defaults to `masked`.",
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "masked",
				ValidateFunc: validation.StringInSlice([]string{"masked", "unmasked", "restricted"}, false),
			},
			"computed": {
				Description: "The computed secret value, after resolving secret references. Not set when `value_wo` is used, so that the value is kept out of the Terraform state.",
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
			},
			"value_type": {
				Description: "The value type of the secret",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "string",
				ValidateFunc: validation.StringInSlice([]string{
					"string", "json", "json5", "boolean", "integer", "decimal", "email",
					"url", "uuidv4", "cuid2", "ulid", "datetime8601", "date8601", "yaml",
					"directReference", "xml",
				}, false),
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
	value, valueDiags := secretValue(d)
	if valueDiags.HasError() {
		return valueDiags
	}
	visibility := d.Get("visibility").(string)
	valueType := d.Get("value_type").(string)

	changeRequest := ChangeRequest{
		Name:       name,
		Value:      &value,
		Visibility: visibility,
		ValueType:  &ValueType{Type: valueType},
	}
	if !d.IsNewResource() {
		previousNameValue, _ := d.GetChange("name")
		previousName := previousNameValue.(string)
		changeRequest.OriginalName = &previousName
	} else {
		changeRequest.OriginalName = &name
	}

	// NOTE: We could set `OriginalValue` here to leverage staleness detection in the Doppler API.
	// However, Terraform already confirms the existing state before making an update.
	// We'll skip the API-level staleness checking to allow Terraform to push over an external change.

	if err := client.UpdateSecrets(ctx, project, config, []ChangeRequest{changeRequest}); err != nil {
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

	// A write-only value is never compared with Doppler, so it does not need to be readable.
	writeOnly := usesWriteOnlyValue(d)
	if !writeOnly {
		nilFields := []string{}
		if secret.Value.Raw == nil {
			nilFields = append(nilFields, "raw")
		}
		if secret.Value.Computed == nil {
			nilFields = append(nilFields, "computed")
		}

		if len(nilFields) > 0 {
			return diag.FromErr(fmt.Errorf(
				"One or more secret fields are restricted: %v. "+
					"You must use a service account or service token to manage these resources. "+
					"Otherwise, Terraform cannot fetch these restricted secrets to check the validity of their state.", nilFields))
		}
	}

	if err = d.Set("project", project); err != nil {
		return diag.FromErr(err)
	}

	if err = d.Set("config", config); err != nil {
		return diag.FromErr(err)
	}

	if err = d.Set("name", name); err != nil {
		return diag.FromErr(err)
	}

	if writeOnly {
		if err = d.Set("value", nil); err != nil {
			return diag.FromErr(err)
		}
		if err = d.Set("computed", nil); err != nil {
			return diag.FromErr(err)
		}
	} else {
		if err = d.Set("value", secret.Value.Raw); err != nil {
			return diag.FromErr(err)
		}
		if err = d.Set("computed", secret.Value.Computed); err != nil {
			return diag.FromErr(err)
		}
	}

	if err = d.Set("visibility", secret.Value.RawVisibility); err != nil {
		return diag.FromErr(err)
	}

	if err = d.Set("value_type", secret.Value.RawValueType.Type); err != nil {
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

	changeRequest := ChangeRequest{OriginalName: &name, Name: name, ShouldDelete: true}
	if err := client.UpdateSecrets(ctx, project, config, []ChangeRequest{changeRequest}); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

// usesWriteOnlyValue reports whether the secret is managed through `value_wo`.
// Read has no configuration, so this is decided from `value_wo_version`, which
// is stored in state and is set exactly when `value_wo` is.
func usesWriteOnlyValue(d *schema.ResourceData) bool {
	return d.Get("value_wo_version").(int) > 0
}

// secretValue returns the value to write to Doppler, from `value_wo` when it is
// set and from `value` otherwise.
func secretValue(d *schema.ResourceData) (string, diag.Diagnostics) {
	if !usesWriteOnlyValue(d) {
		return d.Get("value").(string), nil
	}
	value, diags := d.GetRawConfigAt(cty.GetAttrPath("value_wo"))
	if diags.HasError() {
		return "", diags
	}
	// Validation already requires `value_wo` with `value_wo_version`; this
	// returns an error rather than panicking if that is ever bypassed.
	if value.IsNull() || !value.IsKnown() {
		return "", diag.Errorf("`value_wo` must be set when `value_wo_version` is set")
	}
	return value.AsString(), nil
}
