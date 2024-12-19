package doppler

import (
	"context"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var resourceChangeRequestPolicyRequiredReviewers = schema.Resource{
	Schema: map[string]*schema.Schema{
		"count": {
			Description: "The number of approvals a change request must receive before it can be applied",
			Type:        schema.TypeInt,
			Required:    true,
		},
		"user_slugs": {
			Description: "If set, only approvals from these users will satisfy this rule",
			Type:        schema.TypeSet,
			Optional:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"group_slugs": {
			Description: "If set, only approvals from members of these groups will satisfy this rule",
			Type:        schema.TypeSet,
			Optional:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
	},
}

var resourceChangeRequestPolicyRules = schema.Resource{
	Schema: map[string]*schema.Schema{
		"disallow_self_review": {
			Description: "If true, approvals from the author of a change request will be excluded when evaluating this policy",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
		},
		"required_reviewers": {
			Description: "Enforces that a specific number of users approve a change request before it can be applied",
			Type:        schema.TypeSet,
			Optional:    true,
			Elem:        &resourceChangeRequestPolicyRequiredReviewers,
		},
	},
}

var resourceChangeRequestPolicyTargetProject = schema.Resource{
	Schema: map[string]*schema.Schema{
		"project_name": {
			Description: "The name of the project to apply the policy to",
			Type:        schema.TypeString,
			Required:    true,
		},
		"all": {
			Description: "Whether or not the policy applies to all configs in the project",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
		},
		"environment_slugs": {
			Description: "Entire environments the policy applies to",
			Type:        schema.TypeSet,
			Optional:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"config_names": {
			Description: "Specific configs the policy applies to",
			Type:        schema.TypeSet,
			Optional:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
	},
}

var resourceChangeRequestPolicyTargets = schema.Resource{
	Schema: map[string]*schema.Schema{
		"all_projects": {
			Description: "Whether or not the policy applies to all projects in the workplace",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
		},
		"project": {
			Description: "A project that the policy will apply to",
			Type:        schema.TypeSet,
			Optional:    true,
			Elem:        &resourceChangeRequestPolicyTargetProject,
		},
	},
}

func resourceChangeRequestPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceChangeRequestPolicyCreate,
		ReadContext:   resourceChangeRequestPolicyRead,
		UpdateContext: resourceChangeRequestPolicyUpdate,
		DeleteContext: resourceChangeRequestPolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"slug": {
				Description: "The unique identifier of the change request policy",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				Description: "The name of the change request policy",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "A description of the change request policy",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
			},
			"rules": {
				Description: "Rules that the policy will apply to its targets",
				Type:        schema.TypeList,
				MaxItems:    1,
				MinItems:    1,
				Required:    true,
				Elem:        &resourceChangeRequestPolicyRules,
			},
			"targets": {
				Description: "Where the policy will apply",
				Type:        schema.TypeList,
				MaxItems:    1,
				MinItems:    1,
				Required:    true,
				Elem:        &resourceChangeRequestPolicyTargets,
			},
		},
	}
}

func resourceChangeRequestPolicyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)

	var diags diag.Diagnostics

	payload, diags := toChangeRequestPolicy(d, diags)
	if diags.HasError() {
		return diags
	}

	policy, err := client.CreateChangeRequestPolicy(ctx, &payload)
	if err != nil {
		diags = append(diags, diag.FromErr(err)...)
		return diags
	}

	diags = updateChangeRequestPolicyState(d, policy, diags)
	return diags
}

func resourceChangeRequestPolicyUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)

	var diags diag.Diagnostics

	payload, diags := toChangeRequestPolicy(d, diags)
	if diags.HasError() {
		return diags
	}

	policy, err := client.UpdateChangeRequestPolicy(ctx, &payload)
	if err != nil {
		diags = append(diags, diag.FromErr(err)...)
		return diags
	}

	diags = updateChangeRequestPolicyState(d, policy, diags)
	return diags
}

func resourceChangeRequestPolicyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)

	var diags diag.Diagnostics
	slug := d.Id()

	policy, err := client.GetChangeRequestPolicy(ctx, slug)
	if err != nil {
		return handleNotFoundError(err, d)
	}

	diags = updateChangeRequestPolicyState(d, policy, diags)
	return diags
}

func resourceChangeRequestPolicyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(APIClient)

	var diags diag.Diagnostics
	slug := d.Id()

	if err := client.DeleteChangeRequestPolicy(ctx, slug); err != nil {
		diags = append(diags, diag.FromErr(err)...)
		return diags
	}

	return diags
}

func toChangeRequestPolicy(d *schema.ResourceData, diags diag.Diagnostics) (ChangeRequestPolicy, diag.Diagnostics) {
	policy := ChangeRequestPolicy{
		Slug:        d.Id(),
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Rules:       make([]ChangeRequestPolicyRule, 0),
		Targets: ChangeRequestPolicyTargets{
			AllProjects: false,
			Projects:    make(map[string]ChangeRequestPolicyTargetProject),
		},
	}

	rulesList := d.Get("rules").([]interface{})
	rules := rulesList[0].(map[string]interface{}) // This is required in the schema, panic if it doesn't exist
	if rules["disallow_self_review"] == true {
		policy.Rules = append(policy.Rules, ChangeRequestPolicyRule{Type: "DisallowSelfReview"})
	}
	if rules["required_reviewers"] != nil {
		for _, req := range rules["required_reviewers"].(*schema.Set).List() {
			reqDef := req.(map[string]interface{})
			rule := ChangeRequestPolicyRule{Type: "RequiredReviewer", Count: reqDef["count"].(int)}

			if reqDef["user_slugs"] != nil || reqDef["group_slugs"] != nil {
				rule.Subjects = make([]ChangeRequestPolicySubject, 0)
			}
			if reqDef["user_slugs"] != nil {
				for _, slug := range reqDef["user_slugs"].(*schema.Set).List() {
					rule.Subjects = append(rule.Subjects, ChangeRequestPolicySubject{Type: "WorkplaceUser", Slug: slug.(string)})
				}
			}
			if reqDef["group_slugs"] != nil {
				for _, slug := range reqDef["group_slugs"].(*schema.Set).List() {
					rule.Subjects = append(rule.Subjects, ChangeRequestPolicySubject{Type: "Group", Slug: slug.(string)})
				}
			}

			policy.Rules = append(policy.Rules, rule)
		}
	}

	targetsList := d.Get("targets").([]interface{})
	targets := targetsList[0].(map[string]interface{}) // This is required in the schema, panic if it doesn't exist
	policy.Targets.AllProjects = targets["all_projects"].(bool)

	if targets["all_projects"].(bool) && targets["project"].(*schema.Set).Len() > 0 {
		diags = append(diags, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       "Conflicting values for targets.all_projects and targets.project",
			Detail:        "targets.all_projects supersedes all project assignments. Please remove any project blocks when applying a policy to all projects.",
			AttributePath: cty.GetAttrPath("targets").IndexInt(0).GetAttr("all_projects"),
		})
	}

	for _, p := range targets["project"].(*schema.Set).List() {
		project := p.(map[string]interface{})
		envs := project["environment_slugs"].(*schema.Set).List()
		configs := project["config_names"].(*schema.Set).List()

		target := ChangeRequestPolicyTargetProject{
			All:         project["all"].(bool),
			EnvSlugs:    make([]string, 0),
			ConfigNames: make([]string, 0),
		}

		for _, env := range envs {
			target.EnvSlugs = append(target.EnvSlugs, env.(string))
		}
		for _, config := range configs {
			target.ConfigNames = append(target.ConfigNames, config.(string))
		}
		if target.All {
			if len(target.EnvSlugs) > 0 {
				diags = append(diags, diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       "Conflicting values for project.all and project.environment_slugs",
					Detail:        "project.all supersedes all enviroment assignments. Please remove any specific enviroment assigments when applying a policy to an entire project.",
					AttributePath: cty.GetAttrPath("targets").IndexInt(0).GetAttr("project"),
				})
			}
			if len(target.ConfigNames) > 0 {
				diags = append(diags, diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       "Conflicting values for project.all and project.config_names",
					Detail:        "project.all supersedes all config assignments. Please remove any specific config assigments when applying a policy to an entire project.",
					AttributePath: cty.GetAttrPath("targets").IndexInt(0).GetAttr("project"),
				})
			}

			target.EnvSlugs = nil
			target.ConfigNames = nil
		} else if len(target.EnvSlugs) == 0 && len(target.ConfigNames) == 0 {
			diags = append(diags, diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       "Conflicting values for project.all, project.environment_slugs, and project.config_names",
				Detail:        "When project.all is false, at least one of project.environment_slugs or project.config_names must be set to a non-empty value",
				AttributePath: cty.GetAttrPath("targets").IndexInt(0).GetAttr("project"),
			})
		}
		policy.Targets.Projects[project["project_name"].(string)] = target
	}

	return policy, diags
}

func updateChangeRequestPolicyState(d *schema.ResourceData, policy *ChangeRequestPolicy, diags diag.Diagnostics) diag.Diagnostics {
	if err := d.Set("slug", policy.Slug); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	if err := d.Set("name", policy.Name); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	if err := d.Set("description", policy.Description); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	requiredReviewers := schema.NewSet(schema.HashResource(&resourceChangeRequestPolicyRequiredReviewers), make([]interface{}, 0))
	for _, rule := range policy.Rules {
		if rule.Type != "RequiredReviewer" {
			continue
		}
		item := make(map[string]interface{})
		item["count"] = rule.Count
		if len(rule.Subjects) != 0 {
			groupSlugs := schema.NewSet(schema.HashString, make([]interface{}, 0))
			userSlugs := schema.NewSet(schema.HashString, make([]interface{}, 0))
			if rule.Subjects != nil {
				for _, sub := range rule.Subjects {
					if sub.Type == "WorkplaceUser" {
						userSlugs.Add(sub.Slug)
					}
					if sub.Type == "Group" {
						groupSlugs.Add(sub.Slug)
					}
				}
			}
			item["group_slugs"] = groupSlugs
			item["user_slugs"] = userSlugs
		}
		requiredReviewers.Add(item)
	}

	rules := make(map[string]interface{})
	if requiredReviewers.Len() > 0 {
		rules["required_reviewers"] = requiredReviewers
	}
	rules["disallow_self_review"] = false
	for _, rule := range policy.Rules {
		if rule.Type == "DisallowSelfReview" {
			rules["disallow_self_review"] = true
			break
		}
	}

	rulesList := make([]map[string]interface{}, 1)
	rulesList[0] = rules
	if err := d.Set("rules", rulesList); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	projects := schema.NewSet(schema.HashResource(&resourceChangeRequestPolicyTargetProject), make([]interface{}, 0))
	for projectName, target := range policy.Targets.Projects {
		project := make(map[string]interface{})
		project["project_name"] = projectName
		project["all"] = target.All

		envSlugs := schema.NewSet(schema.HashString, make([]interface{}, 0))
		if target.EnvSlugs != nil {
			for _, env := range target.EnvSlugs {
				envSlugs.Add(env)
			}
		}
		project["environment_slugs"] = envSlugs

		configNames := schema.NewSet(schema.HashString, make([]interface{}, 0))
		if target.ConfigNames != nil {
			for _, config := range target.ConfigNames {
				configNames.Add(config)
			}
		}
		project["config_names"] = configNames

		projects.Add(project)
	}

	targets := make(map[string]interface{})
	targets["all_projects"] = policy.Targets.AllProjects
	targets["project"] = projects

	targetsList := make([]map[string]interface{}, 1)
	targetsList[0] = targets
	if err := d.Set("targets", targetsList); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	d.SetId(policy.Slug)
	return diags
}
