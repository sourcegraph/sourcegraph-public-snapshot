package conf

import (
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/schema"
)

func TestAuthPublic(t *testing.T) {
	tests := map[string]struct {
		input Unified
		want  bool
	}{
		"false": {
			input: Unified{Critical: schema.CriticalConfiguration{AuthPublic: false}},
			want:  false,
		},
		"true, no auth provider": {
			input: Unified{Critical: schema.CriticalConfiguration{AuthPublic: true}},
			want:  false,
		},
		"true, non-builtin auth provider": {
			input: Unified{Critical: schema.CriticalConfiguration{AuthPublic: true, AuthProviders: []schema.AuthProviders{{Openidconnect: &schema.OpenIDConnectAuthProvider{}}}}},
			want:  false,
		},
		"true, builtin auth provider": {
			input: Unified{Critical: schema.CriticalConfiguration{AuthPublic: true, AuthProviders: []schema.AuthProviders{{Builtin: &schema.BuiltinAuthProvider{}}}}},
			want:  true,
		},
		"false, builtin auth provider": {
			input: Unified{Critical: schema.CriticalConfiguration{AuthPublic: false, AuthProviders: []schema.AuthProviders{{Builtin: &schema.BuiltinAuthProvider{}}}}},
			want:  false,
		},
	}
	for label, test := range tests {
		t.Run(label, func(t *testing.T) {
			got := authPublic(&test.input)
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("got %v, want %v", got, test.want)
			}
		})
	}
}
