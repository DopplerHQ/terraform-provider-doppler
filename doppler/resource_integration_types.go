package doppler

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func awsAssumeRoleDataSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"assume_role_arn": {
			Description: "The ARN of the AWS role for Doppler to assume",
			Type:        schema.TypeString,
			Required:    true,
		},
	}
}

func awsAssumeRoleDataBuilder(d *schema.ResourceData) IntegrationData {
	return IntegrationData{
		"aws_assume_role_arn": d.Get("assume_role_arn"),
	}
}

func resourceIntegrationAWSAssumeRoleIntegration(integrationType string) *schema.Resource {
	builder := ResourceIntegrationBuilder{
		Type:        integrationType,
		DataSchema:  awsAssumeRoleDataSchema(),
		DataBuilder: awsAssumeRoleDataBuilder,
	}
	return builder.Build()
}

func resourceIntegrationTerraformCloud() *schema.Resource {
	builder := ResourceIntegrationBuilder{
		Type: "terraform_cloud",
		DataSchema: map[string]*schema.Schema{
			"api_key": {
				Description: "A Terraform Cloud API key.",
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
			},
		},
		DataBuilder: func(d *schema.ResourceData) IntegrationData {
			return map[string]interface{}{
				"api_key": d.Get("api_key"),
			}
		},
	}
	return builder.Build()
}

func resourceIntegrationFlyio() *schema.Resource {
	builder := ResourceIntegrationBuilder{
		Type: "flyio",
		DataSchema: map[string]*schema.Schema{
			"api_key": {
				Description: "A Fly.io API key.",
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
			},
		},
		DataBuilder: func(d *schema.ResourceData) IntegrationData {
			return map[string]interface{}{
				"api_key": d.Get("api_key"),
			}
		},
	}
	return builder.Build()
}
