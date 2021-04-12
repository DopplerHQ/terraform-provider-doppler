package doppler

import (
	"encoding/json"
	"sort"
)

type APIContext struct {
	Host      string
	APIKey    string
	VerifyTLS bool
}

type ComputedSecret struct {
	Name          string `json:"name"`
	RawValue      string `json:"raw"`
	ComputedValue string `json:"computed"`
}

func ParseSecrets(response []byte) ([]ComputedSecret, error) {
	var result map[string]interface{}
	err := json.Unmarshal(response, &result)
	if err != nil {
		return nil, err
	}

	computed := make([]ComputedSecret, 0)
	secrets := result["secrets"].(map[string]interface{})
	for key, secret := range secrets {
		computedSecret := ComputedSecret{Name: key}
		val := secret.(map[string]interface{})
		if val["raw"] != nil {
			computedSecret.RawValue = val["raw"].(string)
		}
		if val["computed"] != nil {
			computedSecret.ComputedValue = val["computed"].(string)
		}
		computed = append(computed, computedSecret)
	}
	sort.Slice(computed, func(i, j int) bool {
		return computed[i].Name < computed[j].Name
	})
	return computed, nil
}
