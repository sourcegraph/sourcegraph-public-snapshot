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
		"single": {
			input: schema.SiteConfiguration{
				AuthProviders: []schema.AuthProviders{
					{HttpHeader: &schema.HTTPHeaderAuthProvider{Type: "http-header"}},
				},
			},
			wantProblems: nil,
		},
		"multiple": {
			input: schema.SiteConfiguration{
				AuthProviders: []schema.AuthProviders{
					{HttpHeader: &schema.HTTPHeaderAuthProvider{Type: "http-header"}},
					{HttpHeader: &schema.HTTPHeaderAuthProvider{Type: "http-header"}},
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
