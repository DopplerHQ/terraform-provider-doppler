package doppler

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"
)

type APIContext struct {
	Host      string
	APIKey    string
	VerifyTLS bool
}

func (ctx *APIContext) GetId() string {
	digester := sha256.New()
	fmt.Fprint(digester, ctx.Host)
	fmt.Fprint(digester, ctx.APIKey)
	fmt.Fprint(digester, ctx.VerifyTLS)
	return fmt.Sprintf("%x", digester.Sum(nil))
}

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
