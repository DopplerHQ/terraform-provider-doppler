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

func getSecondsDuration(seconds int64) *time.Duration {
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
	if err != nil {
		return &APIResponse{HTTPResponse: r, Body: nil}, &APIError{Err: err, Message: "Unable to load response data"}
	}
	response := &APIResponse{HTTPResponse: r, Body: body}

	if !isSuccess(r.StatusCode) {
		if contentType := r.Header.Get("content-type"); strings.HasPrefix(contentType, "application/json") {
			var errResponse ErrorResponse
			err := json.Unmarshal(body, &errResponse)
			if err != nil {
				return response, &APIError{Err: err, Message: "Unable to load response"}
			}

			var retryAfter *time.Duration
			if errResponse.Data["isRetryable"] == true {
				// Retry immediately
				retryAfter = getSecondsDuration(0)
			} else if r.StatusCode == 429 {
				retryAfterStr := r.Header.Get("retry-after")
				retryAfterInt, retryAfterErr := strconv.ParseInt(retryAfterStr, 10, 64)
				if retryAfterErr == nil {
					// Parse successful `retry-after` header result
					retryAfter = getSecondsDuration(retryAfterInt)
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
			}
		}
		return nil, &APIError{Err: fmt.Errorf("%d status code; %d bytes", r.StatusCode, len(body)), Message: "Unable to load response"}
	}
	if err != nil {
		return nil, &APIError{Err: err, Message: "Unable to parse response data"}
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
	result, modelErr := ParseComputedSecrets(response.Body)
	if modelErr != nil {
		return nil, &APIError{Err: modelErr, Message: "Unable to parse secrets"}
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
	jsonErr := json.Unmarshal(response.Body, &result)
	if jsonErr != nil {
		return nil, &APIError{Err: jsonErr, Message: "Unable to parse secret"}
	}
	return &result, nil
}

func (client APIClient) UpdateSecrets(ctx context.Context, project string, config string, secrets []RawSecret) error {
	secretsPayload := map[string]interface{}{}
	for _, secret := range secrets {
		if secret.Value != nil {
			secretsPayload[secret.Name] = *secret.Value
		} else {
			secretsPayload[secret.Name] = nil
		}
	}
	payload := map[string]interface{}{
		"secrets": secretsPayload,
	}
	if project != "" {
		payload["project"] = project
	}
	if config != "" {
		payload["config"] = config
	}
	body, jsonErr := json.Marshal(payload)
	if jsonErr != nil {
		return &APIError{Err: jsonErr, Message: "Unable to parse secrets"}
	}
	_, err := client.PerformRequestWithRetry(ctx, "POST", "/v3/configs/config/secrets", []QueryParam{}, body)
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
	jsonErr := json.Unmarshal(response.Body, &result)
	if jsonErr != nil {
		return nil, &APIError{Err: jsonErr, Message: "Unable to parse project"}
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
	body, jsonErr := json.Marshal(payload)
	if jsonErr != nil {
		return nil, &APIError{Err: jsonErr, Message: "Unable to serialize project"}
	}
	response, err := client.PerformRequestWithRetry(ctx, "POST", "/v3/projects", []QueryParam{}, body)
	if err != nil {
		return nil, err
	}
	var result ProjectResponse
	jsonErr = json.Unmarshal(response.Body, &result)
	if jsonErr != nil {
		return nil, &APIError{Err: jsonErr, Message: "Unable to parse project"}
	}
	return &result.Project, nil
}

func (client APIClient) UpdateProject(ctx context.Context, currentName string, newName string, description string) (*Project, error) {
	payload := map[string]interface{}{
		"project":     currentName,
		"name":        newName,
		"description": description,
	}

	body, jsonErr := json.Marshal(payload)
	if jsonErr != nil {
		return nil, &APIError{Err: jsonErr, Message: "Unable to serialize project"}
	}
	response, err := client.PerformRequestWithRetry(ctx, "POST", "/v3/projects/project", []QueryParam{}, body)
	if err != nil {
		return nil, err
	}
	var result ProjectResponse
	jsonErr = json.Unmarshal(response.Body, &result)
	if jsonErr != nil {
		return nil, &APIError{Err: jsonErr, Message: "Unable to parse project"}
	}
	return &result.Project, nil
}

func (client APIClient) DeleteProject(ctx context.Context, name string) error {
	payload := map[string]interface{}{
		"project": name,
	}
	body, jsonErr := json.Marshal(payload)
	if jsonErr != nil {
		return &APIError{Err: jsonErr, Message: "Unable to serialize project"}
	}
	_, err := client.PerformRequestWithRetry(ctx, "DELETE", "/v3/projects/project", []QueryParam{}, body)
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
	jsonErr := json.Unmarshal(response.Body, &result)
	if jsonErr != nil {
		return nil, &APIError{Err: jsonErr, Message: "Unable to parse environment"}
	}
	return &result.Environment, nil
}

func (client APIClient) CreateEnvironment(ctx context.Context, project string, slug string, name string) (*Environment, error) {
	payload := map[string]interface{}{
		"project": project,
		"name":    name,
		"slug":    slug,
	}
	body, jsonErr := json.Marshal(payload)
	if jsonErr != nil {
		return nil, &APIError{Err: jsonErr, Message: "Unable to serialize environment"}
	}
	response, err := client.PerformRequestWithRetry(ctx, "POST", "/v3/environments", []QueryParam{}, body)
	if err != nil {
		return nil, err
	}
	var result EnvironmentResponse
	jsonErr = json.Unmarshal(response.Body, &result)
	if jsonErr != nil {
		return nil, &APIError{Err: jsonErr, Message: "Unable to parse environment"}
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

	body, jsonErr := json.Marshal(payload)
	if jsonErr != nil {
		return nil, &APIError{Err: jsonErr, Message: "Unable to serialize environment"}
	}
	response, err := client.PerformRequestWithRetry(ctx, "PUT", "/v3/environments/environment", params, body)
	if err != nil {
		return nil, err
	}
	var result EnvironmentResponse
	jsonErr = json.Unmarshal(response.Body, &result)
	if jsonErr != nil {
		return nil, &APIError{Err: jsonErr, Message: "Unable to parse project"}
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
	jsonErr := json.Unmarshal(response.Body, &result)
	if jsonErr != nil {
		return nil, &APIError{Err: jsonErr, Message: "Unable to parse config"}
	}
	return &result.Config, nil
}

func (client APIClient) CreateConfig(ctx context.Context, project string, environment string, name string) (*Config, error) {
	payload := map[string]interface{}{
		"project":     project,
		"environment": environment,
		"name":        name,
	}
	body, jsonErr := json.Marshal(payload)
	if jsonErr != nil {
		return nil, &APIError{Err: jsonErr, Message: "Unable to serialize config"}
	}
	response, err := client.PerformRequestWithRetry(ctx, "POST", "/v3/configs", []QueryParam{}, body)
	if err != nil {
		return nil, err
	}
	var result ConfigResponse
	jsonErr = json.Unmarshal(response.Body, &result)
	if jsonErr != nil {
		return nil, &APIError{Err: jsonErr, Message: "Unable to parse config"}
	}
	return &result.Config, nil
}

func (client APIClient) RenameConfig(ctx context.Context, project string, currentName string, newName string) (*Config, error) {
	payload := map[string]interface{}{
		"project": project,
		"config":  currentName,
		"name":    newName,
	}

	body, jsonErr := json.Marshal(payload)
	if jsonErr != nil {
		return nil, &APIError{Err: jsonErr, Message: "Unable to serialize config"}
	}
	response, err := client.PerformRequestWithRetry(ctx, "POST", "/v3/configs/config", []QueryParam{}, body)
	if err != nil {
		return nil, err
	}
	var result ConfigResponse
	jsonErr = json.Unmarshal(response.Body, &result)
	if jsonErr != nil {
		return nil, &APIError{Err: jsonErr, Message: "Unable to parse config"}
	}
	return &result.Config, nil
}

func (client APIClient) DeleteConfig(ctx context.Context, project string, name string) error {
	payload := map[string]interface{}{
		"project": project,
		"config":  name,
	}

	body, jsonErr := json.Marshal(payload)
	if jsonErr != nil {
		return &APIError{Err: jsonErr, Message: "Unable to serialize config"}
	}
	_, err := client.PerformRequestWithRetry(ctx, "DELETE", "/v3/configs/config", []QueryParam{}, body)
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
	jsonErr := json.Unmarshal(response.Body, &result)
	if jsonErr != nil {
		return nil, &APIError{Err: jsonErr, Message: "Unable to parse service tokens"}
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
	body, jsonErr := json.Marshal(payload)
	if jsonErr != nil {
		return nil, &APIError{Err: jsonErr, Message: "Unable to serialize service token"}
	}
	response, err := client.PerformRequestWithRetry(ctx, "POST", "/v3/configs/config/tokens", []QueryParam{}, body)
	if err != nil {
		return nil, err
	}
	var result ServiceTokenResponse
	jsonErr = json.Unmarshal(response.Body, &result)
	if jsonErr != nil {
		return nil, &APIError{Err: jsonErr, Message: "Unable to parse service token"}
	}
	return &result.ServiceToken, nil
}

func (client APIClient) DeleteServiceToken(ctx context.Context, project string, config string, slug string) error {
	payload := map[string]interface{}{
		"project": project,
		"config":  config,
		"slug":    slug,
	}

	body, jsonErr := json.Marshal(payload)
	if jsonErr != nil {
		return &APIError{Err: jsonErr, Message: "Unable to serialize config"}
	}
	_, err := client.PerformRequestWithRetry(ctx, "DELETE", "/v3/configs/config/tokens/token", []QueryParam{}, body)
	if err != nil {
		return err
	}
	return nil
}
