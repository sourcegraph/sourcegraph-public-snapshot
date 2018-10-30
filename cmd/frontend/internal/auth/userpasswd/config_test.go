package userpasswd

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
		"single": {
			input: schema.SiteConfiguration{
				AuthProviders: []schema.AuthProviders{
					{Builtin: &schema.BuiltinAuthProvider{Type: "builtin"}},
				},
			},
			wantProblems: nil,
		},
		"multiple": {
			input: schema.SiteConfiguration{
				AuthProviders: []schema.AuthProviders{
					{Builtin: &schema.BuiltinAuthProvider{Type: "builtin"}},
					{Builtin: &schema.BuiltinAuthProvider{Type: "builtin"}},
				},
			},
			wantProblems: []string{"at most 1"},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			conf.TestValidator(t, test.input, validateConfig, test.wantProblems)
		})
	}
}
