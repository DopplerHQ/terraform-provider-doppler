package doppler

import (
	"encoding/json"
	"errors"
	"sort"
	"strings"
)

type ComputedSecret struct {
	Name  string
	Value string
}

func ParseComputedSecrets(response []byte) ([]ComputedSecret, error) {
	var result map[string]string
	err := json.Unmarshal(response, &result)
	if err != nil {
		return nil, err
	}

	secrets := make([]ComputedSecret, 0)
	for key, value := range result {
		secret := ComputedSecret{Name: key, Value: value}
		secrets = append(secrets, secret)
	}
	sort.Slice(secrets, func(i, j int) bool {
		return secrets[i].Name < secrets[j].Name
	})
	return secrets, nil
}

type Secret struct {
	Name  string      `json:"name"`
	Value SecretValue `json:"value"`
}

type SecretValue struct {
	Raw                *string `json:"raw,omitempty"`
	Computed           *string `json:"computed,omitempty"`
	RawVisibility      *string `json:"rawVisibility,omitempty"`
	ComputedVisibility *string `json:"computedVisibility,omitempty"`
}

func getSecretId(project string, config string, name string) string {
	return strings.Join([]string{project, config, name}, ".")
}

func parseSecretId(id string) (project string, config string, name string, err error) {
	tokens := strings.Split(id, ".")
	if len(tokens) != 3 {
		return "", "", "", errors.New("invalid secret ID")
	}
	return tokens[0], tokens[1], tokens[2], nil
}

type ChangeRequest struct {
	OriginalName       *string `json:"originalName,omitempty"`
	OriginalValue      *string `json:"originalValue,omitempty"`
	OriginalVisibility *string `json:"originalVisibility,omitempty"`
	Name               string  `json:"name"`
	Value              *string `json:"value"`
	ShouldDelete       bool    `json:"shouldDelete"`
	Visibility         string  `json:"visibility,omitempty"`
}

type Project struct {
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
}

type ProjectResponse struct {
	Project Project `json:"project"`
}

type ProjectMemberRole struct {
	Identifier string `json:"identifier"`
}

type ProjectMember struct {
	Type                  string            `json:"type"`
	Slug                  string            `json:"slug"`
	Role                  ProjectMemberRole `json:"role"`
	AccessAllEnvironments bool              `json:"access_all_environment"`
	Environments          []string          `json:"environments,omitempty"`
}

type ProjectMemberResponse struct {
	Member ProjectMember `json:"member"`
}

func getProjectMemberId(project string, memberType string, memberSlug string) string {
	return strings.Join([]string{project, memberType, memberSlug}, ".")
}

func parseProjectMemberId(id string) (project string, memberType string, memberSlug string, err error) {
	tokens := strings.Split(id, ".")
	if len(tokens) != 3 {
		return "", "", "", errors.New("invalid project member ID")
	}
	return tokens[0], tokens[1], tokens[2], nil
}

type IntegrationData = map[string]interface{}

type Integration struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type IntegrationResponse struct {
	Integration Integration `json:"integration"`
}

type SyncData = map[string]interface{}

type Sync struct {
	Slug        string `json:"slug"`
	Project     string `json:"project"`
	Config      string `json:"config"`
	Integration string `json:"integration"`
}

type SyncResponse struct {
	Sync Sync `json:"sync"`
}

type Environment struct {
	Slug      string `json:"slug"`
	Name      string `json:"name"`
	Project   string `json:"project"`
	CreatedAt string `json:"created_at"`
}

type EnvironmentResponse struct {
	Environment Environment `json:"environment"`
}

func (e Environment) getResourceId() string {
	return strings.Join([]string{e.Project, e.Slug}, ".")
}

func parseEnvironmentResourceId(id string) (project string, name string, err error) {
	tokens := strings.Split(id, ".")
	if len(tokens) != 2 {
		return "", "", errors.New("invalid environment ID")
	}
	return tokens[0], tokens[1], nil
}

type WebhookAuth struct {
	Type     string `json:"type"`
	Token    string `json:"token"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type Webhook struct {
	Slug           string   `json:"id"`
	Url            string   `json:"url"`
	Enabled        bool     `json:"enabled"`
	EnabledConfigs []string `json:"enabledConfigs"`
}

type WebhookResponse struct {
	Webhook Webhook `json:"webhook"`
}

type Config struct {
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Project     string `json:"project"`
	Environment string `json:"environment"`
	Locked      bool   `json:"locked"`
	Root        bool   `json:"root"`
	CreatedAt   string `json:"created_at"`
}

type ConfigResponse struct {
	Config Config `json:"config"`
}

func (c Config) getResourceId() string {
	return strings.Join([]string{c.Project, c.Environment, c.Name}, ".")
}

func parseConfigResourceId(id string) (project string, environment string, name string, err error) {
	tokens := strings.Split(id, ".")
	if len(tokens) != 3 {
		return "", "", "", errors.New("invalid config ID")
	}
	return tokens[0], tokens[1], tokens[2], nil
}

type ServiceToken struct {
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Project     string `json:"project"`
	Environment string `json:"environment"`
	Config      string `json:"config"`
	Access      string `json:"access"`
	Key         string `json:"key"`
	CreatedAt   string `json:"created_at"`
}

type ServiceTokenResponse struct {
	ServiceToken ServiceToken `json:"token"`
}

type ServiceTokenListResponse struct {
	ServiceTokens []ServiceToken `json:"tokens"`
}

func (t ServiceToken) getResourceId() string {
	return strings.Join([]string{t.Project, t.Config, t.Slug}, ".")
}

func parseServiceTokenResourceId(id string) (project string, config string, slug string, err error) {
	tokens := strings.Split(id, ".")
	if len(tokens) != 3 {
		return "", "", "", errors.New("invalid service token ID")
	}
	return tokens[0], tokens[1], tokens[2], nil
}

type ServiceAccountToken struct {
	Name      string `json:"name"`
	ExpiresAt string `json:"expires_at"`
	CreatedAt string `json:"created_at"`
	Slug      string `json:"slug"`
}

type ServiceAccountTokenResponse struct {
	ServiceAccountToken ServiceAccountToken `json:"api_token"`
	ApiKey              string              `json:"api_key"`
}

func (t ServiceAccountToken) getResourceId() string {
	return t.Slug
}

type WorkplaceRole struct {
	Name         string   `json:"name"`
	Permissions  []string `json:"permissions"`
	Identifier   string   `json:"identifier,omitempty"`
	IsCustomRole bool     `json:"is_custom_role"`
	IsInlineRole bool     `json:"is_inline_role"`
	CreatedAt    string   `json:"created_at"`
}

type ServiceAccount struct {
	Slug          string        `json:"slug"`
	Name          string        `json:"name"`
	CreatedAt     string        `json:"created_at"`
	WorkplaceRole WorkplaceRole `json:"workplace_role"`
}

type ServiceAccountResponse struct {
	ServiceAccount ServiceAccount `json:"service_account"`
}

type SimpleProjectRole struct {
	Identifier string `json:"identifier"`
}

type ProjectRole struct {
	Identifier   string   `json:"identifier"`
	Name         string   `json:"name"`
	Permissions  []string `json:"permissions"`
	CreatedAt    string   `json:"created_at"`
	IsCustomRole bool     `json:"is_custom_role"`
}

type GetProjectRoleResponse struct {
	Role ProjectRole `json:"role"`
}

type CreateProjectRoleResponse struct {
	Role ProjectRole `json:"role"`
}

type UpdateProjectRoleResponse struct {
	Role ProjectRole `json:"role"`
}

type Group struct {
	Slug               string            `json:"slug"`
	Name               string            `json:"name"`
	CreatedAt          string            `json:"created_at"`
	DefaultProjectRole SimpleProjectRole `json:"default_project_role"`
}

type GroupResponse struct {
	Group Group `json:"group"`
}

type GroupIsMemberResponse struct {
	IsMember bool `json:"isMember"`
}

type WorkplaceUser struct {
	Slug string `json:"id"`
}

type WorkplaceUsersListResponse struct {
	WorkplaceUsers []WorkplaceUser `json:"workplace_users"`
}

func getGroupMemberId(group string, memberType string, memberSlug string) string {
	return strings.Join([]string{group, memberType, memberSlug}, ".")
}

func parseGroupMemberId(id string) (group string, memberType string, memberSlug string, err error) {
	tokens := strings.Split(id, ".")
	if len(tokens) != 3 {
		return "", "", "", errors.New("invalid group member ID")
	}
	return tokens[0], tokens[1], tokens[2], nil
}
