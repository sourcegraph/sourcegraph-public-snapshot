package auth

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestValidateCustom(t *testing.T) {
	tests := map[string]struct {
		input        conf.UnifiedConfiguration
		wantProblems []string
	}{
		"no auth.providers": {
			input:        conf.UnifiedConfiguration{Core: schema.CoreSiteConfiguration{}},
			wantProblems: []string{"no auth providers set"},
		},
		"empty auth.providers": {
			input:        conf.UnifiedConfiguration{Core: schema.CoreSiteConfiguration{AuthProviders: []schema.AuthProviders{}}},
			wantProblems: []string{"no auth providers set"},
		},
		"single auth provider": {
			input: conf.UnifiedConfiguration{Core: schema.CoreSiteConfiguration{
				AuthProviders: []schema.AuthProviders{
					{Builtin: &schema.BuiltinAuthProvider{Type: "a"}},
				},
			}},
			wantProblems: nil,
		},
		"multiple auth providers": {
			input: conf.UnifiedConfiguration{Core: schema.CoreSiteConfiguration{
				AuthProviders: []schema.AuthProviders{
					{Builtin: &schema.BuiltinAuthProvider{Type: "a"}},
					{Builtin: &schema.BuiltinAuthProvider{Type: "b"}},
				},
			}},
			wantProblems: nil,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			conf.TestValidator(t, test.input, validateConfig, test.wantProblems)
		})
	}
}
