package doppler

import (
	"encoding/json"
	"errors"
	"sort"
	"strings"
)

type RawSecret struct {
	Name  string
	Value *string
}

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
	Raw      string `json:"raw"`
	Computed string `json:"computed"`
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

type Project struct {
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
}

type ProjectResponse struct {
	Project Project `json:"project"`
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
