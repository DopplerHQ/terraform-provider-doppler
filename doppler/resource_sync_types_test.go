package doppler

import "testing"

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
