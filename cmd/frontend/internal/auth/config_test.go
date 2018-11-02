package auth

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestValidateCustom(t *testing.T) {
	tests := map[string]struct {
		input        schema.CoreSiteConfiguration
		wantProblems []string
	}{
		"no auth.providers": {
			input:        schema.CoreSiteConfiguration{},
			wantProblems: []string{"no auth providers set"},
		},
		"empty auth.providers": {
			input: schema.CoreSiteConfiguration{
				AuthProviders: []schema.AuthProviders{},
			},
			wantProblems: []string{"no auth providers set"},
		},
		"single auth provider": {
			input: schema.CoreSiteConfiguration{
				AuthProviders: []schema.AuthProviders{
					{Builtin: &schema.BuiltinAuthProvider{Type: "a"}},
				},
			},
			wantProblems: nil,
		},
		"multiple auth providers": {
			input: schema.CoreSiteConfiguration{
				AuthProviders: []schema.AuthProviders{
					{Builtin: &schema.BuiltinAuthProvider{Type: "a"}},
					{Builtin: &schema.BuiltinAuthProvider{Type: "b"}},
				},
			},
			wantProblems: nil,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			config := conftypes.SiteConfiguration{
				CoreSiteConfiguration: test.input,
			}
			conf.TestValidator(t, config, validateConfig, test.wantProblems)
		})
	}
}
