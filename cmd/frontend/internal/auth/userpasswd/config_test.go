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
		"multiple": {
			input: schema.SiteConfiguration{
				ExperimentalFeatures: &schema.ExperimentalFeatures{MultipleAuthProviders: "enabled"},
				AuthProviders: []schema.AuthProviders{
					{Builtin: &schema.BuiltinAuthProvider{Type: "builtin"}},
					{Builtin: &schema.BuiltinAuthProvider{Type: "builtin"}},
				},
			},
			wantProblems: []string{"at most 1"},
		},
		"auth.allowSignup deprecation": {
			input:        schema.SiteConfiguration{AuthAllowSignup: true},
			wantProblems: []string{"auth.allowSignup requires", "auth.allowSignup is deprecated"},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			conf.TestValidator(t, test.input, validateConfig, test.wantProblems)
		})
	}
}
