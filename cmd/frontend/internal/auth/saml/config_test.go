package saml

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
		"deprecated saml*": {
			input:        schema.SiteConfiguration{AppURL: "x", SamlSPKey: "x"},
			wantProblems: []string{"must set auth.provider", "saml* properties are deprecated"},
		},
		"duplicates": {
			input: schema.SiteConfiguration{
				AppURL:               "x",
				ExperimentalFeatures: &schema.ExperimentalFeatures{MultipleAuthProviders: "enabled"},
				AuthProviders: []schema.AuthProviders{
					{Saml: &schema.SAMLAuthProvider{Type: "saml", IdentityProviderMetadataURL: "x"}},
					{Saml: &schema.SAMLAuthProvider{Type: "saml", IdentityProviderMetadataURL: "x"}},
				},
			},
			wantProblems: []string{"SAML auth provider at index 1 is duplicate of index 0"},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			conf.TestValidator(t, test.input, validateConfig, test.wantProblems)
		})
	}
}
