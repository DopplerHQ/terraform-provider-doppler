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

type TrustedIP struct {
	Project string `json:"project"`
	Config  string `json:"config"`
	IP      string `json:"ip"`
}

func (ip TrustedIP) getResourceId() string {
	props := strings.Join([]string{ip.Project, ip.Config}, ".")
	return strings.Join([]string{props, ip.IP}, "#")
}

func parseTrustedIPResourceId(id string) (project string, config string, ip string, err error) {
	tokens := strings.Split(id, "#")
	if len(tokens) != 2 {
		return "", "", "", errors.New("invalid trusted IP ID")
	}
	props := strings.Split(tokens[0], ".")
	if len(props) != 2 {
		return "", "", "", errors.New("invalid trusted IP ID")
	}
	return props[0], props[1], tokens[1], nil
}

type TrustedIPsListResponse struct {
	IPs []TrustedIP `json:"ips"`
}

type TrustedIPsAddResponse struct {
	IP TrustedIP `json:"ip"`
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

type Group struct {
	Slug               string            `json:"slug"`
	Name               string            `json:"name"`
	CreatedAt          string            `json:"created_at"`
	DefaultProjectRole SimpleProjectRole `json:"default_project_role"`
}

type GroupResponse struct {
	Group Group `json:"group"`
}
