package doppler

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func stubTransientRetryDelay(t *testing.T) *atomic.Int32 {
	original := getTransientRetryDelay
	var calls atomic.Int32
	getTransientRetryDelay = func() *time.Duration {
		calls.Add(1)
		return getSecondsDuration(0)
	}
	t.Cleanup(func() { getTransientRetryDelay = original })
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
	tests := []struct {
		name        string
		method      string
		status      int
		contentType string
		wantCalls   int32
		wantErr     bool
	}{
		{"GET 502 is retried", http.MethodGet, http.StatusBadGateway, "text/html", 2, false},
		{"GET 503 is retried", http.MethodGet, http.StatusServiceUnavailable, "text/html", 2, false},
		{"GET 504 is retried", http.MethodGet, http.StatusGatewayTimeout, "text/html", 2, false},
		{"HEAD 502 is retried", http.MethodHead, http.StatusBadGateway, "text/html", 2, false},
		{"DELETE 502 is retried", http.MethodDelete, http.StatusBadGateway, "text/html", 2, false},
		{"POST 502 is not retried", http.MethodPost, http.StatusBadGateway, "text/html", 1, true},
		{"PUT 502 is not retried", http.MethodPut, http.StatusBadGateway, "text/html", 1, true},
		{"PATCH 502 is not retried", http.MethodPatch, http.StatusBadGateway, "text/html", 1, true},
		{"GET 500 is not retried", http.MethodGet, http.StatusInternalServerError, "text/html", 1, true},
		{"GET JSON 503 without retry data is not retried", http.MethodGet, http.StatusServiceUnavailable, "application/json", 1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stubTransientRetryDelay(t)
			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if calls.Add(1) == 1 {
					w.Header().Set("Content-Type", tt.contentType)
					w.WriteHeader(tt.status)
					if tt.contentType == "application/json" {
						_, _ = w.Write([]byte(`{"messages":["Service unavailable"],"success":false}`))
					} else {
						_, _ = w.Write([]byte("<html>error</html>"))
					}
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
			stubTransientRetryDelay(t)
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
			stubTransientRetryDelay(t)
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
			retries := stubTransientRetryDelay(t)
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

func TestPerformRequestWithRetry_ConnectionRefused(t *testing.T) {
	retries := stubTransientRetryDelay(t)
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
	if got := retries.Load(); got != MAX_RETRIES {
		t.Fatalf("retries = %d, want %d", got, MAX_RETRIES)
	}
}

func TestPerformRequestWithRetry_CanceledContextIsNotRetried(t *testing.T) {
	retries := stubTransientRetryDelay(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeSuccess(w)
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := APIClient{Host: server.URL, APIKey: "test", VerifyTLS: true}
	_, err := client.PerformRequestWithRetry(ctx, http.MethodGet, "/v3/test", nil, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if got := retries.Load(); got != 0 {
		t.Fatalf("retries = %d, want 0", got)
	}
}

func TestPerformRequestWithRetry_GivesUpAfterMaxRetries(t *testing.T) {
	stubTransientRetryDelay(t)
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
}

func TestGetTransientRetryDelay(t *testing.T) {
	for range 1000 {
		delay := *getTransientRetryDelay()
		if delay < 500*time.Millisecond || delay >= 1500*time.Millisecond {
			t.Fatalf("delay = %v, want within [500ms, 1500ms)", delay)
		}
	}
}
