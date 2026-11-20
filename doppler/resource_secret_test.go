package doppler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// fakeSecretsAPI serves the two Doppler endpoints that doppler_secret uses and
// records the last value written to each secret.
type fakeSecretsAPI struct {
	mu      sync.Mutex
	secrets map[string]*fakeSecret
	writes  int
}

type fakeSecret struct {
	value      string
	visibility string
	valueType  string
}

func newFakeSecretsAPI(t *testing.T) (*fakeSecretsAPI, *httptest.Server) {
	api := &fakeSecretsAPI{secrets: map[string]*fakeSecret{}}
	server := httptest.NewServer(http.HandlerFunc(api.serve))
	t.Cleanup(server.Close)
	return api, server
}

func (api *fakeSecretsAPI) serve(w http.ResponseWriter, r *http.Request) {
	api.mu.Lock()
	defer api.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")

	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/v3/configs/config/secret":
		q := r.URL.Query()
		secret, ok := api.secrets[getSecretId(q.Get("project"), q.Get("config"), q.Get("name"))]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"messages":["Could not find requested secret"],"success":false}`))
			return
		}
		_ = json.NewEncoder(w).Encode(Secret{
			Name: q.Get("name"),
			Value: SecretValue{
				Raw:           &secret.value,
				Computed:      &secret.value,
				RawVisibility: &secret.visibility,
				RawValueType:  &ValueType{Type: secret.valueType},
			},
		})
	case r.Method == http.MethodPost && r.URL.Path == "/v3/configs/config/secrets":
		var body struct {
			Project        string          `json:"project"`
			Config         string          `json:"config"`
			ChangeRequests []ChangeRequest `json:"change_requests"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		for _, change := range body.ChangeRequests {
			id := getSecretId(body.Project, body.Config, change.Name)
			if change.OriginalName != nil && *change.OriginalName != change.Name {
				delete(api.secrets, getSecretId(body.Project, body.Config, *change.OriginalName))
			}
			if change.ShouldDelete {
				delete(api.secrets, id)
				continue
			}
			api.writes++
			api.secrets[id] = &fakeSecret{value: *change.Value, visibility: change.Visibility, valueType: change.ValueType.Type}
		}
		_, _ = w.Write([]byte(`{"success":true}`))
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

// check verifies that Doppler holds exactly one secret, with the given name,
// value and visibility, after the given number of writes.
func (api *fakeSecretsAPI) check(name, wantValue, wantVisibility string, wantWrites int) resource.TestCheckFunc {
	return func(*terraform.State) error {
		api.mu.Lock()
		defer api.mu.Unlock()
		secret, ok := api.secrets[getSecretId("backend", "dev", name)]
		if !ok || len(api.secrets) != 1 {
			return fmt.Errorf("Doppler should hold only %s, but holds %d secrets", name, len(api.secrets))
		}
		if secret.value != wantValue || secret.visibility != wantVisibility || api.writes != wantWrites {
			return fmt.Errorf("Doppler has %q (%s) after %d writes, want %q (%s) after %d writes",
				secret.value, secret.visibility, api.writes, wantValue, wantVisibility, wantWrites)
		}
		return nil
	}
}

// checkNotInState fails if any attribute of any resource in the state holds value.
func checkNotInState(value string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, module := range state.Modules {
			for name, res := range module.Resources {
				for attr, v := range res.Primary.Attributes {
					if strings.Contains(v, value) {
						return fmt.Errorf("%s.%s holds the secret value in state", name, attr)
					}
				}
			}
		}
		return nil
	}
}

// skipBeforeTerraform111 skips the test when the Terraform CLI that the test
// framework will use is older than 1.11, which has no write-only arguments.
// When no CLI is installed, the framework downloads the latest release.
func skipBeforeTerraform111(t *testing.T) {
	path := os.Getenv("TF_ACC_TERRAFORM_PATH")
	if path == "" {
		var err error
		if path, err = exec.LookPath("terraform"); err != nil {
			return
		}
	}
	out, err := exec.Command(path, "version", "-json").Output()
	if err != nil {
		t.Fatalf("reading the Terraform version: %v", err)
	}
	var v struct {
		Version string `json:"terraform_version"`
	}
	if err := json.Unmarshal(out, &v); err != nil {
		t.Fatalf("parsing the Terraform version: %v", err)
	}
	if version.Must(version.NewVersion(v.Version)).LessThan(version.Must(version.NewVersion("1.11.0"))) {
		t.Skipf("write-only arguments need Terraform 1.11 or later, found %s", v.Version)
	}
}

func secretConfig(host, name, args string) string {
	return fmt.Sprintf(`
provider "doppler" {
  host          = %q
  doppler_token = "dp.st.test"
}

resource "doppler_secret" "db_password" {
  project = "backend"
  config  = "dev"
  name    = %q
%s
}
`, host, name, args)
}

// TestAccResourceSecret_writeOnly runs the Terraform CLI against a local fake of
// the Doppler API, so it needs Terraform 1.11 or later but no Doppler token.
func TestAccResourceSecret_writeOnly(t *testing.T) {
	api, server := newFakeSecretsAPI(t)
	config := func(name, args string) string { return secretConfig(server.URL, name, args) }
	writeOnly := func(value string, version int) string {
		return fmt.Sprintf("  value_wo = %q\n  value_wo_version = %d", value, version)
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { skipBeforeTerraform111(t) },
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"doppler": func() (*schema.Provider, error) { return Provider(), nil },
		},
		Steps: []resource.TestStep{
			{
				Config:      config("DB_PASSWORD", `  value = "s1"`+"\n"+writeOnly("s1", 1)),
				ExpectError: regexp.MustCompile(`only one of`),
			},
			{
				Config:      config("DB_PASSWORD", `  value_wo = "s1"`),
				ExpectError: regexp.MustCompile(`all of .value_wo,value_wo_version. must be specified`),
			},
			{
				Config: config("DB_PASSWORD", writeOnly("wo-secret-1", 1)),
				Check: resource.ComposeTestCheckFunc(
					api.check("DB_PASSWORD", "wo-secret-1", "masked", 1),
					checkNotInState("wo-secret-1"),
					resource.TestCheckResourceAttr("doppler_secret.db_password", "value_wo_version", "1"),
				),
			},
			{
				// Import reads the value into state; the next apply removes it.
				ResourceName:            "doppler_secret.db_password",
				ImportState:             true,
				ImportStateId:           "backend.dev.DB_PASSWORD",
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"value", "computed", "value_wo_version"},
			},
			{
				// A new value_wo without a version change is not written.
				Config: config("DB_PASSWORD", writeOnly("wo-secret-2", 1)),
				Check:  api.check("DB_PASSWORD", "wo-secret-1", "masked", 1),
			},
			{
				Config: config("DB_PASSWORD", writeOnly("wo-secret-2", 2)),
				Check: resource.ComposeTestCheckFunc(
					api.check("DB_PASSWORD", "wo-secret-2", "masked", 2),
					checkNotInState("wo-secret-2"),
				),
			},
			{
				// Other updates write the current value_wo.
				Config: config("DB_PASS", writeOnly("wo-secret-2", 2)+"\n  visibility = \"unmasked\""),
				Check: resource.ComposeTestCheckFunc(
					api.check("DB_PASS", "wo-secret-2", "unmasked", 3),
					checkNotInState("wo-secret-2"),
				),
			},
			{
				Config: config("DB_PASS", `  value = "plain-secret"`),
				Check: resource.ComposeTestCheckFunc(
					api.check("DB_PASS", "plain-secret", "masked", 4),
					resource.TestCheckResourceAttr("doppler_secret.db_password", "value", "plain-secret"),
					resource.TestCheckResourceAttr("doppler_secret.db_password", "computed", "plain-secret"),
				),
			},
			{
				Config: config("DB_PASS", writeOnly("wo-secret-3", 3)),
				Check: resource.ComposeTestCheckFunc(
					api.check("DB_PASS", "wo-secret-3", "masked", 5),
					checkNotInState("wo-secret-3"),
					checkNotInState("plain-secret"),
				),
			},
		},
	})
}
