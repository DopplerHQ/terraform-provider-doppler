package doppler

import (
	"bytes"
	"context"
	"crypto/sha256"
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

func (client APIClient) GetId() string {
	digester := sha256.New()
	fmt.Fprint(digester, client.Host)
	fmt.Fprint(digester, client.APIKey)
	fmt.Fprint(digester, client.VerifyTLS)
	return fmt.Sprintf("%x", digester.Sum(nil))
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
	defer r.Body.Close()

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

func (client APIClient) CreateEnvironment(ctx context.Context, project string, slug string, name string) (*Environment, error) {
	payload := map[string]interface{}{
		"project": project,
		"name":    name,
		"slug":    slug,
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

func (client APIClient) RenameEnvironment(ctx context.Context, project string, currentSlug string, newSlug string, newName string) (*Environment, error) {
	params := []QueryParam{
		{Key: "project", Value: project},
		{Key: "environment", Value: currentSlug},
	}
	payload := map[string]interface{}{
		"slug": newSlug,
		"name": newName,
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

// Trusted IPs

func (client APIClient) GetTrustedIPs(ctx context.Context, project, config string) ([]TrustedIP, error) {
	params := []QueryParam{
		{Key: "project", Value: project},
		{Key: "config", Value: config},
	}
	response, err := client.PerformRequestWithRetry(ctx, "GET", "/v3/configs/config/trusted_ips", params, nil)
	if err != nil {
		return nil, err
	}
	var result TrustedIPsListResponse
	if err = json.Unmarshal(response.Body, &result); err != nil {
		return nil, &APIError{Err: err, Message: "Unable to parse trusted IPs"}
	}
	ips := []TrustedIP{}
	for _, ipID := range result.IPs {
		if ipID == "" {
			return nil, &APIError{Err: err, Message: "Unable to parse trusted IP"}
		}
		proj, config, ip, err := parseTrustedIPResourceId(ipID)
		if err != nil {
			return nil, &APIError{Err: err, Message: "Unable to parse trusted IP"}
		}
		ips = append(ips, TrustedIP{
			Project: proj,
			Config:  config,
			IP:      ip,
		})
	}
	return ips, nil
}

func (client APIClient) AddTrustedIP(ctx context.Context, project, config, ip string) (*TrustedIP, error) {
	params := []QueryParam{
		{Key: "project", Value: project},
		{Key: "config", Value: config},
	}
	payload := map[string]interface{}{
		"ip": ip,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, &APIError{Err: err, Message: "Unable to serialize trusted IP"}
	}
	response, err := client.PerformRequestWithRetry(ctx, "POST", "/v3/configs/config/trusted_ips", params, body)
	if err != nil {
		return nil, err
	}
	var result TrustedIPsAddResponse
	if err = json.Unmarshal(response.Body, &result); err != nil {
		return nil, &APIError{Err: err, Message: "Unable to parse trusted IP"}
	}
	return &result.IP, nil
}

func (client APIClient) DeleteTrustedIP(ctx context.Context, project, config, ip string) error {
	params := []QueryParam{
		{Key: "project", Value: project},
		{Key: "config", Value: config},
	}
	payload := map[string]interface{}{
		"ip": ip,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return &APIError{Err: err, Message: "Unable to serialize trusted IP"}
	}
	_, err = client.PerformRequestWithRetry(ctx, "DELETE", "/v3/configs/config/trusted_ips", params, body)
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
