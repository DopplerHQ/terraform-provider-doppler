package doppler

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func configuredTestClient(t *testing.T, host, token string, verifyTLS bool) APIClient {
	t.Helper()
	d := schema.TestResourceDataRaw(t, Provider().Schema, map[string]interface{}{
		"host": host, "doppler_token": token, "verify_tls": verifyTLS,
		"oidc_identity": "", "oidc_token": "", "oidc_token_file": "",
	})
	m, diags := providerConfigure(context.Background(), d)
	if diags.HasError() {
		t.Fatalf("configure: %v", diags)
	}
	client := m.(APIClient)
	if client.httpClient != nil {
		t.Cleanup(client.httpClient.CloseIdleConnections)
	}
	return client
}

func TestConfiguredClientReusesConnections(t *testing.T) {
	var connections atomic.Int32
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, _, ok := r.BasicAuth()
		if !ok || user != "test-only" {
			t.Error("missing request authentication")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true}`))
	}))
	server.Config.ConnState = func(_ net.Conn, state http.ConnState) {
		if state == http.StateNew {
			connections.Add(1)
		}
	}
	server.Start()
	defer server.Close()
	client := configuredTestClient(t, server.URL, "test-only", true)
	for i := 0; i < 5; i++ {
		copy := client // Resource callbacks receive the provider client by value.
		if _, err := copy.PerformRequestWithRetry(context.Background(), "GET", "/test", nil, nil); err != nil {
			t.Fatal(err)
		}
	}
	if got := connections.Load(); got != 1 {
		t.Fatalf("opened %d connections for 5 sequential requests; want 1", got)
	}
}

func TestConfiguredClientConcurrentRequests(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		_, _ = w.Write([]byte(`{"success":true}`))
	}))
	defer server.Close()
	client := configuredTestClient(t, server.URL, "test-only", true)
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := client.PerformRequestWithRetry(context.Background(), "GET", "/test", nil, nil); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	if requests.Load() != 20 {
		t.Fatal("lost requests")
	}
}

func TestConfiguredClientPreservesTLSIsolation(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"success":true}`))
	}))
	defer server.Close()
	insecure := configuredTestClient(t, server.URL, "test-only", false)
	secure := configuredTestClient(t, server.URL, "test-only", true)
	if _, err := insecure.PerformRequestWithRetry(context.Background(), "GET", "/test", nil, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := secure.PerformRequestWithRetry(context.Background(), "GET", "/test", nil, nil); err == nil {
		t.Fatal("verified client accepted untrusted certificate")
	}
	for _, client := range []APIClient{secure, insecure} {
		if client.httpClient.Transport.(*http.Transport).TLSClientConfig.MinVersion != tls.VersionTLS12 {
			t.Fatal("TLS minimum changed")
		}
	}
}
