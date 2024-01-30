package graphqlbackend_test

import (
	"context"
	"testing"

	"github.com/graph-gophers/graphql-go"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/guardrails"
	"github.com/sourcegraph/sourcegraph/internal/codeintel"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestGuardrails(t *testing.T) {
	// This test is just asserting that our interface is correct. It seems
	// graphql-go only does the schema check if your interface is non-nil.
	_, err := graphqlbackend.NewSchema(nil, nil, []graphqlbackend.OptionalResolver{{GuardrailsResolver: guardrailsFake{}}})
	if err != nil {
		t.Fatal(err)
	}
}

type guardrailsFake struct{}

func (guardrailsFake) SnippetAttribution(context.Context, *graphqlbackend.SnippetAttributionArgs) (graphqlbackend.SnippetAttributionConnectionResolver, error) {
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
	graphqlbackend.RunTest(t, &graphqlbackend.Test{
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
	test := &graphqlbackend.Test{
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
		graphqlbackend.RunTest(t, test)
	})
	t.Run("explicitly disabled", func(t *testing.T) {
		conf.Mock(&conf.Unified{
			SiteConfiguration: schema.SiteConfiguration{
				CodyEnabled:        pointers.Ptr(true),
				AttributionEnabled: pointers.Ptr(false),
			},
		})
		t.Cleanup(func() { conf.Mock(nil) })
		graphqlbackend.RunTest(t, test)
	})
}

func TestGuardrailsFeatureCodyDisabled(t *testing.T) {
	db := dbmocks.NewMockDB()
	ctx := context.Background()
	conf.Mock(&conf.Unified{
		SiteConfiguration: schema.SiteConfiguration{},
	})
	t.Cleanup(func() { conf.Mock(nil) })
	graphqlbackend.RunTest(t, &graphqlbackend.Test{
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

func TestSnippetAttributionReactsToSiteConfigChanges(t *testing.T) {
	db := dbmocks.NewMockDB()
	ctx := context.Background()
	conf.Mock(&conf.Unified{
		SiteConfiguration: schema.SiteConfiguration{
			CodyEnabled:        pointers.Ptr(true),
			AttributionEnabled: pointers.Ptr(true),
			Completions: &schema.Completions{
				AccessToken: "1234",
				Endpoint: "https://example.com",
				Model: "testing-model",
				ChatModel: "testing-model-turbo",
				CompletionModel: "testing-model-turbo",
				Provider: "hermetic-test",
				PerUserDailyLimit: 1000,
			},
		},
	})
	t.Cleanup(func() { conf.Mock(nil) })
	var enterpriseServices enterprise.Services
	g := gitserver.NewClient("graphql.test")
	require.NoError(t, guardrails.Init(ctx, nil, db, codeintel.Services{}, nil, &enterpriseServices))
	schema, err := graphqlbackend.NewSchema(db, g, []graphqlbackend.OptionalResolver{{GuardrailsResolver: enterpriseServices.OptionalResolver.GuardrailsResolver}})
	require.NoError(t, err)
	query := `query SnippetAttribution {
		snippetAttribution(snippet: "new URL(", first: 2) {
			nodes {
				repositoryName
			}
		}
	}`
	// TODO: Mock gateway
	t.Run("attribution endpoint not configured", func(t *testing.T) {
		response := schema.Exec(ctx, query, "", nil)
		require.NotEmpty(t, response.Errors)
		for _, e := range response.Errors {
			require.Equal(t, "Attribution is not initialized. Please update site config.", e.Message)
		}
	})
}

func mustParseGraphQLSchema(t *testing.T, db database.DB) *graphql.Schema {
	t.Helper()
	gitserverClient := gitserver.NewClient("graphql.test")
	parsedSchema, parseSchemaErr := graphqlbackend.NewSchema(
		db,
		gitserverClient,
		[]graphqlbackend.OptionalResolver{},
		graphql.MaxDepth(conf.RateLimits().GraphQLMaxDepth),
	)
	if parseSchemaErr != nil {
		t.Fatal(parseSchemaErr)
	}
	return parsedSchema
}
