package doppler

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceSyncAWSSecretsManager() *schema.Resource {
	builder := ResourceSyncBuilder{
		DataSchema: map[string]*schema.Schema{
			"region": {
				Description: "The AWS region",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"path": {
				Description: "The path to the secret in AWS",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"kms_key_id": {
				Description: "The AWS KMS key used to encrypt the secret (ID, Alias, or ARN)",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"tags": {
				Description: "AWS tags to attach to the secrets",
				Type:        schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
				ForceNew: true,
			},
			"update_metadata": {
				Description: "If enabled, Doppler will update the AWS secret metadata (e.g. KMS key) during every sync. If disabled, Doppler will only set secret metadata for new AWS secrets. Note that Doppler never updates tags for existing AWS secrets.",
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    true,
			},

			"update_resource_tags": {
				Description:  "Behavior for AWS resource tags on updates (never update, upsert tags (leaving non-Doppler tags alone), replace tags (remove non-Doppler tags))",
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"never", "upsert", "replace"}, false),
				DiffSuppressFunc: func(k, oldValue, newValue string, d *schema.ResourceData) bool {
					if oldValue == "" && newValue == "never" {
						return true
					} else if oldValue == "never" && newValue == "" {
						return true
					} else {
						return newValue == oldValue
					}
				},
			},

			"name_transform": {
				Description:  fmt.Sprintf("An optional secret name transformer (e.g. DOPPLER_CONFIG in lower-kebab would be doppler-config). Valid transformers: %v", strings.Join(NameTransformers, ", ")),
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(NameTransformers, false),
				DiffSuppressFunc: func(k, oldValue, newValue string, d *schema.ResourceData) bool {
					if oldValue == "" && newValue == "none" {
						return true
					} else if oldValue == "none" && newValue == "" {
						return true
					} else {
						return newValue == oldValue
					}
				},
			},

			"path_behavior": {
				Description: "The behavior to modify the provided path. Either `add_doppler_suffix` (default) which appends `doppler` to the provided path or `none` which leaves the path unchanged.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				// Implicitly defaults to "add_doppler_suffix" but not defined here to avoid state migration
				ValidateFunc: validation.StringInSlice([]string{"add_doppler_suffix", "none"}, false),
				DiffSuppressFunc: func(k, oldValue, newValue string, d *schema.ResourceData) bool {
					if oldValue == "" && newValue == "add_doppler_suffix" {
						// Adding the default value explicitly
						return true
					} else if oldValue == "add_doppler_suffix" && newValue == "" {
						// Removing the explicit default value
						return true
					} else {
						return false
					}
				},
			},
		},
		DataBuilder: func(d *schema.ResourceData) IntegrationData {
			payload := map[string]interface{}{
				"region": d.Get("region"),
				"path":   d.Get("path"),
				"tags":   d.Get("tags"),
			}
			if kmsKeyId, ok := d.GetOk("kms_key_id"); ok {
				payload["kms_key_id"] = kmsKeyId
			}
			if updateMetadata, ok := d.GetOk("update_metadata"); ok {
				payload["update_metadata"] = updateMetadata
			}
			if updateResourceTags, ok := d.GetOk("update_resource_tags"); ok {
				payload["update_resource_tags"] = updateResourceTags
			}
			if nameTransform, ok := d.GetOk("name_transform"); ok {
				payload["name_transform"] = nameTransform
			}
			if pathBehavior, ok := d.GetOk("path_behavior"); ok {
				payload["use_doppler_suffix"] = pathBehavior == "add_doppler_suffix"
			} else {
				payload["use_doppler_suffix"] = true
			}
			return payload
		},
	}
	return builder.Build()
}

func resourceSyncAWSParameterStore() *schema.Resource {
	builder := ResourceSyncBuilder{
		DataSchema: map[string]*schema.Schema{
			"region": {
				Description: "The AWS region",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"path": {
				Description: "The path to the parameters in AWS",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"secure_string": {
				Description: "Whether or not the parameters are stored as a secure string",
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    true,
				Default:     true,
			},
			"kms_key_id": {
				Description: "The AWS KMS key used to encrypt the parameter (ID, Alias, or ARN) ",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"tags": {
				Description: "AWS tags to attach to the parameters",
				Type:        schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
				ForceNew: true,
			},
			"update_resource_tags": {
				Description:  "Behavior for AWS resource tags on updates (never update, upsert tags (leaving non-Doppler tags alone), replace tags (remove non-Doppler tags))",
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"never", "upsert", "replace"}, false),
				DiffSuppressFunc: func(k, oldValue, newValue string, d *schema.ResourceData) bool {
					if oldValue == "" && newValue == "never" {
						return true
					} else if oldValue == "never" && newValue == "" {
						return true
					} else {
						return newValue == oldValue
					}
				},
			},
			"name_transform": {
				Description:  fmt.Sprintf("An optional secret name transformer (e.g. DOPPLER_CONFIG in lower-kebab would be doppler-config). Valid transformers: %v", strings.Join(NameTransformers, ", ")),
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(NameTransformers, false),
				DiffSuppressFunc: func(k, oldValue, newValue string, d *schema.ResourceData) bool {
					if oldValue == "" && newValue == "none" {
						return true
					} else if oldValue == "none" && newValue == "" {
						return true
					} else {
						return newValue == oldValue
					}
				},
			},
		},
		DataBuilder: func(d *schema.ResourceData) IntegrationData {
			payload := map[string]interface{}{
				"region":        d.Get("region"),
				"path":          d.Get("path"),
				"secure_string": d.Get("secure_string"),
				"tags":          d.Get("tags"),
			}
			if kmsKeyId, ok := d.GetOk("kms_key_id"); ok {
				payload["kms_key_id"] = kmsKeyId
			}
			if updateResourceTags, ok := d.GetOk("update_resource_tags"); ok {
				payload["update_resource_tags"] = updateResourceTags
			}
			if nameTransform, ok := d.GetOk("name_transform"); ok {
				payload["name_transform"] = nameTransform
			}
			return payload
		},
	}
	return builder.Build()
}

func resourceSyncCircleCi() *schema.Resource {
	builder := ResourceSyncBuilder{
		DataSchema: map[string]*schema.Schema{
			"resource_type": {
				Description:  "Either \"project\" or \"context\", based on the resource type to sync to",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"project", "context"}, false),
			},
			"resource_id": {
				Description: "The resource ID (either project or context) to sync to",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"organization_slug": {
				Description: "The organization slug where the resource is located",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
		},
		DataBuilder: func(d *schema.ResourceData) IntegrationData {
			return map[string]interface{}{
				"resource_type":     d.Get("resource_type"),
				"resource_id":       d.Get("resource_id"),
				"organization_slug": d.Get("organization_slug"),
			}
		},
	}
	return builder.Build()
}

func resourceSyncGitHubActions() *schema.Resource {
	builder := ResourceSyncBuilder{
		DataSchema: map[string]*schema.Schema{
			"sync_target": {
				Description:  "Either \"repo\" or \"org\", based on the resource type to sync to",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"repo", "org"}, false),
			},
			"repo_name": {
				Description:  "The GitHub repo name to sync to (only used when `sync_target` is set to \"repo\")",
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"repo_name", "org_scope"},
			},
			"org_scope": {
				Description:  "Either \"all\" or \"private\", based on the which repos you want to have access (only used when `sync_target` is set to \"org\")",
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"repo_name", "org_scope"},
				ValidateFunc: validation.StringInSlice([]string{"all", "private"}, false),
			},
			"environment_name": {
				Description: "The GitHub repo environment name to sync to (only used when `sync_target` is set to \"repo\")",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
		},
		DataBuilder: func(d *schema.ResourceData) IntegrationData {
			payload := map[string]interface{}{
				"sync_target": d.Get("sync_target"),
			}
			repo_name := d.Get("repo_name")
			if repo_name != "" {
				payload["repo_name"] = repo_name
			}
			org_scope := d.Get("org_scope")
			if org_scope != "" {
				payload["org_scope"] = org_scope
			}
			environment_name := d.Get("environment_name")
			if environment_name != "" {
				payload["environment_name"] = environment_name
			}
			return payload
		},
	}
	return builder.Build()
}

func resourceSyncTerraformCloud() *schema.Resource {
	builder := ResourceSyncBuilder{
		DataSchema: map[string]*schema.Schema{
			"sync_target": {
				Description: "Either \"workspace\" or \"variableSet\", based on the resource type to sync to",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"workspace_id": {
				Description:  "The Terraform Cloud workspace ID to sync to",
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"workspace_id", "variable_set_id"},
			},
			"variable_set_id": {
				Description:  "The Terraform Cloud variable set ID to sync to",
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"workspace_id", "variable_set_id"},
			},
			"variable_sync_type": {
				Description: "Either \"terraform\" to sync secrets as Terraform variables or \"env\" to sync as environment variables",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name_transform": {
				Description: "A name transform to apply before syncing secrets: \"none\" or \"lowercase\"",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
		},
		DataBuilder: func(d *schema.ResourceData) IntegrationData {
			return map[string]interface{}{
				"sync_target":        d.Get("sync_target"),
				"workspace_id":       d.Get("workspace_id"),
				"variable_set_id":    d.Get("variable_set_id"),
				"variable_sync_type": d.Get("variable_sync_type"),
				"name_transform":     d.Get("name_transform"),
			}
		},
	}
	return builder.Build()
}

func resourceSyncFlyio() *schema.Resource {
	builder := ResourceSyncBuilder{
		DataSchema: map[string]*schema.Schema{
			"app_id": {
				Description: "The app ID ",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"restart_machines": {
				Description: "Whether or not to restart the Fly.io machines when secrets are updated",
				Type:        schema.TypeBool,
				Required:    true,
				ForceNew:    true,
			},
		},
		DataBuilder: func(d *schema.ResourceData) IntegrationData {
			return map[string]interface{}{
				"app_id":           d.Get("app_id"),
				"restart_machines": d.Get("restart_machines"),
			}
		},
	}
	return builder.Build()
}
