package doppler

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

func stubTransientRetryDelay(t *testing.T, delay time.Duration) *atomic.Int32 {
	original := transientRetryDelay
	var calls atomic.Int32
	transientRetryDelay = func(int) time.Duration {
		calls.Add(1)
		return delay
	}
	t.Cleanup(func() { transientRetryDelay = original })
	return &calls
}

func writeSuccess(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{}`))
}

func dropConnection(t *testing.T, w http.ResponseWriter, raw string) {
	conn, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		t.Error(err)
		return
	}
	if raw != "" {
		_, _ = conn.Write([]byte(raw))
	}
	_ = conn.Close()
}

type dropFirstConnListener struct {
	net.Listener
	dropped atomic.Bool
}

func (l *dropFirstConnListener) Accept() (net.Conn, error) {
	for {
		conn, err := l.Listener.Accept()
		if err != nil {
			return nil, err
		}
		if l.dropped.CompareAndSwap(false, true) {
			_ = conn.Close()
			continue
		}
		return conn, nil
	}
}

func TestPerformRequestWithRetry_TransientStatusCodes(t *testing.T) {
	html := "<html>error</html>"
	jsonError := `{"messages":["Service unavailable"],"success":false}`
	tests := []struct {
		name        string
		method      string
		status      int
		contentType string
		body        string
		wantCalls   int32
		wantErr     bool
	}{
		{"GET 502 is retried", http.MethodGet, http.StatusBadGateway, "text/html", html, 2, false},
		{"GET 503 is retried", http.MethodGet, http.StatusServiceUnavailable, "text/html", html, 2, false},
		{"GET 504 is retried", http.MethodGet, http.StatusGatewayTimeout, "text/html", html, 2, false},
		{"GET 520 is retried", http.MethodGet, 520, "text/html", html, 2, false},
		{"GET 524 is retried", http.MethodGet, 524, "text/html", html, 2, false},
		{"HEAD 502 is retried", http.MethodHead, http.StatusBadGateway, "text/html", html, 2, false},
		{"DELETE 502 is retried", http.MethodDelete, http.StatusBadGateway, "text/html", html, 2, false},
		{"POST 502 is not retried", http.MethodPost, http.StatusBadGateway, "text/html", html, 1, true},
		{"PUT 502 is not retried", http.MethodPut, http.StatusBadGateway, "text/html", html, 1, true},
		{"PATCH 502 is not retried", http.MethodPatch, http.StatusBadGateway, "text/html", html, 1, true},
		{"GET 500 is not retried", http.MethodGet, http.StatusInternalServerError, "text/html", html, 1, true},
		{"GET 525 is not retried", http.MethodGet, 525, "text/html", html, 1, true},
		{"GET JSON 503 is retried", http.MethodGet, http.StatusServiceUnavailable, "application/json", jsonError, 2, false},
		{"GET 502 labelled JSON with an HTML body is retried", http.MethodGet, http.StatusBadGateway, "application/json", html, 2, false},
		{"POST JSON 503 is not retried", http.MethodPost, http.StatusServiceUnavailable, "application/json", jsonError, 1, true},
		{"GET JSON 400 is not retried", http.MethodGet, http.StatusBadRequest, "application/json", `{"messages":["Bad request"],"success":false}`, 1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stubTransientRetryDelay(t, 0)
			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if calls.Add(1) == 1 {
					w.Header().Set("Content-Type", tt.contentType)
					w.WriteHeader(tt.status)
					_, _ = w.Write([]byte(tt.body))
					return
				}
				writeSuccess(w)
			}))
			defer server.Close()

			client := APIClient{Host: server.URL, APIKey: "test", VerifyTLS: true}
			_, err := client.PerformRequestWithRetry(context.Background(), tt.method, "/v3/test", nil, nil)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if got := calls.Load(); got != tt.wantCalls {
				t.Fatalf("calls = %d, want %d", got, tt.wantCalls)
			}
		})
	}
}

func TestPerformRequestWithRetry_ConnectionDroppedAfterRequestWritten(t *testing.T) {
	tests := []struct {
		name      string
		method    string
		wantCalls int32
		wantErr   bool
	}{
		{"GET is retried", http.MethodGet, 2, false},
		{"DELETE is retried", http.MethodDelete, 2, false},
		{"POST is not retried", http.MethodPost, 1, true},
		{"PATCH is not retried", http.MethodPatch, 1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stubTransientRetryDelay(t, 0)
			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if calls.Add(1) == 1 {
					dropConnection(t, w, "")
					return
				}
				writeSuccess(w)
			}))
			defer server.Close()

			client := APIClient{Host: server.URL, APIKey: "test", VerifyTLS: true}
			_, err := client.PerformRequestWithRetry(context.Background(), tt.method, "/v3/test", nil, []byte(`{}`))
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if got := calls.Load(); got != tt.wantCalls {
				t.Fatalf("calls = %d, want %d", got, tt.wantCalls)
			}
		})
	}
}

func TestPerformRequestWithRetry_TruncatedResponseBody(t *testing.T) {
	tests := []struct {
		name      string
		method    string
		wantCalls int32
		wantErr   bool
	}{
		{"GET is retried", http.MethodGet, 2, false},
		{"POST is not retried", http.MethodPost, 1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stubTransientRetryDelay(t, 0)
			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if calls.Add(1) == 1 {
					dropConnection(t, w, "HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nContent-Length: 100\r\n\r\n{}")
					return
				}
				writeSuccess(w)
			}))
			defer server.Close()

			client := APIClient{Host: server.URL, APIKey: "test", VerifyTLS: true}
			_, err := client.PerformRequestWithRetry(context.Background(), tt.method, "/v3/test", nil, nil)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if got := calls.Load(); got != tt.wantCalls {
				t.Fatalf("calls = %d, want %d", got, tt.wantCalls)
			}
		})
	}
}

func TestPerformRequestWithRetry_ConnectionDroppedDuringTLSHandshake(t *testing.T) {
	for _, method := range []string{http.MethodGet, http.MethodPost} {
		t.Run(method, func(t *testing.T) {
			retries := stubTransientRetryDelay(t, 0)
			var calls atomic.Int32
			server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				writeSuccess(w)
			}))
			server.Listener = &dropFirstConnListener{Listener: server.Listener}
			server.StartTLS()
			defer server.Close()

			client := APIClient{Host: server.URL, APIKey: "test", VerifyTLS: false}
			_, err := client.PerformRequestWithRetry(context.Background(), method, "/v3/test", nil, []byte(`{}`))
			if err != nil {
				t.Fatalf("err = %v", err)
			}
			if got := retries.Load(); got != 1 {
				t.Fatalf("retries = %d, want 1", got)
			}
			if got := calls.Load(); got != 1 {
				t.Fatalf("calls = %d, want 1", got)
			}
		})
	}
}

func TestPerformRequestWithRetry_TLSVersionRejectedIsNotRetried(t *testing.T) {
	retries := stubTransientRetryDelay(t, 0)
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeSuccess(w)
	}))
	server.TLS = &tls.Config{MaxVersion: tls.VersionTLS11}
	server.StartTLS()
	defer server.Close()

	client := APIClient{Host: server.URL, APIKey: "test", VerifyTLS: false}
	_, err := client.PerformRequestWithRetry(context.Background(), http.MethodGet, "/v3/test", nil, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if got := retries.Load(); got != 0 {
		t.Fatalf("retries = %d, want 0 (err = %v)", got, err)
	}
}

type timeoutError struct{}

func (timeoutError) Error() string   { return "i/o timeout" }
func (timeoutError) Timeout() bool   { return true }
func (timeoutError) Temporary() bool { return true }

func TestIsTransientNetworkError(t *testing.T) {
	wrap := func(err error) error {
		return &url.Error{Op: "Get", URL: "https://api.doppler.com", Err: err}
	}
	sysErr := func(errno syscall.Errno) error {
		return wrap(&net.OpError{Op: "read", Net: "tcp", Err: os.NewSyscallError("read", errno)})
	}
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"timeout", wrap(timeoutError{}), true},
		{"unknown host", wrap(&net.OpError{Op: "dial", Err: &net.DNSError{Err: "no such host", IsNotFound: true}}), false},
		{"temporary DNS failure", wrap(&net.OpError{Op: "dial", Err: &net.DNSError{Err: "server misbehaving", IsTemporary: true}}), true},
		{"TLS alert", wrap(&net.OpError{Op: "remote error", Err: tls.AlertError(70)}), false},
		{"connection refused", sysErr(syscall.ECONNREFUSED), true},
		{"connection reset", sysErr(syscall.ECONNRESET), true},
		{"connection aborted", sysErr(syscall.ECONNABORTED), true},
		{"broken pipe", sysErr(syscall.EPIPE), true},
		{"EOF", wrap(io.EOF), true},
		{"unexpected EOF", io.ErrUnexpectedEOF, true},
		{"nil", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isTransientNetworkError(tt.err); got != tt.want {
				t.Fatalf("isTransientNetworkError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRetryDelay(t *testing.T) {
	stubTransientRetryDelay(t, 0)
	timeout := &url.Error{Op: "Get", URL: "https://api.doppler.com", Err: timeoutError{}}
	status := func(code int) *APIResponse {
		return &APIResponse{HTTPResponse: &http.Response{StatusCode: code}}
	}
	canceled, cancel := context.WithCancel(context.Background())
	cancel()

	tests := []struct {
		name      string
		ctx       context.Context
		method    string
		apiError  APIError
		wantRetry bool
	}{
		{"timeout before the request was written, POST", context.Background(), http.MethodPost, APIError{Err: timeout}, true},
		{"timeout after the request was written, POST", context.Background(), http.MethodPost, APIError{Err: timeout, requestWritten: true}, false},
		{"timeout after the request was written, GET", context.Background(), http.MethodGet, APIError{Err: timeout, requestWritten: true}, true},
		{"timeout with a canceled context", canceled, http.MethodGet, APIError{Err: timeout}, false},
		{"API asked for a retry, POST", context.Background(), http.MethodPost, APIError{RetryAfter: getSecondsDuration(0), Response: status(http.StatusConflict), requestWritten: true}, true},
		{"502, GET", context.Background(), http.MethodGet, APIError{Response: status(http.StatusBadGateway), requestWritten: true}, true},
		{"502, POST", context.Background(), http.MethodPost, APIError{Response: status(http.StatusBadGateway), requestWritten: true}, false},
		{"500, GET", context.Background(), http.MethodGet, APIError{Response: status(http.StatusInternalServerError), requestWritten: true}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, retry := retryDelay(tt.ctx, tt.method, &tt.apiError, 0); retry != tt.wantRetry {
				t.Fatalf("retry = %v, want %v", retry, tt.wantRetry)
			}
		})
	}
}

func TestPerformRequestWithRetry_ConnectionRefused(t *testing.T) {
	retries := stubTransientRetryDelay(t, 0)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	host := "http://" + listener.Addr().String()
	_ = listener.Close()

	client := APIClient{Host: host, APIKey: "test", VerifyTLS: true}
	_, err = client.PerformRequestWithRetry(context.Background(), http.MethodPost, "/v3/test", nil, []byte(`{}`))
	if err == nil {
		t.Fatal("expected error")
	}
	// No delay after the final attempt.
	if got := retries.Load(); got != MAX_RETRIES-1 {
		t.Fatalf("retries = %d, want %d", got, MAX_RETRIES-1)
	}
}

func TestPerformRequestWithRetry_ExpiredDeadlineIsNotRetried(t *testing.T) {
	retries := stubTransientRetryDelay(t, 0)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(5 * time.Second):
		}
		writeSuccess(w)
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	client := APIClient{Host: server.URL, APIKey: "test", VerifyTLS: true}
	_, err := client.PerformRequestWithRetry(ctx, http.MethodGet, "/v3/test", nil, nil)
	var apiError *APIError
	if !errors.As(err, &apiError) || !errors.Is(apiError.Err, context.DeadlineExceeded) {
		t.Fatalf("err = %v, want a deadline error", err)
	}
	if got := retries.Load(); got != 0 {
		t.Fatalf("retries = %d, want 0", got)
	}
}

func TestPerformRequestWithRetry_CancelDuringRetryDelayReturnsPromptly(t *testing.T) {
	stubTransientRetryDelay(t, time.Hour)
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(100*time.Millisecond, cancel)

	client := APIClient{Host: server.URL, APIKey: "test", VerifyTLS: true}
	start := time.Now()
	_, err := client.PerformRequestWithRetry(ctx, http.MethodGet, "/v3/test", nil, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("returned after %v, want prompt return on cancel", elapsed)
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("calls = %d, want 1", got)
	}
}

func TestPerformRequestWithRetry_GivesUpAfterMaxRetries(t *testing.T) {
	retries := stubTransientRetryDelay(t, 0)
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()

	client := APIClient{Host: server.URL, APIKey: "test", VerifyTLS: true}
	_, err := client.PerformRequestWithRetry(context.Background(), http.MethodGet, "/v3/test", nil, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if got := calls.Load(); got != MAX_RETRIES {
		t.Fatalf("calls = %d, want %d", got, MAX_RETRIES)
	}
	if got := retries.Load(); got != MAX_RETRIES-1 {
		t.Fatalf("retries = %d, want %d", got, MAX_RETRIES-1)
	}
}

func TestTransientRetryDelay(t *testing.T) {
	ranges := []struct{ min, max time.Duration }{
		{500 * time.Millisecond, time.Second},
		{time.Second, 2 * time.Second},
		{2 * time.Second, 4 * time.Second},
	}
	for attempt := range MAX_RETRIES {
		want := ranges[min(attempt, len(ranges)-1)]
		for range 200 {
			if delay := transientRetryDelay(attempt); delay < want.min || delay >= want.max {
				t.Fatalf("attempt %d: delay = %v, want within [%v, %v)", attempt, delay, want.min, want.max)
			}
		}
	}
}
