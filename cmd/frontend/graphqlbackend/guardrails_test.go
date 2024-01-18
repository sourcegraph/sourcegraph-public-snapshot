package graphqlbackend

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestGuardrails(t *testing.T) {
	// This test is just asserting that our interface is correct. It seems
	// graphql-go only does the schema check if your interface is non-nil.
	_, err := NewSchema(nil, nil, []OptionalResolver{{GuardrailsResolver: guardrailsFake{}}})
	if err != nil {
		t.Fatal(err)
	}
}

type guardrailsFake struct{}

func (guardrailsFake) SnippetAttribution(context.Context, *SnippetAttributionArgs) (SnippetAttributionConnectionResolver, error) {
	return nil, errors.New("fake")
}

func TestGuardrailsFeatureEnabled(t *testing.T) {
	db := dbmocks.NewMockDB()
	ctx := context.Background()
	conf.Mock(&conf.Unified{
		SiteConfiguration: schema.SiteConfiguration{
			CodyEnabled:        pointers.Ptr(true),
			AttributionEnabled: pointers.Ptr(true),
		},
	})
	t.Cleanup(func() { conf.Mock(nil) })
	RunTest(t, &Test{
		Schema:  mustParseGraphQLSchema(t, db),
		Context: ctx,
		Query: `query AttributionEnabled {
			site {
				codyConfigFeatures {
					attribution
				}
			}
		}`,
		ExpectedResult: `{
			"site": {
				"codyConfigFeatures": {
					"attribution": true
				}
			}
		}
		`,
	})
}

func TestGuardrailsFeatureDisabled(t *testing.T) {
	db := dbmocks.NewMockDB()
	ctx := context.Background()
	test := &Test{
		Schema:  mustParseGraphQLSchema(t, db),
		Context: ctx,
		Query: `query AttributionEnabled {
			site {
				codyConfigFeatures {
					attribution
				}
			}
		}`,
		ExpectedResult: `{
			"site": {
				"codyConfigFeatures": {
					"attribution": false
				}
			}
		}
		`,
	}
	t.Run("default - not configured", func(t *testing.T) {
		conf.Mock(&conf.Unified{
			SiteConfiguration: schema.SiteConfiguration{
				CodyEnabled: pointers.Ptr(true),
			},
		})
		t.Cleanup(func() { conf.Mock(nil) })
		RunTest(t, test)
	})
	t.Run("explicitly disabled", func(t *testing.T) {
		conf.Mock(&conf.Unified{
			SiteConfiguration: schema.SiteConfiguration{
				CodyEnabled:        pointers.Ptr(true),
				AttributionEnabled: pointers.Ptr(false),
			},
		})
		t.Cleanup(func() { conf.Mock(nil) })
		RunTest(t, test)
	})
}

func TestGuardrailsFeatureCodyDisabled(t *testing.T) {
	db := dbmocks.NewMockDB()
	ctx := context.Background()
	conf.Mock(&conf.Unified{
		SiteConfiguration: schema.SiteConfiguration{},
	})
	t.Cleanup(func() { conf.Mock(nil) })
	RunTest(t, &Test{
		Schema:  mustParseGraphQLSchema(t, db),
		Context: ctx,
		Query: `query AttributionEnabled {
			site {
				codyConfigFeatures {
					attribution
				}
			}
		}`,
		ExpectedResult: `{
			"site": {
				"codyConfigFeatures": null
			}
		}
		`,
	})
}
