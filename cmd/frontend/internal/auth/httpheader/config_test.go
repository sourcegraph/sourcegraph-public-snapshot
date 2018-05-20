package httpheader

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
					{HttpHeader: &schema.HTTPHeaderAuthProvider{Type: "http-header"}},
					{HttpHeader: &schema.HTTPHeaderAuthProvider{Type: "http-header"}},
				},
			},
			wantProblems: []string{"at most 1"},
		},
		"deprecated auth.userIdentityHTTPHeader": {
			input:        schema.SiteConfiguration{AuthUserIdentityHTTPHeader: "x"},
			wantProblems: []string{"must set auth.provider", "auth.userIdentityHTTPHeader is deprecated"},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			conf.TestValidator(t, test.input, validateConfig, test.wantProblems)
		})
	}
}
