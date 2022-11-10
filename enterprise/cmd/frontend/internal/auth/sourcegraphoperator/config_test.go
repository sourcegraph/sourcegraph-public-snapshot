package sourcegraphoperator

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name         string
		siteConfig   schema.SiteConfiguration
		wantProblems conf.Problems
	}{
		{
			name: "duplicates",
			siteConfig: schema.SiteConfiguration{
				ExternalURL: "https://example.com",
				AuthProviders: []schema.AuthProviders{
					{
						SourcegraphOperator: &schema.SourcegraphOperatorAuthProvider{
							Issuer: "https://example.com/alice",
							Type:   ProviderType,
						},
					}, {
						SourcegraphOperator: &schema.SourcegraphOperatorAuthProvider{
							Issuer: "https://example.com/bob",
							Type:   ProviderType,
						},
					},
				},
			},
			wantProblems: conf.NewSiteProblems("Sourcegraph Operator authentication provider at index 1 is duplicate of index 0, ignoring"),
		},
		{
			name: "no externalURL",
			siteConfig: schema.SiteConfiguration{
				AuthProviders: []schema.AuthProviders{
					{
						SourcegraphOperator: &schema.SourcegraphOperatorAuthProvider{
							Issuer: "https://example.com/alice",
							Type:   ProviderType,
						},
					},
				},
			},
			wantProblems: conf.NewSiteProblems("Sourcegraph Operator authentication provider requires `externalURL` to be set to the external URL of your site (example: https://sourcegraph.example.com)"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			conf.TestValidator(
				t,
				conf.Unified{
					SiteConfiguration: test.siteConfig,
				},
				validateConfig,
				test.wantProblems,
			)
		})
	}
}
