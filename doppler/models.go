package doppler

import (
	"encoding/json"
	"sort"
)

type Secret struct {
	Name  string
	Value string
}

func ParseSecrets(response []byte) ([]Secret, error) {
	var result map[string]string
	err := json.Unmarshal(response, &result)
	if err != nil {
		return nil, err
	}

	secrets := make([]Secret, 0)
	for key, value := range result {
		secret := Secret{Name: key, Value: value}
		secrets = append(secrets, secret)
	}
	sort.Slice(secrets, func(i, j int) bool {
		return secrets[i].Name < secrets[j].Name
	})
	return secrets, nil
}
