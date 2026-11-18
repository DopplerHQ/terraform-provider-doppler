package doppler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Contract: https://docs.doppler.com/reference/syncs-create (Vercel variable_type).
func TestVercelSyncVariableTypes(t *testing.T) {
	validator := resourceSyncVercel().Schema["variable_type"].ValidateFunc
	for _, value := range []string{"config", "secret", "dynamic", "encrypted", "sensitive"} {
		t.Run(value, func(t *testing.T) {
			if _, errs := validator(value, "variable_type"); len(errs) != 0 {
				t.Fatalf("documented variable type %q rejected: %v", value, errs)
			}
		})
	}
	for _, value := range []string{"plain", "invalid"} {
		t.Run(value, func(t *testing.T) {
			if _, errs := validator(value, "variable_type"); len(errs) == 0 {
				t.Fatalf("unsupported variable type %q accepted", value)
			}
		})
	}
}

func TestVercelSyncCreatePayload(t *testing.T) {
	for _, tc := range []struct{ name, team, kind, wantTeam, wantKind string }{
		{"omitted", "", "", "personal", "encrypted"},
		{"team", "team_sample", "", "team_sample", "encrypted"},
		{"personal", "", "dynamic", "personal", "dynamic"},
		{"legacy", "team_sample", "sensitive", "team_sample", "sensitive"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var body map[string]interface{}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost || r.URL.Path != "/v3/configs/config/syncs" {
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
				}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Error(err)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"sync":{"slug":"sample_sync"}}`))
			}))
			defer server.Close()
			resource := resourceSyncVercel()
			config := map[string]interface{}{"integration": "sample_integration", "project": "sample_project", "config": "sample_config", "project_id": "sample_target", "target_id": "preview"}
			if tc.team != "" {
				config["team_id"] = tc.team
			}
			if tc.kind != "" {
				config["variable_type"] = tc.kind
			}
			data := schema.TestResourceDataRaw(t, resource.Schema, config)
			if diags := resource.CreateContext(context.Background(), data, APIClient{Host: server.URL, APIKey: "synthetic-test-only"}); diags.HasError() {
				t.Fatal(diags)
			}
			payload, ok := body["data"].(map[string]interface{})
			if !ok {
				t.Fatalf("missing sync data: %v", body)
			}
			if payload["team_id"] != tc.wantTeam || payload["variable_type"] != tc.wantKind {
				t.Fatalf("unexpected Vercel defaults: %v, want team_id=%q variable_type=%q", payload, tc.wantTeam, tc.wantKind)
			}
		})
	}
}
