package doppler

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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
	Err     error
	Message string
}

type ErrorResponse struct {
	Messages []string
	Success  bool
}

type QueryParam struct {
	Key   string
	Value string
}

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

func (client APIClient) GetRequest(ctx context.Context, path string, params []QueryParam) (*APIResponse, *APIError) {
	url := fmt.Sprintf("%s%s", client.Host, path)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, &APIError{Err: err, Message: "Unable to form request"}
	}

	return client.PerformRequest(req, params)
}

func (client APIClient) PostRequest(ctx context.Context, path string, params []QueryParam, body []byte) (*APIResponse, *APIError) {
	url := fmt.Sprintf("%s%s", client.Host, path)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, &APIError{Err: err, Message: "Unable to form request"}
	}

	return client.PerformRequest(req, params)
}

func (client APIClient) PutRequest(ctx context.Context, path string, params []QueryParam, body []byte) (*APIResponse, *APIError) {
	url := fmt.Sprintf("%s%s", client.Host, path)
	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewReader(body))
	if err != nil {
		return nil, &APIError{Err: err, Message: "Unable to form request"}
	}

	return client.PerformRequest(req, params)
}

func (client APIClient) DeleteRequest(ctx context.Context, path string, params []QueryParam) (*APIResponse, *APIError) {
	url := fmt.Sprintf("%s%s", client.Host, path)
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return nil, &APIError{Err: err, Message: "Unable to form request"}
	}

	return client.PerformRequest(req, params)
}

func (client APIClient) PerformRequest(req *http.Request, params []QueryParam) (*APIResponse, *APIError) {
	httpClient := &http.Client{Timeout: 10 * time.Second}

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
		return nil, &APIError{Err: err, Message: "Unable to load response"}
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
			return response, &APIError{Err: nil, Message: strings.Join(errResponse.Messages, "\n")}
		}
		return nil, &APIError{Err: fmt.Errorf("%d status code; %d bytes", r.StatusCode, len(body)), Message: "Unable to load response"}
	}
	if err != nil {
		return nil, &APIError{Err: err, Message: "Unable to parse response data"}
	}
	return response, nil
}

func (client APIClient) GetSecrets(ctx context.Context) ([]Secret, *APIError) {
	response, err := client.GetRequest(ctx, "/v3/configs/config/secrets/download", []QueryParam{})
	if err != nil {
		return nil, err
	}
	result, modelErr := ParseSecrets(response.Body)
	if modelErr != nil {
		return nil, &APIError{Err: modelErr, Message: "Unable to parse secrets"}
	}
	return result, nil
}
