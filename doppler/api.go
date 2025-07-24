package doppler

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type APIClient struct {
	Host      string
	APIKey    string
	VerifyTLS bool
}

type APIResponse struct {
	HTTPResponse *http.Response
	Body         []byte
}

type APIError struct {
	Err        error
	Message    string
	RetryAfter *time.Duration
	Response   *APIResponse
}

type ErrorResponse struct {
	Messages []string
	Success  bool
	Data     map[string]interface{}
}

type QueryParam struct {
	Key   string
	Value string
}

type PageOptions struct {
	Page    int
	PerPage int
}

const MAX_RETRIES = 10

func (e *APIError) Error() string {
	message := fmt.Sprintf("Doppler Error: %s", e.Message)
	if underlyingError := e.Err; underlyingError != nil {
		message = fmt.Sprintf("%s\n%s", message, underlyingError.Error())
	}
	return message
}

func isSuccess(statusCode int) bool {
	return (statusCode >= 200 && statusCode <= 299) || (statusCode >= 300 && statusCode <= 399)
}

func getSecondsDuration(seconds int) *time.Duration {
	duration := time.Duration(seconds) * time.Second
	return &duration
}

func (client APIClient) PerformRequestWithRetry(ctx context.Context, method string, path string, params []QueryParam, body []byte) (*APIResponse, error) {
	var lastErr error
	for i := 0; i < MAX_RETRIES; i++ {
		url := fmt.Sprintf("%s%s", client.Host, path)
		var bodyReader io.Reader
		if body != nil {
			bodyReader = bytes.NewReader(body)
		}
		req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
		if err != nil {
			return nil, &APIError{Err: err, Message: "Unable to form request"}
		}

		response, err := client.PerformRequest(req, params)
		lastErr = err
		if err == nil {
			return response, nil
		}
		apiError, isAPIError := err.(*APIError)
		if !isAPIError || apiError.RetryAfter == nil {
			return nil, err
		}
		time.Sleep(*apiError.RetryAfter)
	}
	return nil, lastErr
}

func (client APIClient) PerformRequest(req *http.Request, params []QueryParam) (*APIResponse, error) {
	httpClient := &http.Client{Timeout: 30 * time.Second}

	userAgent := fmt.Sprintf("terraform-provider-doppler/%s", ProviderVersion)
	req.Header.Set("user-agent", userAgent)
	req.SetBasicAuth(client.APIKey, "")
	if req.Header.Get("accept") == "" {
		req.Header.Set("accept", "application/json")
	}
	req.Header.Set("Content-Type", "application/json")

	query := req.URL.Query()
	for _, param := range params {
		query.Add(param.Key, param.Value)
	}
	req.URL.RawQuery = query.Encode()

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	if !client.VerifyTLS {
		tlsConfig.InsecureSkipVerify = true
	}

	httpClient.Transport = &http.Transport{
		DisableKeepAlives: true,
		TLSClientConfig:   tlsConfig,
	}

	r, err := httpClient.Do(req)
	if err != nil {
		var retryAfter *time.Duration
		if e, ok := err.(net.Error); ok && e.Timeout() {
			retryAfter = getSecondsDuration(1)
		}

		return nil, &APIError{Err: err, Message: "Unable to load response", RetryAfter: retryAfter}
	}
	defer func() {
		// G307: We're OK to ignore failures to close the body.
		// If the body wasn't read properly, we would have failed to serialize the response.
		_ = r.Body.Close()
	}()

	body, err := ioutil.ReadAll(r.Body)
	response := &APIResponse{HTTPResponse: r, Body: body}
	if err != nil {
		return response, &APIError{Err: err, Message: "Unable to load response data", Response: response}
	}

	if !isSuccess(r.StatusCode) {
		if contentType := r.Header.Get("content-type"); strings.HasPrefix(contentType, "application/json") {
			var errResponse ErrorResponse
			if err := json.Unmarshal(body, &errResponse); err != nil {
				return response, &APIError{Err: err, Message: "Unable to load response", Response: response}
			}

			var retryAfter *time.Duration
			if errResponse.Data["isRetryable"] == true {
				// Retry immediately
				retryAfter = getSecondsDuration(0)
			} else if retryableAfterSec, ok := errResponse.Data["isRetryableAfterSec"].(float64); ok {
				// Retry after specified time
				retryAfter = getSecondsDuration(int(retryableAfterSec))
			} else if r.StatusCode == 429 {
				retryAfterStr := r.Header.Get("retry-after")
				retryAfterInt, err := strconv.ParseInt(retryAfterStr, 10, 64)
				if err == nil {
					// Parse successful `retry-after` header result
					retryAfter = getSecondsDuration(int(retryAfterInt))
				} else {
					// There was some issue parsing, this shouldn't happen but retry after 1 second
					retryAfter = getSecondsDuration(1)
				}
			} else {
				// Otherwise, do not retry
				retryAfter = nil
			}
			return response, &APIError{
				Err:        nil,
				Message:    strings.Join(errResponse.Messages, "\n"),
				RetryAfter: retryAfter,
				Response:   response,
			}
		}
		return nil, &APIError{Err: fmt.Errorf("%d status code; %d bytes", r.StatusCode, len(body)), Message: "Unable to load response", Response: response}
	}
	if err != nil {
		return nil, &APIError{Err: err, Message: "Unable to parse response data", Response: response}
	}
	return response, nil
}

// Secrets

func (client APIClient) GetComputedSecrets(ctx context.Context, project string, config string) ([]ComputedSecret, error) {
	var params []QueryParam
	if project != "" {
		params = append(params, QueryParam{Key: "project", Value: project})
	}
	if config != "" {
		params = append(params, QueryParam{Key: "config", Value: config})
	}
	response, err := client.PerformRequestWithRetry(ctx, "GET", "/v3/configs/config/secrets/download", params, nil)
	if err != nil {
		return nil, err
	}
	result, err := ParseComputedSecrets(response.Body)
	if err != nil {
		return nil, &APIError{Err: err, Message: "Unable to parse secrets"}
	}
	return result, nil
}

func (client APIClient) GetSecret(ctx context.Context, project string, config string, secretName string) (*Secret, error) {
	var params []QueryParam
	if project != "" {
		params = append(params, QueryParam{Key: "project", Value: project})
	}
	if config != "" {
		params = append(params, QueryParam{Key: "config", Value: config})
	}
	params = append(params, QueryParam{Key: "name", Value: secretName})
	response, err := client.PerformRequestWithRetry(ctx, "GET", "/v3/configs/config/secret", params, nil)
	if err != nil {
		return nil, err
	}
	var result Secret
	if err := json.Unmarshal(response.Body, &result); err != nil {
		return nil, &APIError{Err: err, Message: "Unable to parse secret"}
	}
	if result.Value.Raw == nil &&
		result.Value.Computed == nil &&
		result.Value.RawVisibility == nil &&
		result.Value.ComputedVisibility == nil &&
		result.Value.RawValueType == nil &&
		result.Value.ComputedValueType == nil {
		return nil, &CustomNotFoundError{Message: "Secret does not exist"}
	}
	return &result, nil
}

func (client APIClient) UpdateSecrets(ctx context.Context, project string, config string, changeRequests []ChangeRequest) error {
	payload := map[string]interface{}{
		"change_requests": changeRequests,
	}
	if project != "" {
		payload["project"] = project
	}
	if config != "" {
		payload["config"] = config
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return &APIError{Err: err, Message: "Unable to parse secrets"}
	}
	_, err = client.PerformRequestWithRetry(ctx, "POST", "/v3/configs/config/secrets", []QueryParam{}, body)
	if err != nil {
		return err
	}
	return nil
}

// Projects

func (client APIClient) GetProject(ctx context.Context, name string) (*Project, error) {
	params := []QueryParam{
		{Key: "project", Value: name},
	}
	response, err := client.PerformRequestWithRetry(ctx, "GET", "/v3/projects/project", params, nil)
	if err != nil {
		return nil, err
	}
	var result ProjectResponse
	if err := json.Unmarshal(response.Body, &result); err != nil {
		return nil, &APIError{Err: err, Message: "Unable to parse project"}
	}
	return &result.Project, nil
}

func (client APIClient) CreateProject(ctx context.Context, name string, description string) (*Project, error) {
	payload := map[string]interface{}{
		"name":                        name,
		"create_default_environments": false,
	}
	if description != "" {
		payload["description"] = description
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, &APIError{Err: err, Message: "Unable to serialize project"}
	}
	response, err := client.PerformRequestWithRetry(ctx, "POST", "/v3/projects", []QueryParam{}, body)
	if err != nil {
		return nil, err
	}
	var result ProjectResponse
	if err = json.Unmarshal(response.Body, &result); err != nil {
		return nil, &APIError{Err: err, Message: "Unable to parse project"}
	}
	return &result.Project, nil
}

func (client APIClient) UpdateProject(ctx context.Context, currentName string, newName string, description string) (*Project, error) {
	payload := map[string]interface{}{
		"project":     currentName,
		"name":        newName,
		"description": description,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, &APIError{Err: err, Message: "Unable to serialize project"}
	}

	response, err := client.PerformRequestWithRetry(ctx, "POST", "/v3/projects/project", []QueryParam{}, body)
	if err != nil {
		return nil, err
	}

	var result ProjectResponse
	if err = json.Unmarshal(response.Body, &result); err != nil {
		return nil, &APIError{Err: err, Message: "Unable to parse project"}
	}
	return &result.Project, nil
}

func (client APIClient) DeleteProject(ctx context.Context, name string) error {
	payload := map[string]interface{}{
		"project": name,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return &APIError{Err: err, Message: "Unable to serialize project"}
	}

	_, err = client.PerformRequestWithRetry(ctx, "DELETE", "/v3/projects/project", []QueryParam{}, body)
	if err != nil {
		return err
	}

	return nil
}

// Project Roles

func (client APIClient) CreateProjectRole(ctx context.Context, name string, permissions []string) (*ProjectRole, error) {
	payload := map[string]interface{}{
		"name":        name,
		"permissions": permissions,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, &APIError{Err: err, Message: "Unable to serialize project role"}
	}
	response, err := client.PerformRequestWithRetry(ctx, "POST", "/v3/projects/roles", []QueryParam{}, body)
	if err != nil {
		return nil, err
	}
	var result CreateProjectRoleResponse
	if err = json.Unmarshal(response.Body, &result); err != nil {
		return nil, &APIError{Err: err, Message: "Unable to parse project role"}
	}
	return &result.Role, nil
}

func (client APIClient) GetProjectRole(ctx context.Context, identifier string) (*ProjectRole, error) {
	response, err := client.PerformRequestWithRetry(ctx, "GET", fmt.Sprintf("/v3/projects/roles/role/%s", url.PathEscape(identifier)), []QueryParam{}, nil)
	if err != nil {
		return nil, err
	}
	var result UpdateProjectRoleResponse
	if err = json.Unmarshal(response.Body, &result); err != nil {
		return nil, &APIError{Err: err, Message: "Unable to parse project role"}
	}
	return &result.Role, nil
}

func (client APIClient) UpdateProjectRole(ctx context.Context, identifier string, name string, permissions []string) (*ProjectRole, error) {
	payload := map[string]interface{}{
		"name":        name,
		"permissions": permissions,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, &APIError{Err: err, Message: "Unable to serialize project role"}
	}
	response, err := client.PerformRequestWithRetry(ctx, "PATCH", fmt.Sprintf("/v3/projects/roles/role/%s", url.PathEscape(identifier)), []QueryParam{}, body)
	if err != nil {
		return nil, err
	}
	var result UpdateProjectRoleResponse
	if err = json.Unmarshal(response.Body, &result); err != nil {
		return nil, &APIError{Err: err, Message: "Unable to parse project role"}
	}
	return &result.Role, nil
}

func (client APIClient) DeleteProjectRole(ctx context.Context, identifier string) error {
	_, err := client.PerformRequestWithRetry(ctx, "DELETE", fmt.Sprintf("/v3/projects/roles/role/%s", url.PathEscape(identifier)), []QueryParam{}, nil)
	if err != nil {
		return err
	}
	return nil
}

// Project Members
func (client APIClient) CreateProjectMember(ctx context.Context, project string, memberType string, memberSlug string, role string, environments []string) (*ProjectMember, error) {
	payload := map[string]interface{}{
		"project":      project,
		"slug":         memberSlug,
		"type":         memberType,
		"role":         role,
		"environments": environments,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, &APIError{Err: err, Message: "Unable to serialize project member"}
	}
	response, err := client.PerformRequestWithRetry(ctx, "POST", "/v3/projects/project/members", []QueryParam{}, body)
	if err != nil {
		return nil, err
	}
	var result ProjectMemberResponse
	if err = json.Unmarshal(response.Body, &result); err != nil {
		return nil, &APIError{Err: err, Message: "Unable to parse project member"}
	}
	return &result.Member, nil
}

func (client APIClient) GetProjectMember(ctx context.Context, project string, memberType string, memberSlug string) (*ProjectMember, error) {
	params := []QueryParam{
		{Key: "project", Value: project},
	}
	response, err := client.PerformRequestWithRetry(ctx, "GET", fmt.Sprintf("/v3/projects/project/members/member/%s/%s", url.QueryEscape(memberType), url.QueryEscape(memberSlug)), params, nil)
	if err != nil {
		return nil, err
	}
	var result ProjectMemberResponse
	if err = json.Unmarshal(response.Body, &result); err != nil {
		return nil, &APIError{Err: err, Message: "Unable to parse project member"}
	}
	return &result.Member, nil
}

func (client APIClient) UpdateProjectMember(ctx context.Context, project string, memberType string, memberSlug string, role *string, environments []string) (*ProjectMember, error) {
	payload := map[string]interface{}{
		"project": project,
	}
	if role != nil {
		payload["role"] = role
	}
	if environments != nil {
		payload["environments"] = environments
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, &APIError{Err: err, Message: "Unable to serialize project member update"}
	}
	response, err := client.PerformRequestWithRetry(ctx, "PATCH", fmt.Sprintf("/v3/projects/project/members/member/%s/%s", url.QueryEscape(memberType), url.QueryEscape(memberSlug)), []QueryParam{}, body)
	if err != nil {
		return nil, err
	}
	var result ProjectMemberResponse
	if err = json.Unmarshal(response.Body, &result); err != nil {
		return nil, &APIError{Err: err, Message: "Unable to parse project member"}
	}
	return &result.Member, nil
}

func (client APIClient) DeleteProjectMember(ctx context.Context, project string, memberType string, memberSlug string) error {
	params := []QueryParam{
		{Key: "project", Value: project},
	}
	_, err := client.PerformRequestWithRetry(ctx, "DELETE", fmt.Sprintf("/v3/projects/project/members/member/%s/%s", url.QueryEscape(memberType), url.QueryEscape(memberSlug)), params, nil)
	if err != nil {
		return err
	}
	return nil
}

// Integrations

func (client APIClient) GetIntegration(ctx context.Context, slug string) (*Integration, error) {
	params := []QueryParam{
		{Key: "integration", Value: slug},
	}
	response, err := client.PerformRequestWithRetry(ctx, "GET", "/v3/integrations/integration", params, nil)
	if err != nil {
		return nil, err
	}
	var result IntegrationResponse
	if err = json.Unmarshal(response.Body, &result); err != nil {
		return nil, &APIError{Err: err, Message: "Unable to parse integration"}
	}
	return &result.Integration, nil
}

func (client APIClient) CreateIntegration(ctx context.Context, data IntegrationData, name, integType string) (*Integration, error) {
	payload := map[string]interface{}{
		"name": name,
		"type": integType,
		"data": data,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, &APIError{Err: err, Message: "Unable to serialize integration"}
	}
	response, err := client.PerformRequestWithRetry(ctx, "POST", "/v3/integrations", []QueryParam{}, body)
	if err != nil {
		return nil, err
	}

	var result IntegrationResponse
	if err = json.Unmarshal(response.Body, &result); err != nil {
		return nil, &APIError{Err: err, Message: "Unable to parse integration"}
	}
	return &result.Integration, nil
}

func (client APIClient) UpdateIntegration(ctx context.Context, slug, name string, data IntegrationData) (*Integration, error) {
	params := []QueryParam{
		{Key: "integration", Value: slug},
	}

	payload := map[string]interface{}{}
	if name != "" {
		payload["name"] = name
	}
	if data != nil {
		payload["data"] = data
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, &APIError{Err: err, Message: "Unable to serialize integration"}
	}
	response, err := client.PerformRequestWithRetry(ctx, "PUT", "/v3/integrations/integration", params, body)
	if err != nil {
		return nil, err
	}
	var result IntegrationResponse
	if err = json.Unmarshal(response.Body, &result); err != nil {
		return nil, &APIError{Err: err, Message: "Unable to parse integration"}
	}
	return &result.Integration, nil
}

func (client APIClient) DeleteIntegration(ctx context.Context, name string) error {
	params := []QueryParam{
		{Key: "integration", Value: name},
	}
	_, err := client.PerformRequestWithRetry(ctx, "DELETE", "/v3/integrations/integration", params, nil)
	if err != nil {
		return err
	}
	return nil
}

// Syncs

func (client APIClient) GetSync(ctx context.Context, config, project, sync string) (*Sync, error) {
	params := []QueryParam{
		{Key: "config", Value: config},
		{Key: "project", Value: project},
		{Key: "sync", Value: sync},
	}
	response, err := client.PerformRequestWithRetry(ctx, "GET", "/v3/configs/config/syncs/sync", params, nil)
	if err != nil {
		return nil, err
	}
	var result SyncResponse
	if err = json.Unmarshal(response.Body, &result); err != nil {
		return nil, &APIError{Err: err, Message: "Unable to parse sync"}
	}
	return &result.Sync, nil
}

func (client APIClient) CreateSync(ctx context.Context, data SyncData, config, project, integration string) (*Sync, error) {
	params := []QueryParam{
		{Key: "config", Value: config},
		{Key: "project", Value: project},
	}
	payload := map[string]interface{}{
		"integration": integration,
		"data":        data,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, &APIError{Err: err, Message: "Unable to serialize sync"}
	}
	response, err := client.PerformRequestWithRetry(ctx, "POST", "/v3/configs/config/syncs", params, body)
	if err != nil {
		return nil, err
	}
	var result SyncResponse
	if err = json.Unmarshal(response.Body, &result); err != nil {
		return nil, &APIError{Err: err, Message: "Unable to parse sync"}
	}
	return &result.Sync, nil
}

func (client APIClient) DeleteSync(ctx context.Context, slug string, deleteTarget bool, config, project string) error {
	params := []QueryParam{
		{Key: "config", Value: config},
		{Key: "project", Value: project},
		{Key: "sync", Value: slug},
		{Key: "delete_from_target", Value: strconv.FormatBool(deleteTarget)},
	}
	_, err := client.PerformRequestWithRetry(ctx, "DELETE", "/v3/configs/config/syncs/sync", params, nil)
	if err != nil {
		return err
	}
	return nil
}

// Rotated Secrets

func (client APIClient) GetRotatedSecret(ctx context.Context, config, project, slug string) (*RotatedSecret, error) {
	params := []QueryParam{
		{Key: "config", Value: config},
		{Key: "project", Value: project},
		{Key: "slug", Value: slug},
	}
	response, err := client.PerformRequestWithRetry(ctx, "GET", "/v3/configs/config/rotated_secrets/rotated_secret", params, nil)
	if err != nil {
		return nil, err
	}
	var result RotatedSecretResponse
	if err = json.Unmarshal(response.Body, &result); err != nil {
		return nil, &APIError{Err: err, Message: "Unable to parse rotated secret"}
	}
	return &result.RotatedSecret, nil
}

func (client APIClient) CreateExternalId(ctx context.Context, integrationType string) (string, error) {
	payload := map[string]interface{}{
		"type": integrationType,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", &APIError{Err: err, Message: "Unable to serialize external id request"}
	}
	response, err := client.PerformRequestWithRetry(ctx, "POST", "/v3/integrations/generate_external_id", nil, body)
	if err != nil {
		return "", err
	}
	var result ExternalIdResponse
	if err = json.Unmarshal(response.Body, &result); err != nil {
		return "", &APIError{Err: err, Message: "Unable to parse external ID"}
	}
	return result.ExternalId, nil
}

func (client APIClient) CreateRotatedSecret(ctx context.Context, name string, rotationPeriodSec int, parameters RotatedSecretParameters, credentials RotatedSecretCredentials, config, project, integration string) (*RotatedSecret, error) {
	params := []QueryParam{
		{Key: "config", Value: config},
		{Key: "project", Value: project},
	}
	payload := map[string]interface{}{
		"integration":         integration,
		"name":                name,
		"rotation_period_sec": rotationPeriodSec,
		"parameters":          parameters,
		"credentials":         credentials,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, &APIError{Err: err, Message: "Unable to serialize rotated secret"}
	}
	response, err := client.PerformRequestWithRetry(ctx, "POST", "/v3/configs/config/rotated_secrets", params, body)
	if err != nil {
		return nil, err
	}
	var result RotatedSecretResponse
	if err = json.Unmarshal(response.Body, &result); err != nil {
		return nil, &APIError{Err: err, Message: "Unable to parse rotated secret"}
	}
	return &result.RotatedSecret, nil
}

func (client APIClient) UpdateRotatedSecret(ctx context.Context, name string, rotationPeriodSec int, config, project, slug string) (*RotatedSecret, error) {
	params := []QueryParam{
		{Key: "config", Value: config},
		{Key: "project", Value: project},
		{Key: "slug", Value: slug},
	}
	payload := map[string]interface{}{
		"name":                name,
		"rotation_period_sec": rotationPeriodSec,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, &APIError{Err: err, Message: "Unable to serialize rotated secret"}
	}
	response, err := client.PerformRequestWithRetry(ctx, "PUT", "/v3/configs/config/rotated_secrets/rotated_secret", params, body)
	if err != nil {
		return nil, err
	}
	var result RotatedSecretResponse
	if err = json.Unmarshal(response.Body, &result); err != nil {
		return nil, &APIError{Err: err, Message: "Unable to parse rotated secret"}
	}
	return &result.RotatedSecret, nil
}

func (client APIClient) DeleteRotatedSecret(ctx context.Context, slug string, config, project string) error {
	params := []QueryParam{
		{Key: "config", Value: config},
		{Key: "project", Value: project},
		{Key: "slug", Value: slug},
	}
	_, err := client.PerformRequestWithRetry(ctx, "DELETE", "/v3/configs/config/rotated_secrets/rotated_secret", params, nil)
	if err != nil {
		return err
	}
	return nil
}

// Environments

func (client APIClient) GetEnvironment(ctx context.Context, project string, name string) (*Environment, error) {
	params := []QueryParam{
		{Key: "project", Value: project},
		{Key: "environment", Value: name},
	}
	response, err := client.PerformRequestWithRetry(ctx, "GET", "/v3/environments/environment", params, nil)
	if err != nil {
		return nil, err
	}
	var result EnvironmentResponse
	if err = json.Unmarshal(response.Body, &result); err != nil {
		return nil, &APIError{Err: err, Message: "Unable to parse environment"}
	}
	return &result.Environment, nil
}

func (client APIClient) CreateEnvironment(ctx context.Context, project string, slug string, name string, personalConfigs bool) (*Environment, error) {
	payload := map[string]interface{}{
		"project":          project,
		"name":             name,
		"slug":             slug,
		"personal_configs": personalConfigs,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, &APIError{Err: err, Message: "Unable to serialize environment"}
	}

	response, err := client.PerformRequestWithRetry(ctx, "POST", "/v3/environments", []QueryParam{}, body)
	if err != nil {
		return nil, err
	}

	var result EnvironmentResponse
	if err = json.Unmarshal(response.Body, &result); err != nil {
		return nil, &APIError{Err: err, Message: "Unable to parse environment"}
	}
	return &result.Environment, nil
}

func (client APIClient) UpdateEnvironment(ctx context.Context, project string, currentSlug string, newSlug string, newName string, personalConfigs bool) (*Environment, error) {
	params := []QueryParam{
		{Key: "project", Value: project},
		{Key: "environment", Value: currentSlug},
	}
	payload := map[string]interface{}{
		"slug":             newSlug,
		"name":             newName,
		"personal_configs": personalConfigs,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, &APIError{Err: err, Message: "Unable to serialize environment"}
	}
	response, err := client.PerformRequestWithRetry(ctx, "PUT", "/v3/environments/environment", params, body)
	if err != nil {
		return nil, err
	}
	var result EnvironmentResponse
	if err = json.Unmarshal(response.Body, &result); err != nil {
		return nil, &APIError{Err: err, Message: "Unable to parse project"}
	}
	return &result.Environment, nil
}

func (client APIClient) DeleteEnvironment(ctx context.Context, project string, slug string) error {
	params := []QueryParam{
		{Key: "project", Value: project},
		{Key: "environment", Value: slug},
	}
	_, err := client.PerformRequestWithRetry(ctx, "DELETE", "/v3/environments/environment", params, nil)
	if err != nil {
		return err
	}
	return nil
}

func (client APIClient) ListEnvironments(ctx context.Context, project string) ([]Environment, error) {
	params := []QueryParam{
		{Key: "project", Value: project},
	}
	response, err := client.PerformRequestWithRetry(ctx, "GET", "/v3/environments", params, nil)
	if err != nil {
		return nil, err
	}
	var result EnvironmentsResponse
	if err = json.Unmarshal(response.Body, &result); err != nil {
		return nil, &APIError{Err: err, Message: "Unable to parse environments"}
	}
	return result.Environments, nil
}

// Webhooks

func (client APIClient) GetWebhook(ctx context.Context, project string, slug string) (*Webhook, error) {
	params := []QueryParam{
		{Key: "project", Value: project},
	}
	response, err := client.PerformRequestWithRetry(ctx, "GET", fmt.Sprintf("/v3/webhooks/webhook/%s", url.QueryEscape(slug)), params, nil)
	if err != nil {
		return nil, err
	}
	var result WebhookResponse
	if err = json.Unmarshal(response.Body, &result); err != nil {
		return nil, &APIError{Err: err, Message: "Unable to parse webhook"}
	}
	return &result.Webhook, nil
}

type CreateWebhookOptionalParameters struct {
	Secret         string
	Auth           *WebhookAuth
	WebhookPayload string
	EnabledConfigs []string
	Name           string
}

func (client APIClient) CreateWebhook(ctx context.Context, project string, url string, enabled bool, options *CreateWebhookOptionalParameters) (*Webhook, error) {
	params := []QueryParam{
		{Key: "project", Value: project},
	}

	payload := map[string]interface{}{
		"url":     url,
		"enabled": enabled,
	}

	if options != nil {
		if options.Secret != "" {
			payload["secret"] = options.Secret
		}
		if options.Auth != nil {
			payload["authentication"] = *options.Auth
		}
		if options.WebhookPayload != "" {
			payload["payload"] = options.WebhookPayload
		}
		if options.EnabledConfigs != nil {
			payload["enableConfigs"] = options.EnabledConfigs
		}
		if options.Name != "" {
			payload["name"] = options.Name
		}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, &APIError{Err: err, Message: "Unable to serialize webhook"}
	}

	response, err := client.PerformRequestWithRetry(ctx, "POST", "/v3/webhooks", params, body)
	if err != nil {
		return nil, err
	}

	var result WebhookResponse
	if err = json.Unmarshal(response.Body, &result); err != nil {
		return nil, &APIError{Err: err, Message: "Unable to parse webhook"}
	}
	return &result.Webhook, nil
}

func (client APIClient) EnableWebhook(ctx context.Context, project string, slug string) (*Webhook, error) {
	params := []QueryParam{
		{Key: "project", Value: project},
	}
	response, err := client.PerformRequestWithRetry(ctx, "POST", fmt.Sprintf("/v3/webhooks/webhook/%s/enable", url.QueryEscape(slug)), params, nil)
	if err != nil {
		return nil, err
	}

	var result WebhookResponse
	if err = json.Unmarshal(response.Body, &result); err != nil {
		return nil, &APIError{Err: err, Message: "Unable to parse webhook"}
	}
	return &result.Webhook, nil
}

func (client APIClient) DisableWebhook(ctx context.Context, project string, slug string) (*Webhook, error) {
	params := []QueryParam{
		{Key: "project", Value: project},
	}
	response, err := client.PerformRequestWithRetry(ctx, "POST", fmt.Sprintf("/v3/webhooks/webhook/%s/disable", url.QueryEscape(slug)), params, nil)
	if err != nil {
		return nil, err
	}

	var result WebhookResponse
	if err = json.Unmarshal(response.Body, &result); err != nil {
		return nil, &APIError{Err: err, Message: "Unable to parse webhook"}
	}
	return &result.Webhook, nil
}

func (client APIClient) UpdateWebhook(ctx context.Context, project string, slug string, webhookUrl string, secret string, webhookPayload string, webhookName string, enabledConfigs []string, disabledConfigs []string, auth WebhookAuth) (*Webhook, error) {
	params := []QueryParam{
		{Key: "project", Value: project},
	}

	payload := map[string]interface{}{}
	payload["url"] = webhookUrl
	payload["secret"] = secret
	payload["payload"] = webhookPayload
	if webhookName != "" {
		payload["name"] = webhookName
	} else {
		payload["name"] = nil
	}
	payload["enableConfigs"] = enabledConfigs
	payload["disableConfigs"] = disabledConfigs
	payload["authentication"] = auth

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, &APIError{Err: err, Message: "Unable to serialize webhook"}
	}

	response, err := client.PerformRequestWithRetry(ctx, "PATCH", fmt.Sprintf("/v3/webhooks/webhook/%s", url.QueryEscape(slug)), params, body)
	if err != nil {
		return nil, err
	}

	var result WebhookResponse
	if err = json.Unmarshal(response.Body, &result); err != nil {
		return nil, &APIError{Err: err, Message: "Unable to parse webhook"}
	}
	return &result.Webhook, nil
}

func (client APIClient) DeleteWebhook(ctx context.Context, project string, slug string) error {
	params := []QueryParam{
		{Key: "project", Value: project},
	}
	_, err := client.PerformRequestWithRetry(ctx, "DELETE", fmt.Sprintf("/v3/webhooks/webhook/%s", url.QueryEscape(slug)), params, nil)
	if err != nil {
		return err
	}
	return nil
}

// Configs

func (client APIClient) GetConfig(ctx context.Context, project string, name string) (*Config, error) {
	params := []QueryParam{
		{Key: "project", Value: project},
		{Key: "config", Value: name},
	}
	response, err := client.PerformRequestWithRetry(ctx, "GET", "/v3/configs/config", params, nil)
	if err != nil {
		return nil, err
	}
	var result ConfigResponse
	if err = json.Unmarshal(response.Body, &result); err != nil {
		return nil, &APIError{Err: err, Message: "Unable to parse config"}
	}
	return &result.Config, nil
}

func (client APIClient) CreateConfig(ctx context.Context, project string, environment string, name string) (*Config, error) {
	payload := map[string]interface{}{
		"project":     project,
		"environment": environment,
		"name":        name,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, &APIError{Err: err, Message: "Unable to serialize config"}
	}
	response, err := client.PerformRequestWithRetry(ctx, "POST", "/v3/configs", []QueryParam{}, body)
	if err != nil {
		return nil, err
	}
	var result ConfigResponse
	if err = json.Unmarshal(response.Body, &result); err != nil {
		return nil, &APIError{Err: err, Message: "Unable to parse config"}
	}
	return &result.Config, nil
}

func (client APIClient) RenameConfig(ctx context.Context, project string, currentName string, newName string) (*Config, error) {
	payload := map[string]interface{}{
		"project": project,
		"config":  currentName,
		"name":    newName,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, &APIError{Err: err, Message: "Unable to serialize config"}
	}
	response, err := client.PerformRequestWithRetry(ctx, "POST", "/v3/configs/config", []QueryParam{}, body)
	if err != nil {
		return nil, err
	}
	var result ConfigResponse
	if err = json.Unmarshal(response.Body, &result); err != nil {
		return nil, &APIError{Err: err, Message: "Unable to parse config"}
	}
	return &result.Config, nil
}

func (client APIClient) UpdateConfigInheritable(ctx context.Context, project string, config string, inheritable bool) (*Config, error) {
	payload := map[string]interface{}{
		"project":     project,
		"config":      config,
		"inheritable": inheritable,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, &APIError{Err: err, Message: "Unable to serialize config"}
	}
	response, err := client.PerformRequestWithRetry(ctx, "POST", "/v3/configs/config/inheritable", []QueryParam{}, body)
	if err != nil {
		return nil, err
	}
	var result ConfigResponse
	if err = json.Unmarshal(response.Body, &result); err != nil {
		return nil, &APIError{Err: err, Message: "Unable to parse config"}
	}
	return &result.Config, nil
}

func (client APIClient) UpdateConfigInherits(ctx context.Context, project string, config string, inherits []ConfigDescriptor) (*Config, error) {
	payload := map[string]interface{}{
		"project":  project,
		"config":   config,
		"inherits": inherits,
	}

	if len(inherits) == 0 {
		// If we don't manually instantiate an empty array here, go will marshal the inherits property as nil instead of []
		payload["inherits"] = []ConfigDescriptor{}
	}

	body, err := json.Marshal(payload)

	if err != nil {
		return nil, &APIError{Err: err, Message: "Unable to serialize config"}
	}
	response, err := client.PerformRequestWithRetry(ctx, "POST", "/v3/configs/config/inherits", []QueryParam{}, body)
	if err != nil {
		return nil, err
	}
	var result ConfigResponse
	if err = json.Unmarshal(response.Body, &result); err != nil {
		return nil, &APIError{Err: err, Message: "Unable to parse config"}
	}
	return &result.Config, nil
}

func (client APIClient) DeleteConfig(ctx context.Context, project string, name string) error {
	payload := map[string]interface{}{
		"project": project,
		"config":  name,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return &APIError{Err: err, Message: "Unable to serialize config"}
	}
	_, err = client.PerformRequestWithRetry(ctx, "DELETE", "/v3/configs/config", []QueryParam{}, body)
	if err != nil {
		return err
	}
	return nil
}

// Service Tokens

func (client APIClient) GetServiceTokens(ctx context.Context, project string, config string) ([]ServiceToken, error) {
	params := []QueryParam{
		{Key: "project", Value: project},
		{Key: "config", Value: config},
	}
	response, err := client.PerformRequestWithRetry(ctx, "GET", "/v3/configs/config/tokens", params, nil)
	if err != nil {
		return nil, err
	}
	var result ServiceTokenListResponse
	if err = json.Unmarshal(response.Body, &result); err != nil {
		return nil, &APIError{Err: err, Message: "Unable to parse service tokens"}
	}
	return result.ServiceTokens, nil
}

func (client APIClient) CreateServiceToken(ctx context.Context, project string, config string, access string, name string) (*ServiceToken, error) {
	payload := map[string]interface{}{
		"project": project,
		"config":  config,
		"access":  access,
		"name":    name,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, &APIError{Err: err, Message: "Unable to serialize service token"}
	}
	response, err := client.PerformRequestWithRetry(ctx, "POST", "/v3/configs/config/tokens", []QueryParam{}, body)
	if err != nil {
		return nil, err
	}
	var result ServiceTokenResponse
	if err = json.Unmarshal(response.Body, &result); err != nil {
		return nil, &APIError{Err: err, Message: "Unable to parse service token"}
	}
	return &result.ServiceToken, nil
}

func (client APIClient) DeleteServiceToken(ctx context.Context, project string, config string, slug string) error {
	payload := map[string]interface{}{
		"project": project,
		"config":  config,
		"slug":    slug,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return &APIError{Err: err, Message: "Unable to serialize config"}
	}
	_, err = client.PerformRequestWithRetry(ctx, "DELETE", "/v3/configs/config/tokens/token", []QueryParam{}, body)
	if err != nil {
		return err
	}
	return nil
}

// Service Accounts

func (client APIClient) GetServiceAccount(ctx context.Context, slug string) (*ServiceAccount, error) {
	response, err := client.PerformRequestWithRetry(ctx, "GET", fmt.Sprintf("/v3/workplace/service_accounts/service_account/%s", url.QueryEscape(slug)), []QueryParam{}, nil)
	if err != nil {
		return nil, err
	}
	var result ServiceAccountResponse
	if err = json.Unmarshal(response.Body, &result); err != nil {
		return nil, &APIError{Err: err, Message: "Unable to parse service account"}
	}
	return &result.ServiceAccount, nil
}

func (client APIClient) CreateServiceAccount(ctx context.Context, name string, workplaceRole string, workplacePermissions []string) (*ServiceAccount, error) {
	payload := map[string]interface{}{
		"name": name,
	}
	if workplaceRole != "" {
		payload["workplace_role"] = map[string]string{"identifier": workplaceRole}
	} else if workplacePermissions != nil {
		payload["workplace_role"] = map[string][]string{"permissions": workplacePermissions}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, &APIError{Err: err, Message: "Unable to serialize service account"}
	}
	response, err := client.PerformRequestWithRetry(ctx, "POST", "/v3/workplace/service_accounts", []QueryParam{}, body)
	if err != nil {
		return nil, err
	}
	var result ServiceAccountResponse
	if err = json.Unmarshal(response.Body, &result); err != nil {
		return nil, &APIError{Err: err, Message: "Unable to parse service account"}
	}
	return &result.ServiceAccount, nil
}

func (client APIClient) UpdateServiceAccount(ctx context.Context, slug string, name string, workplaceRole string, workplacePermissions []string) (*ServiceAccount, error) {
	payload := map[string]interface{}{}
	if name != "" {
		payload["name"] = name
	}
	if workplaceRole != "" {
		payload["workplace_role"] = map[string]string{"identifier": workplaceRole}
	} else if workplacePermissions != nil {
		payload["workplace_role"] = map[string][]string{"permissions": workplacePermissions}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, &APIError{Err: err, Message: "Unable to serialize service account"}
	}
	response, err := client.PerformRequestWithRetry(ctx, "PATCH", fmt.Sprintf("/v3/workplace/service_accounts/service_account/%s", url.QueryEscape(slug)), []QueryParam{}, body)
	if err != nil {
		return nil, err
	}
	var result ServiceAccountResponse
	if err = json.Unmarshal(response.Body, &result); err != nil {
		return nil, &APIError{Err: err, Message: "Unable to parse service account"}
	}
	return &result.ServiceAccount, nil
}

func (client APIClient) DeleteServiceAccount(ctx context.Context, slug string) error {
	_, err := client.PerformRequestWithRetry(ctx, "DELETE", fmt.Sprintf("/v3/workplace/service_accounts/service_account/%s", url.QueryEscape(slug)), []QueryParam{}, nil)
	if err != nil {
		return err
	}
	return nil
}

// Service Account Tokens

func (client APIClient) GetServiceAccountToken(ctx context.Context, serviceAccountSlug string, slug string) (ServiceAccountTokenResponse, error) {
	response, err := client.PerformRequestWithRetry(ctx, "GET", fmt.Sprintf("/v3/workplace/service_accounts/service_account/%s/tokens/token/%s", url.QueryEscape(serviceAccountSlug), url.QueryEscape(slug)), []QueryParam{}, nil)
	if err != nil {
		return ServiceAccountTokenResponse{}, err
	}
	var result ServiceAccountTokenResponse
	if err = json.Unmarshal(response.Body, &result); err != nil {
		return ServiceAccountTokenResponse{}, &APIError{Err: err, Message: "Unable to parse service account tokens"}
	}
	return result, nil
}

func (client APIClient) CreateServiceAccountToken(ctx context.Context, serviceAccountSlug string, name string, expiresAt string) (*ServiceAccountTokenResponse, error) {
	payload := map[string]interface{}{
		"name": name,
	}
	if expiresAt != "" {
		payload["expires_at"] = expiresAt
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, &APIError{Err: err, Message: "Unable to serialize account service token"}
	}
	response, err := client.PerformRequestWithRetry(ctx, "POST", fmt.Sprintf("/v3/workplace/service_accounts/service_account/%s/tokens", url.QueryEscape(serviceAccountSlug)), []QueryParam{}, body)
	if err != nil {
		return nil, err
	}
	var result ServiceAccountTokenResponse
	if err = json.Unmarshal(response.Body, &result); err != nil {
		return nil, &APIError{Err: err, Message: "Unable to parse service account token"}
	}
	return &result, nil
}

func (client APIClient) DeleteServiceAccountToken(ctx context.Context, serviceAccountSlug string, slug string) error {
	_, err := client.PerformRequestWithRetry(ctx, "DELETE", fmt.Sprintf("/v3/workplace/service_accounts/service_account/%s/tokens/token/%s", url.QueryEscape(serviceAccountSlug), url.QueryEscape(slug)), []QueryParam{}, nil)
	if err != nil {
		return err
	}
	return nil
}

// Groups

func (client APIClient) GetGroup(ctx context.Context, slug string) (*Group, error) {
	response, err := client.PerformRequestWithRetry(ctx, "GET", fmt.Sprintf("/v3/workplace/groups/group/%s", url.QueryEscape(slug)), []QueryParam{}, nil)
	if err != nil {
		return nil, err
	}
	var result GroupResponse
	if err = json.Unmarshal(response.Body, &result); err != nil {
		return nil, &APIError{Err: err, Message: "Unable to parse group"}
	}
	return &result.Group, nil
}

func (client APIClient) GetGroupByName(ctx context.Context, name string) (*Group, error) {
	params := []QueryParam{
		{Key: "name", Value: name},
	}
	response, err := client.PerformRequestWithRetry(ctx, "GET", "/v3/workplace/groups", params, nil)
	if err != nil {
		return nil, err
	}
	var result GroupsResponse
	if err = json.Unmarshal(response.Body, &result); err != nil {
		return nil, &APIError{Err: err, Message: "Unable to parse group"}
	}
	if len(result.Groups) > 1 {
		return nil, &APIError{Err: err, Message: "Multiple workplace groups returned"}
	}
	if len(result.Groups) > 0 {
		return &result.Groups[0], nil
	} else {
		return nil, &CustomNotFoundError{Message: "Could not find requested workplace group"}
	}
}

func (client APIClient) CreateGroup(ctx context.Context, name string, defaultProjectRole string) (*Group, error) {
	payload := map[string]interface{}{
		"name": name,
	}
	if defaultProjectRole != "" {
		payload["default_project_role"] = defaultProjectRole
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, &APIError{Err: err, Message: "Unable to serialize group"}
	}
	response, err := client.PerformRequestWithRetry(ctx, "POST", "/v3/workplace/groups", []QueryParam{}, body)
	if err != nil {
		return nil, err
	}
	var result GroupResponse
	if err = json.Unmarshal(response.Body, &result); err != nil {
		return nil, &APIError{Err: err, Message: "Unable to parse group"}
	}
	return &result.Group, nil
}

func (client APIClient) UpdateGroup(ctx context.Context, slug string, name string, defaultProjectRole *string) (*Group, error) {
	payload := map[string]interface{}{}
	if name != "" {
		payload["name"] = name
	}
	if defaultProjectRole != nil {
		if *defaultProjectRole == "" {
			payload["default_project_role"] = nil
		} else {
			payload["default_project_role"] = defaultProjectRole
		}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, &APIError{Err: err, Message: "Unable to serialize group"}
	}
	response, err := client.PerformRequestWithRetry(ctx, "PATCH", fmt.Sprintf("/v3/workplace/groups/group/%s", url.QueryEscape(slug)), []QueryParam{}, body)
	if err != nil {
		return nil, err
	}
	var result GroupResponse
	if err = json.Unmarshal(response.Body, &result); err != nil {
		return nil, &APIError{Err: err, Message: "Unable to parse group"}
	}
	return &result.Group, nil
}

func (client APIClient) DeleteGroup(ctx context.Context, slug string) error {
	_, err := client.PerformRequestWithRetry(ctx, "DELETE", fmt.Sprintf("/v3/workplace/groups/group/%s", url.QueryEscape(slug)), []QueryParam{}, nil)
	if err != nil {
		return err
	}
	return nil
}

// Group Members

func (client APIClient) CreateGroupMember(ctx context.Context, group string, memberType string, memberSlug string) error {
	payload := map[string]interface{}{
		"type": memberType,
		"slug": memberSlug,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return &APIError{Err: err, Message: "Unable to serialize group member"}
	}
	_, err = client.PerformRequestWithRetry(ctx, "POST", fmt.Sprintf("/v3/workplace/groups/group/%s/members", url.QueryEscape(group)), []QueryParam{}, body)
	if err != nil {
		return err
	}
	return nil
}

func (client APIClient) GetGroupMember(ctx context.Context, group string, memberType string, memberSlug string) error {
	_, err := client.PerformRequestWithRetry(ctx, "GET", fmt.Sprintf("/v3/workplace/groups/group/%s/members/%s/%s", url.QueryEscape(group), url.QueryEscape(memberType), url.QueryEscape(memberSlug)), []QueryParam{}, nil)
	return err
}

func (client APIClient) DeleteGroupMember(ctx context.Context, group string, memberType string, memberSlug string) error {
	_, err := client.PerformRequestWithRetry(ctx, "DELETE", fmt.Sprintf("/v3/workplace/groups/group/%s/members/%s/%s", url.QueryEscape(group), url.QueryEscape(memberType), url.QueryEscape(memberSlug)), []QueryParam{}, nil)
	if err != nil {
		return err
	}
	return nil
}

func (client APIClient) GetGroupMembers(ctx context.Context, group string, pageOptions PageOptions) ([]GroupMember, error) {
	params := []QueryParam{
		{Key: "page", Value: strconv.Itoa(pageOptions.Page)},
		{Key: "per_page", Value: strconv.Itoa(pageOptions.PerPage)},
	}
	response, err := client.PerformRequestWithRetry(ctx, "GET", fmt.Sprintf("/v3/workplace/groups/group/%s/members", url.QueryEscape(group)), params, nil)
	if err != nil {
		return nil, err
	}
	var result GetGroupMembersResponse
	if err = json.Unmarshal(response.Body, &result); err != nil {
		return nil, &APIError{Err: err, Message: "Unable to parse group members"}
	}
	return result.Members, nil
}

func (client APIClient) ReplaceGroupMembers(ctx context.Context, group string, members []GroupMember) error {
	payload := map[string]interface{}{
		"members": members,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return &APIError{Err: err, Message: "Unable to serialize group members"}
	}
	_, err = client.PerformRequestWithRetry(ctx, "PUT", fmt.Sprintf("/v3/workplace/groups/group/%s/members", url.QueryEscape(group)), []QueryParam{}, body)
	if err != nil {
		return err
	}
	return nil
}

// Workplace Users

func (client APIClient) GetWorkplaceUser(ctx context.Context, email string) (*WorkplaceUser, error) {
	params := []QueryParam{
		{Key: "email", Value: email},
	}
	response, err := client.PerformRequestWithRetry(ctx, "GET", "/v3/workplace/users", params, nil)
	if err != nil {
		return nil, err
	}
	var result WorkplaceUsersListResponse
	if err = json.Unmarshal(response.Body, &result); err != nil {
		return nil, &APIError{Err: err, Message: "Unable to parse workplace user"}
	}
	if len(result.WorkplaceUsers) > 1 {
		return nil, &APIError{Err: err, Message: "Multiple workplace users returned"}
	}
	if len(result.WorkplaceUsers) == 0 {
		return nil, &CustomNotFoundError{Message: "Could not find requested workplace user"}
	}
	return &result.WorkplaceUsers[0], nil
}

// Change Request Policies

func (client APIClient) GetChangeRequestPolicy(ctx context.Context, slug string) (*ChangeRequestPolicy, error) {
	response, err := client.PerformRequestWithRetry(ctx, "GET", fmt.Sprintf("/v3/workplace/change_request_policies/change_request_policy/%s", url.QueryEscape(slug)), nil, nil)
	if err != nil {
		return nil, err
	}
	var result ChangeRequestPolicyResponse
	if err = json.Unmarshal(response.Body, &result); err != nil {
		return nil, &APIError{Err: err, Message: "Unable to parse change request policy"}
	}
	return &result.Policy, nil
}

func (client APIClient) CreateChangeRequestPolicy(ctx context.Context, policy *ChangeRequestPolicy) (*ChangeRequestPolicy, error) {
	body, err := json.Marshal(policy)
	if err != nil {
		return nil, &APIError{Err: err, Message: "Unable to serialize change request policy"}
	}

	response, err := client.PerformRequestWithRetry(ctx, "POST", "/v3/workplace/change_request_policies", nil, body)
	if err != nil {
		return nil, err
	}

	var result ChangeRequestPolicyResponse
	if err = json.Unmarshal(response.Body, &result); err != nil {
		return nil, &APIError{Err: err, Message: "Unable to parse change request policy"}
	}
	return &result.Policy, nil
}

func (client APIClient) UpdateChangeRequestPolicy(ctx context.Context, policy *ChangeRequestPolicy) (*ChangeRequestPolicy, error) {
	body, err := json.Marshal(policy)
	if err != nil {
		return nil, &APIError{Err: err, Message: "Unable to serialize change request policy"}
	}

	response, err := client.PerformRequestWithRetry(ctx, "POST", fmt.Sprintf("/v3/workplace/change_request_policies/change_request_policy/%s", url.QueryEscape(policy.Slug)), nil, body)
	if err != nil {
		return nil, err
	}

	var result ChangeRequestPolicyResponse
	if err = json.Unmarshal(response.Body, &result); err != nil {
		return nil, &APIError{Err: err, Message: "Unable to parse change request policy"}
	}
	return &result.Policy, nil
}

func (client APIClient) DeleteChangeRequestPolicy(ctx context.Context, slug string) error {
	_, err := client.PerformRequestWithRetry(ctx, "DELETE", fmt.Sprintf("/v3/workplace/change_request_policies/change_request_policy/%s", url.QueryEscape(slug)), nil, nil)
	if err != nil {
		return err
	}
	return nil
}

// Workplace Roles

func (client APIClient) CreateWorkplaceRole(ctx context.Context, name string, permissions []string) (*WorkplaceRole, error) {
	payload := map[string]interface{}{
		"name":        name,
		"permissions": permissions,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, &APIError{Err: err, Message: "Unable to serialize workplace role"}
	}
	response, err := client.PerformRequestWithRetry(ctx, "POST", "/v3/workplace/roles", []QueryParam{}, body)
	if err != nil {
		return nil, err
	}
	var result CreateWorkplaceRoleResponse
	if err = json.Unmarshal(response.Body, &result); err != nil {
		return nil, &APIError{Err: err, Message: "Unable to parse workplace role"}
	}
	return &result.Role, nil
}

func (client APIClient) GetWorkplaceRole(ctx context.Context, identifier string) (*WorkplaceRole, error) {
	response, err := client.PerformRequestWithRetry(ctx, "GET", fmt.Sprintf("/v3/workplace/roles/role/%s", url.PathEscape(identifier)), []QueryParam{}, nil)
	if err != nil {
		return nil, err
	}
	var result GetWorkplaceRoleResponse
	if err = json.Unmarshal(response.Body, &result); err != nil {
		return nil, &APIError{Err: err, Message: "Unable to parse workplace role"}
	}
	return &result.Role, nil
}

func (client APIClient) UpdateWorkplaceRole(ctx context.Context, identifier string, name string, permissions []string) (*WorkplaceRole, error) {
	payload := map[string]interface{}{
		"name":        name,
		"permissions": permissions,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, &APIError{Err: err, Message: "Unable to serialize workplace role"}
	}
	response, err := client.PerformRequestWithRetry(ctx, "PATCH", fmt.Sprintf("/v3/workplace/roles/role/%s", url.PathEscape(identifier)), []QueryParam{}, body)
	if err != nil {
		return nil, err
	}
	var result UpdateWorkplaceRoleResponse
	if err = json.Unmarshal(response.Body, &result); err != nil {
		return nil, &APIError{Err: err, Message: "Unable to parse workplace role"}
	}
	return &result.Role, nil
}

func (client APIClient) DeleteWorkplaceRole(ctx context.Context, identifier string) error {
	_, err := client.PerformRequestWithRetry(ctx, "DELETE", fmt.Sprintf("/v3/workplace/roles/role/%s", url.PathEscape(identifier)), []QueryParam{}, nil)
	if err != nil {
		return err
	}
	return nil
}
