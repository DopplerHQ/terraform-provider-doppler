package doppler

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceRotatedSecretTwilio() *schema.Resource {
	return ResourceRotatedSecretBuilder{}.Build()
}

func resourceRotatedSecretCloudflareTokens() *schema.Resource {
	builder := ResourceRotatedSecretBuilder{
		CredentialsSchema: map[string]*schema.Schema{
			"credentials": {
				Description: "Rotated secret credentials",
				Type:        schema.TypeList,
				MaxItems:    2,
				MinItems:    2,
				Required:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"value": {
							Type:      schema.TypeString,
							Required:  true,
							Sensitive: true,
						},
					},
				},
			},
		},
		CredentialsBuilder: func(d *schema.ResourceData) RotatedSecretCredentials {
			rawCredentials := d.Get("credentials").([]interface{})
			credentials := make([]map[string]interface{}, len(rawCredentials))
			for i, cred := range rawCredentials {
				credentials[i] = map[string]interface{}{
					"VALUE": cred.(map[string]interface{})["value"],
				}
			}
			return credentials
		},
	}
	return builder.Build()
}

func resourceRotatedSecretMongoDBAtlas() *schema.Resource {
	builder := ResourceRotatedSecretBuilder{
		ParametersSchema: map[string]*schema.Schema{
			"project_id": {
				Description: "The Mongo DB Atlas Project ID",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
		ParametersBuilder: func(d *schema.ResourceData) RotatedSecretParameters {
			return map[string]interface{}{
				"projectId": d.Get("project_id"),
			}
		},
		CredentialsSchema: map[string]*schema.Schema{
			"credentials": {
				Description: "Rotated secret credentials",
				Type:        schema.TypeList,
				MaxItems:    2,
				MinItems:    2,
				Required:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"username": {
							Type:     schema.TypeString,
							Required: true,
						},
						"password": {
							Type:      schema.TypeString,
							Required:  true,
							Sensitive: true,
						},
					},
				},
			},
		},
		CredentialsBuilder: func(d *schema.ResourceData) RotatedSecretCredentials {
			rawCredentials := d.Get("credentials").([]interface{})
			credentials := make([]map[string]interface{}, len(rawCredentials))
			for i, cred := range rawCredentials {
				credentials[i] = map[string]interface{}{
					"USERNAME": cred.(map[string]interface{})["username"],
					"PASSWORD": cred.(map[string]interface{})["password"],
				}
			}
			return credentials
		},
	}
	return builder.Build()
}

func resourceRotatedSecretSendGrid() *schema.Resource {
	builder := ResourceRotatedSecretBuilder{
		ParametersSchema: map[string]*schema.Schema{
			"scopes": {
				Description: "SendGrid scopes",
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				MinItems:    1,
				Required:    true,
			},
		},
		ParametersBuilder: func(d *schema.ResourceData) RotatedSecretParameters {
			return map[string]interface{}{
				"scopes": d.Get("scopes"),
			}
		},
	}
	return builder.Build()
}

func resourceRotatedSecretGCPCloudSQL() *schema.Resource {
	builder := ResourceRotatedSecretBuilder{
		ParametersSchema: map[string]*schema.Schema{
			"database_instance": {
				Description: "Cloud SQL database instance name",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
		ParametersBuilder: func(d *schema.ResourceData) RotatedSecretParameters {
			return map[string]interface{}{
				"DATABASE_INSTANCE": d.Get("database_instance"),
			}
		},
		CredentialsSchema: map[string]*schema.Schema{
			"credentials": {
				Description: "Rotated secret credentials",
				Type:        schema.TypeList,
				MaxItems:    2,
				MinItems:    2,
				Required:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"username": {
							Type:     schema.TypeString,
							Required: true,
						},
						"user_host_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"password": {
							Type:      schema.TypeString,
							Required:  true,
							Sensitive: true,
						},
					},
				},
			},
		},
		CredentialsBuilder: func(d *schema.ResourceData) RotatedSecretCredentials {
			rawCredentials := d.Get("credentials").([]interface{})
			credentials := make([]map[string]interface{}, len(rawCredentials))
			for i, cred := range rawCredentials {
				credentials[i] = map[string]interface{}{
					"USERNAME":       cred.(map[string]interface{})["username"],
					"USER_HOST_NAME": cred.(map[string]interface{})["user_host_name"],
					"PASSWORD":       cred.(map[string]interface{})["password"],
				}
			}
			return credentials
		},
	}
	return builder.Build()
}

func resourceRotatedSecretGCPServiceAccountKeys() *schema.Resource {
	builder := ResourceRotatedSecretBuilder{
		ParametersSchema: map[string]*schema.Schema{
			"service_account": {
				Description: "The Service Account Email whose keys should be rotated",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
		ParametersBuilder: func(d *schema.ResourceData) RotatedSecretParameters {
			return map[string]interface{}{
				"serviceAccount": d.Get("service_account"),
			}
		},
	}
	return builder.Build()
}

func resourceRotatedSecretAWSIAMUserKeys() *schema.Resource {
	builder := ResourceRotatedSecretBuilder{
		ParametersSchema: map[string]*schema.Schema{
			"username": {
				Description: "The IAM username (excluding ARN and scope prefix)",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
		ParametersBuilder: func(d *schema.ResourceData) RotatedSecretParameters {
			return map[string]interface{}{
				"username": d.Get("username"),
			}
		},
	}
	return builder.Build()
}

func resourceRotatedSecretAWSMySQL() *schema.Resource {
	builder := ResourceRotatedSecretBuilder{
		ParametersSchema: map[string]*schema.Schema{
			"host": {
				Description: "The DB Host",
				Type:        schema.TypeString,
				Required:    true,
			},
			"port": {
				Description: "The DB Port",
				Type:        schema.TypeInt,
				Required:    true,
			},
			"database": {
				Description: "The DB database",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"managing_user_username": {
				Description: "The managing user username",
				Type:        schema.TypeString,
				Required:    true,
			},
			"managing_user_password": {
				Description: "The managing user password",
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
			},
		},
		ParametersBuilder: func(d *schema.ResourceData) RotatedSecretParameters {
			return map[string]interface{}{
				"HOST":     d.Get("host"),
				"PORT":     d.Get("port"),
				"DATABASE": d.Get("database"),
				"MANAGING_USER": map[string]interface{}{
					"USERNAME": d.Get("managing_user_username"),
					"PASSWORD": d.Get("managing_user_password"),
				},
			}
		},
		CredentialsSchema: map[string]*schema.Schema{
			"credentials": {
				Description: "Rotated secret credentials",
				Type:        schema.TypeList,
				MaxItems:    2,
				MinItems:    2,
				Required:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"username": {
							Type:     schema.TypeString,
							Required: true,
						},
						"password": {
							Type:      schema.TypeString,
							Required:  true,
							Sensitive: true,
						},
					},
				},
			},
		},
		CredentialsBuilder: func(d *schema.ResourceData) RotatedSecretCredentials {
			rawCredentials := d.Get("credentials").([]interface{})
			credentials := make([]map[string]interface{}, len(rawCredentials))
			for i, cred := range rawCredentials {
				credentials[i] = map[string]interface{}{
					"USERNAME": cred.(map[string]interface{})["username"],
					"PASSWORD": cred.(map[string]interface{})["password"],
				}
			}
			return credentials
		},
	}
	return builder.Build()
}

func resourceRotatedSecretAWSPostgres() *schema.Resource {
	builder := ResourceRotatedSecretBuilder{
		ParametersSchema: map[string]*schema.Schema{
			"host": {
				Description: "The DB Host",
				Type:        schema.TypeString,
				Required:    true,
			},
			"port": {
				Description: "The DB Port",
				Type:        schema.TypeInt,
				Required:    true,
			},
			"database": {
				Description: "The DB database",
				Type:        schema.TypeString,
				Required:    true,
			},
			"ssl_ca": {
				Description: "An optional SSL CA for the AWS region the DB is in",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"managing_user_username": {
				Description: "The managing user username",
				Type:        schema.TypeString,
				Required:    true,
			},
			"managing_user_password": {
				Description: "The managing user password",
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
			},
		},
		ParametersBuilder: func(d *schema.ResourceData) RotatedSecretParameters {
			return map[string]interface{}{
				"HOST":     d.Get("host"),
				"PORT":     d.Get("port"),
				"DATABASE": d.Get("database"),
				"SSL_CA":   d.Get("ssl_ca"),
				"MANAGING_USER": map[string]interface{}{
					"USERNAME": d.Get("managing_user_username"),
					"PASSWORD": d.Get("managing_user_password"),
				},
			}
		},
		CredentialsSchema: map[string]*schema.Schema{
			"credentials": {
				Description: "Rotated secret credentials",
				Type:        schema.TypeList,
				MaxItems:    2,
				MinItems:    2,
				Required:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"username": {
							Type:     schema.TypeString,
							Required: true,
						},
						"password": {
							Type:      schema.TypeString,
							Required:  true,
							Sensitive: true,
						},
					},
				},
			},
		},
		CredentialsBuilder: func(d *schema.ResourceData) RotatedSecretCredentials {
			rawCredentials := d.Get("credentials").([]interface{})
			credentials := make([]map[string]interface{}, len(rawCredentials))
			for i, cred := range rawCredentials {
				credentials[i] = map[string]interface{}{
					"USERNAME": cred.(map[string]interface{})["username"],
					"PASSWORD": cred.(map[string]interface{})["password"],
				}
			}
			return credentials
		},
	}
	return builder.Build()
}

func resourceRotatedSecretAWSMSSQLServer() *schema.Resource {
	builder := ResourceRotatedSecretBuilder{
		ParametersSchema: map[string]*schema.Schema{
			"host": {
				Description: "The DB Host",
				Type:        schema.TypeString,
				Required:    true,
			},
			"port": {
				Description: "The DB Port",
				Type:        schema.TypeInt,
				Required:    true,
			},
			"database": {
				Description: "The DB database",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
		ParametersBuilder: func(d *schema.ResourceData) RotatedSecretParameters {
			return map[string]interface{}{
				"HOST":     d.Get("host"),
				"PORT":     d.Get("port"),
				"DATABASE": d.Get("database"),
			}
		},
		CredentialsSchema: map[string]*schema.Schema{
			"credentials": {
				Description: "Rotated secret credentials",
				Type:        schema.TypeList,
				MaxItems:    2,
				MinItems:    2,
				Required:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"username": {
							Type:     schema.TypeString,
							Required: true,
						},
						"password": {
							Type:      schema.TypeString,
							Required:  true,
							Sensitive: true,
						},
					},
				},
			},
		},
		CredentialsBuilder: func(d *schema.ResourceData) RotatedSecretCredentials {
			rawCredentials := d.Get("credentials").([]interface{})
			credentials := make([]map[string]interface{}, len(rawCredentials))
			for i, cred := range rawCredentials {
				credentials[i] = map[string]interface{}{
					"USERNAME": cred.(map[string]interface{})["username"],
					"PASSWORD": cred.(map[string]interface{})["password"],
				}
			}
			return credentials
		},
	}
	return builder.Build()
}
