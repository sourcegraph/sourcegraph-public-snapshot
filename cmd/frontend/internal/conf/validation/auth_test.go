package validation

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestValidateAuthProvidersConfig(t *testing.T) {
	tests := map[string]struct {
		input        conf.Unified
		wantProblems conf.Problems
	}{
		"no auth.providers": {
			input:        conf.Unified{SiteConfiguration: schema.SiteConfiguration{}},
			wantProblems: conf.NewSiteProblems("no auth providers set"),
		},
		"empty auth.providers": {
			input:        conf.Unified{SiteConfiguration: schema.SiteConfiguration{AuthProviders: []schema.AuthProviders{}}},
			wantProblems: conf.NewSiteProblems("no auth providers set"),
		},
		"single auth provider": {
			input: conf.Unified{SiteConfiguration: schema.SiteConfiguration{
				AuthProviders: []schema.AuthProviders{
					{Builtin: &schema.BuiltinAuthProvider{Type: "a"}},
				},
			}},
			wantProblems: nil,
		},
		"multiple auth providers": {
			input: conf.Unified{SiteConfiguration: schema.SiteConfiguration{
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
			conf.TestValidator(t, test.input, validateAuthProvidersConfig, test.wantProblems)
		})
	}
}

func TestValidateBuiltinAuthConfig(t *testing.T) {
	tests := map[string]struct {
		input        conf.Unified
		wantProblems conf.Problems
	}{
		"single": {
			input: conf.Unified{SiteConfiguration: schema.SiteConfiguration{
				AuthProviders: []schema.AuthProviders{
					{Builtin: &schema.BuiltinAuthProvider{Type: "builtin"}},
				},
			}},
			wantProblems: nil,
		},
		"multiple": {
			input: conf.Unified{SiteConfiguration: schema.SiteConfiguration{
				AuthProviders: []schema.AuthProviders{
					{Builtin: &schema.BuiltinAuthProvider{Type: "builtin"}},
					{Builtin: &schema.BuiltinAuthProvider{Type: "builtin"}},
				},
			}},
			wantProblems: conf.NewSiteProblems("at most 1"),
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			conf.TestValidator(t, test.input, validateBuiltinAuthConfig, test.wantProblems)
		})
	}
}
