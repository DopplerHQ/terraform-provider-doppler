package doppler

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceWebhook() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWebhookCreate,
		ReadContext:   resourceWebhookRead,
		UpdateContext: resourceWebhookUpdate,
		DeleteContext: resourceWebhookDelete,
		Schema: map[string]*schema.Schema{
			"slug": {
				Description: "The slug of the Webhook",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"project": {
				Description: "The name of the Doppler project where the webhook is located",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"url": {
				Description: "The URL of the webhook endpoint",
				Type:        schema.TypeString,
				Required:    true,
			},
			"enabled": {
				Description: "Whether the webhook is enabled or disabled.  Default to true.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
			"secret": {
				Description: "Secret used for request signing",
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
			},
			"name": {
				Description: "Name of the webhook",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"authentication": {
				Description: "Authentication method used by the webhook",
				Type:        schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"token": {
							Type:      schema.TypeString,
							Optional:  true,
							Sensitive: true,
						},
						"username": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"password": {
							Type:      schema.TypeString,
							Optional:  true,
							Sensitive: true,
						},
					},
				},
				MaxItems: 1,
				Optional: true,
			},
			"payload": {
				Description: "The webhook's payload as a JSON string.  Leave empty to use the default webhook payload",
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
			},
			"enabled_configs": {
				Description: "Configs this webhook will trigger for",
				Type:        schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
			},
		},
	}
}

func resourceWebhookCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)

	var diags diag.Diagnostics
	project := d.Get("project").(string)
	url := d.Get("url").(string)
	enabled := d.Get("enabled").(bool)
	secret := d.Get("secret").(string)
	payload := d.Get("payload").(string)
	name := d.Get("name").(string)

	rawEnabledConfigs := d.Get("enabled_configs").(*schema.Set).List()
	enabledConfigs := make([]string, len(rawEnabledConfigs))
	for i, v := range rawEnabledConfigs {
		enabledConfigs[i] = v.(string)
	}

	options := CreateWebhookOptionalParameters{Secret: secret, WebhookPayload: payload, EnabledConfigs: enabledConfigs, Name: name}

	authConfigList := d.Get("authentication").([]interface{})

	if len(authConfigList) > 0 {
		authMap, ok := authConfigList[0].(map[string]interface{}) // schema allows only 1 item
		if !ok {
			return diag.FromErr(fmt.Errorf("unexpected type for authentication element: %T", authConfigList))
		}

		options.Auth = &WebhookAuth{
			Type:     authMap["type"].(string),
			Token:    authMap["token"].(string),
			Username: authMap["username"].(string),
			Password: authMap["password"].(string),
		}
	}

	webhook, err := client.CreateWebhook(ctx, project, url, enabled, &options)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(webhook.Slug)

	return diags
}

func resourceWebhookUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)

	var diags diag.Diagnostics
	slug := d.Id()
	project := d.Get("project").(string)
	secret := d.Get("secret").(string)
	payload := d.Get("payload").(string)
	name := d.Get("name").(string)

	if d.HasChange("enabled") {
		if d.Get("enabled").(bool) {
			_, err := client.EnableWebhook(ctx, project, slug)
			if err != nil {
				return diag.FromErr(err)
			}
		} else {
			_, err := client.DisableWebhook(ctx, project, slug)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	url := d.Get("url").(string)
	enabledConfigs := []string{}
	disabledConfigs := []string{}
	if d.HasChange("enabled_configs") {
		oldRawEnabledConfigs, newRawEnabledConfigs := d.GetChange("enabled_configs")
		oldRawEnabledConfigsArr := oldRawEnabledConfigs.(*schema.Set).List()
		newRawEnabledConfigsArr := newRawEnabledConfigs.(*schema.Set).List()
		oldEnabledConfigsMap := make(map[string]string)
		for _, v := range oldRawEnabledConfigsArr {
			oldEnabledConfigsMap[v.(string)] = v.(string)
		}
		newEnabledConfigsMap := make(map[string]string)
		for _, v := range newRawEnabledConfigsArr {
			newEnabledConfigsMap[v.(string)] = v.(string)
		}

		for _, v := range newRawEnabledConfigsArr {
			if _, ok := oldEnabledConfigsMap[v.(string)]; !ok {
				enabledConfigs = append(enabledConfigs, v.(string))
			}
		}

		for _, v := range oldRawEnabledConfigsArr {
			if _, ok := newEnabledConfigsMap[v.(string)]; !ok {
				disabledConfigs = append(disabledConfigs, v.(string))
			}
		}
	}

	authConfigList := d.Get("authentication").([]interface{})
	var auth WebhookAuth

	if len(authConfigList) > 0 {
		authMap, ok := authConfigList[0].(map[string]interface{}) // schema allows only 1 item
		if !ok {
			return diag.FromErr(fmt.Errorf("unexpected type for authentication element: %T", authConfigList))
		}

		auth = WebhookAuth{
			Type:     authMap["type"].(string),
			Token:    authMap["token"].(string),
			Username: authMap["username"].(string),
			Password: authMap["password"].(string),
		}
	} else {
		auth = WebhookAuth{Type: "None"}
	}

	webhook, err := client.UpdateWebhook(ctx, project, slug, url, secret, payload, name, enabledConfigs, disabledConfigs, auth)

	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(webhook.Slug)
	return diags
}

func resourceWebhookRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)

	var diags diag.Diagnostics
	slug := d.Id()
	project := d.Get("project").(string)

	webhook, err := client.GetWebhook(ctx, project, slug)
	if err != nil {
		return handleNotFoundError(err, d)
	}

	if err = d.Set("slug", webhook.Slug); err != nil {
		return diag.FromErr(err)
	}

	if err = d.Set("url", webhook.Url); err != nil {
		return diag.FromErr(err)
	}

	if err = d.Set("enabled", webhook.Enabled); err != nil {
		return diag.FromErr(err)
	}

	if err = d.Set("enabled_configs", webhook.EnabledConfigs); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceWebhookDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)

	var diags diag.Diagnostics
	slug := d.Id()
	project := d.Get("project").(string)

	if err := client.DeleteWebhook(ctx, project, slug); err != nil {
		return diag.FromErr(err)
	}

	return diags
}
