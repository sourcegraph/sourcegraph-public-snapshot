package auth

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestValidateCustom(t *testing.T) {
	tests := map[string]struct {
		input        schema.SiteConfiguration
		wantProblems []string
	}{
		"no auth.provider": {
			input:        schema.SiteConfiguration{},
			wantProblems: []string{"no auth providers set"},
		},
		"unrecognized auth.provider": {
			input:        schema.SiteConfiguration{AuthProvider: "x"},
			wantProblems: []string{"no auth providers set", "auth.provider is deprecated"},
		},
		"deprecated auth.provider": {
			input:        schema.SiteConfiguration{AuthProvider: "builtin"},
			wantProblems: []string{"auth.provider is deprecated"},
		},
		"auth.provider and auth.providers": {
			input: schema.SiteConfiguration{
				AuthProvider:  "builtin",
				AuthProviders: []schema.AuthProviders{{Builtin: &schema.BuiltinAuthProvider{Type: "builtin"}}},
			},
			wantProblems: []string{"auth.providers takes precedence"},
		},
		"multiple auth providers with experimentalFeatures == nil": {
			input: schema.SiteConfiguration{
				AuthProviders: []schema.AuthProviders{
					{Builtin: &schema.BuiltinAuthProvider{Type: "a"}},
					{Builtin: &schema.BuiltinAuthProvider{Type: "b"}},
				},
			},
			wantProblems: nil,
		},
		"multiple auth providers with experimentalFeatures.multipleAuthProviders unset": {
			input: schema.SiteConfiguration{
				ExperimentalFeatures: &schema.ExperimentalFeatures{},
				AuthProviders: []schema.AuthProviders{
					{Builtin: &schema.BuiltinAuthProvider{Type: "a"}},
					{Builtin: &schema.BuiltinAuthProvider{Type: "b"}},
				},
			},
			wantProblems: nil,
		},
		"multiple auth providers with experimentalFeatures.multipleAuthProviders == enabled": {
			input: schema.SiteConfiguration{
				ExperimentalFeatures: &schema.ExperimentalFeatures{MultipleAuthProviders: "enabled"},
				AuthProviders: []schema.AuthProviders{
					{Builtin: &schema.BuiltinAuthProvider{Type: "a"}},
					{Builtin: &schema.BuiltinAuthProvider{Type: "b"}},
				},
			},
			wantProblems: nil,
		},
		"multiple auth providers with experimentalFeatures.multipleAuthProviders == disabled": {
			input: schema.SiteConfiguration{
				ExperimentalFeatures: &schema.ExperimentalFeatures{MultipleAuthProviders: "disabled"},
				AuthProviders: []schema.AuthProviders{
					{Builtin: &schema.BuiltinAuthProvider{Type: "a"}},
					{Builtin: &schema.BuiltinAuthProvider{Type: "b"}},
				},
			},
			wantProblems: []string{"auth.providers supports only a single"},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			conf.TestValidator(t, test.input, validateConfig, test.wantProblems)
		})
	}
}
