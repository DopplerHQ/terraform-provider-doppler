package doppler

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
		},
		DataBuilder: func(d *schema.ResourceData) IntegrationData {
			return map[string]interface{}{
				"region": d.Get("region"),
				"path":   d.Get("path"),
			}
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
			"tags": {
				Description: "AWS tags to attach to the parameters",
				Type:        schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
				ForceNew: true,
			},
		},
		DataBuilder: func(d *schema.ResourceData) IntegrationData {
			return map[string]interface{}{
				"region":        d.Get("region"),
				"path":          d.Get("path"),
				"secure_string": d.Get("secure_string"),
				"tags":          d.Get("tags"),
			}
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
				Description: "The Terraform Cloud workspace ID to sync to",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"variable_set_id": {
				Description: "The Terraform Cloud variable set ID to sync to",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
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
