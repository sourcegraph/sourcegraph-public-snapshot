package sourcegraphoperator

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/cloud"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestValidateConfig(t *testing.T) {
	cloud.MockSiteConfig(
		t,
		&cloud.SchemaSiteConfig{
			AuthProviders: &cloud.SchemaAuthProviders{
				SourcegraphOperator: &cloud.SchemaAuthProviderSourcegraphOperator{
					Issuer: "https://example.com/alice",
				},
			},
		},
	)

	conf.TestValidator(
		t,
		conf.Unified{
			SiteConfiguration: schema.SiteConfiguration{},
		},
		validateConfig,
		conf.NewSiteProblems("Sourcegraph Operator authentication provider requires `externalURL` to be set to the external URL of your site (example: https://sourcegraph.example.com)"),
	)
}
