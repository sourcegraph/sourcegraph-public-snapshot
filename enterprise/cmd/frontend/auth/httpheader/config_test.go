package httpheader

import (
	"testing"

	"sourcegraph.com/pkg/conf"
	"sourcegraph.com/schema"
)

func TestValidateCustom(t *testing.T) {
	tests := map[string]struct {
		input        conf.Unified
		wantProblems []string
	}{
		"single": {
			input: conf.Unified{Critical: schema.CriticalConfiguration{
				AuthProviders: []schema.AuthProviders{
					{HttpHeader: &schema.HTTPHeaderAuthProvider{Type: "http-header"}},
				},
			}},
			wantProblems: nil,
		},
		"multiple": {
			input: conf.Unified{Critical: schema.CriticalConfiguration{
				AuthProviders: []schema.AuthProviders{
					{HttpHeader: &schema.HTTPHeaderAuthProvider{Type: "http-header"}},
					{HttpHeader: &schema.HTTPHeaderAuthProvider{Type: "http-header"}},
				},
			}},
			wantProblems: []string{"at most 1"},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			conf.TestValidator(t, test.input, validateConfig, test.wantProblems)
		})
	}
}
