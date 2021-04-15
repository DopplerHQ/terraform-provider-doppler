package doppler

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

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

func (e *APIError) Error() string {
	return e.Message
}

func isSuccess(statusCode int) bool {
	return (statusCode >= 200 && statusCode <= 299) || (statusCode >= 300 && statusCode <= 399)
}

func GetRequest(context APIContext, path string) (*APIResponse, *APIError) {
	url := fmt.Sprintf("%s%s", context.Host, path)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, &APIError{Err: err, Message: "Unable to form request"}
	}

	return PerformRequest(context, req)
}

func PerformRequest(context APIContext, req *http.Request) (*APIResponse, *APIError) {
	client := &http.Client{Timeout: 10 * time.Second}

	req.Header.Set("user-agent", "terraform-provider-doppler")
	req.SetBasicAuth(context.APIKey, "")
	if req.Header.Get("accept") == "" {
		req.Header.Set("accept", "application/json")
	}

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	if !context.VerifyTLS {
		tlsConfig.InsecureSkipVerify = true
	}

	client.Transport = &http.Transport{
		DisableKeepAlives: true,
		TLSClientConfig:   tlsConfig,
	}

	r, err := client.Do(req)
	if err != nil {
		return nil, &APIError{Err: err, Message: "Unable to load response"}
	}
	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)
	response := &APIResponse{HTTPResponse: r, Body: body}

	if !isSuccess(r.StatusCode) {
		if contentType := r.Header.Get("content-type"); strings.HasPrefix(contentType, "application/json") {
			var errResponse ErrorResponse
			err := json.Unmarshal(body, &errResponse)
			if err != nil {
				return response, &APIError{Err: nil, Message: "Unable to load response"}
			}
			return response, &APIError{Err: nil, Message: strings.Join(errResponse.Messages, "\n")}
		}
		return nil, &APIError{Err: nil, Message: "Unable to load response"}
	}
	if err != nil {
		return nil, &APIError{Err: err, Message: "Unable to parse response data"}
	}
	return response, nil
}

func GetSecrets(context APIContext) ([]Secret, *APIError) {
	response, err := GetRequest(context, "/v3/configs/config/secrets/download")
	if err != nil {
		return nil, err
	}
	result, modelErr := ParseSecrets(response.Body)
	if modelErr != nil {
		return nil, &APIError{Err: modelErr, Message: "Unable to parse secrets"}
	}
	return result, nil
}
