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
