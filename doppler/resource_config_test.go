package doppler

import "testing"

// The API returns `"inherits": []` for a config that inherits nothing, which
// unmarshals to an empty non-nil slice, while inheritsArgToDescriptors returns
// nil for an empty list. reflect.DeepEqual reports those as unequal, so the
// previous comparison issued an UpdateConfigInherits call on every create -
// which fails outright on a workplace without the config-inheritance
// entitlement.
func TestConfigDescriptorsEqual(t *testing.T) {
	descriptor := ConfigDescriptor{Project: "backend", Config: "stg"}
	other := ConfigDescriptor{Project: "backend", Config: "dev"}

	cases := []struct {
		name string
		a    []ConfigDescriptor
		b    []ConfigDescriptor
		want bool
	}{
		{"nil and empty both mean inherits nothing", nil, []ConfigDescriptor{}, true},
		{"empty and nil both mean inherits nothing", []ConfigDescriptor{}, nil, true},
		{"both nil", nil, nil, true},
		{"both empty", []ConfigDescriptor{}, []ConfigDescriptor{}, true},
		{"same single descriptor", []ConfigDescriptor{descriptor}, []ConfigDescriptor{descriptor}, true},
		{"different descriptor", []ConfigDescriptor{descriptor}, []ConfigDescriptor{other}, false},
		{"added descriptor", nil, []ConfigDescriptor{descriptor}, false},
		{"removed descriptor", []ConfigDescriptor{descriptor}, nil, false},
		{"different length", []ConfigDescriptor{descriptor}, []ConfigDescriptor{descriptor, other}, false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := configDescriptorsEqual(c.a, c.b); got != c.want {
				t.Errorf("configDescriptorsEqual(%v, %v) = %v, want %v", c.a, c.b, got, c.want)
			}
		})
	}
}
