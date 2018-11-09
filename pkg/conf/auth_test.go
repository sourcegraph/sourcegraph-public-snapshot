package conf

import (
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/schema"
)

func TestAuthPublic(t *testing.T) {
	tests := map[string]struct {
		input UnifiedConfiguration
		want  bool
	}{
		"false": {
			input: UnifiedConfiguration{Core: schema.CoreSiteConfiguration{AuthPublic: false}},
			want:  false,
		},
		"true, no auth provider": {
			input: UnifiedConfiguration{Core: schema.CoreSiteConfiguration{AuthPublic: true}},
			want:  false,
		},
		"true, non-builtin auth provider": {
			input: UnifiedConfiguration{Core: schema.CoreSiteConfiguration{AuthPublic: true, AuthProviders: []schema.AuthProviders{{Openidconnect: &schema.OpenIDConnectAuthProvider{}}}}},
			want:  false,
		},
		"true, builtin auth provider": {
			input: UnifiedConfiguration{Core: schema.CoreSiteConfiguration{AuthPublic: true, AuthProviders: []schema.AuthProviders{{Builtin: &schema.BuiltinAuthProvider{}}}}},
			want:  true,
		},
		"false, builtin auth provider": {
			input: UnifiedConfiguration{Core: schema.CoreSiteConfiguration{AuthPublic: false, AuthProviders: []schema.AuthProviders{{Builtin: &schema.BuiltinAuthProvider{}}}}},
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
