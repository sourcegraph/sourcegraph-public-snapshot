// TODO: Consider making config thread-safe.
//go:build !race

package graphqlbackend_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
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
	"github.com/sourcegraph/sourcegraph/internal/observation"
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

// syncConfMocking is a helper to allow mocking site config in a syncronous manner.
// Specifically, observers that are watching config changes will have been updated
// by the time `Update`	returns.
type syncConfMocking struct {
	// cond is used to synchronize between the watchers and the update method.
	cond *sync.Cond
	// watching is true iff this instance of mocking has observers notified of config changes.
	watching bool
	// lastConfig seen is memoized so that `Update` can ensure the value of the config
	// reaches the parameter before it returns.
	lastConfig schema.SiteConfiguration
}

func newSyncConfMocking(t *testing.T) *syncConfMocking {
	t.Helper()
	m := &syncConfMocking{
		cond: sync.NewCond(&sync.Mutex{}),
	}
	m.cond.L.Lock()
	m.watching = true
	m.cond.L.Unlock()
	conf.Watch(m.onConfigChange)
	t.Cleanup(m.Cleanup)
	return m
}

// Update the site config and await the new config to be propagated to the watchers.
func (m *syncConfMocking) Update(c schema.SiteConfiguration) {
	conf.MockAndNotifyWatchers(&conf.Unified{SiteConfiguration: c})
	m.cond.L.Lock()
	defer m.cond.L.Unlock()
	diff := cmp.Diff(m.lastConfig, c)
	for m.watching && diff != "" {
		m.cond.Wait()
		diff = cmp.Diff(m.lastConfig, c)
	}
}

// onConfigChange is used in `conf.Watch`, and wakes up `Update` which is waiting
// on config change to propagate to watchers.
func (m *syncConfMocking) onConfigChange() {
	m.cond.L.Lock()
	defer m.cond.L.Unlock()
	if m.watching {
		m.lastConfig = conf.Get().SiteConfiguration
		m.cond.Broadcast()
	}
}

// Cleanup invalidates the watcher.
func (m *syncConfMocking) Cleanup() {
	m.cond.L.Lock()
	defer m.cond.L.Unlock()
	m.watching = false
	m.cond.Broadcast() // ensure to wake up every waiting goroutine
}

// gatewayResponse that `makeGatewayEndpoint` responds with.
const gatewayResponse = `{
	"Repositories": [
		{"name": "github.com/sourcegraph/sourcegraph"},
		{"name": "npm/sourcegraph/basic-code-intel"}
	],
	"totalCount": 2,
	"limitHit":true
}`

func makeGatewayEndpoint(t *testing.T) string {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, gatewayResponse)
	}))
	t.Cleanup(ts.Close)
	return ts.URL
}

func TestSnippetAttributionReactsToSiteConfigChanges(t *testing.T) {
	// Use a regular HTTP client as default external doer cannot hit localhost.
	guardrails.MockHttpClient = http.DefaultClient
	t.Cleanup(func() { guardrails.MockHttpClient = nil })
	// Starting attribution configuration has no endpoints to use.
	noAttributionConfigured := schema.SiteConfiguration{
		CodyEnabled:        pointers.Ptr(true),
		AttributionEnabled: pointers.Ptr(true),
	}
	confMock := newSyncConfMocking(t)
	confMock.Update(noAttributionConfigured)
	t.Cleanup(func() { confMock.Update(schema.SiteConfiguration{}) })
	// Initialize graphQL schema with snippetAttribution.
	db := dbmocks.NewMockDB()
	ctx := context.Background()
	g := gitserver.NewClient("graphql.test")
	var enterpriseServices enterprise.Services
	require.NoError(t, guardrails.Init(ctx, &observation.TestContext, db, codeintel.Services{}, nil, &enterpriseServices))
	s, err := graphqlbackend.NewSchema(db, g, []graphqlbackend.OptionalResolver{{GuardrailsResolver: enterpriseServices.OptionalResolver.GuardrailsResolver}})
	require.NoError(t, err)
	// Same query runs in every test:
	query := `query SnippetAttribution {
		snippetAttribution(snippet: "sourcegraph.Location(new URL\n\n\n\n\n\n\n\n\n\n", first: 2) {
			nodes {
				repositoryName
			}
		}
	}`

	t.Run("attribution endpoint not configured", func(t *testing.T) {
		response := s.Exec(ctx, query, "", nil)
		require.NotEmpty(t, response.Errors)
		for _, e := range response.Errors {
			require.Equal(t, "Attribution is not initialized. Please update site config.", e.Message)
		}
	})

	t.Run("attribution endpoint explicitly configured", func(t *testing.T) {
		confMock.Update(schema.SiteConfiguration{
			CodyEnabled:        pointers.Ptr(true),
			AttributionEnabled: pointers.Ptr(true),
			AttributionGateway: &schema.AttributionGateway{
				Endpoint:    makeGatewayEndpoint(t),
				AccessToken: "1234",
			},
		})
		t.Cleanup(func() { confMock.Update(noAttributionConfigured) })
		response := s.Exec(ctx, query, "", nil)
		require.Empty(t, response.Errors)
		require.JSONEq(t, `{
			"snippetAttribution": {
				"nodes": [
					{"repositoryName": "github.com/sourcegraph/sourcegraph"},
					{"repositoryName": "npm/sourcegraph/basic-code-intel"}
				]
			}
		}`, string(response.Data))
	})

	t.Run("attribution endpoint defaults to gateway completions config", func(t *testing.T) {
		confMock.Update(schema.SiteConfiguration{
			CodyEnabled:        pointers.Ptr(true),
			AttributionEnabled: pointers.Ptr(true),
			Completions: &schema.Completions{
				AccessToken:       "1234",
				Endpoint:          makeGatewayEndpoint(t),
				Model:             "testing-model",
				ChatModel:         "testing-model-turbo",
				CompletionModel:   "testing-model-turbo",
				Provider:          "sourcegraph",
				PerUserDailyLimit: 1000,
			},
		})
		t.Cleanup(func() { confMock.Update(noAttributionConfigured) })
		response := s.Exec(ctx, query, "", nil)
		require.Empty(t, response.Errors)
		require.JSONEq(t, `{
			"snippetAttribution": {
				"nodes": [
					{"repositoryName": "github.com/sourcegraph/sourcegraph"},
					{"repositoryName": "npm/sourcegraph/basic-code-intel"}
				]
			}
		}`, string(response.Data))
	})

	t.Run("attribution not configured for non-sourcegraph completions", func(t *testing.T) {
		confMock.Update(schema.SiteConfiguration{
			CodyEnabled:        pointers.Ptr(true),
			AttributionEnabled: pointers.Ptr(true),
			Completions: &schema.Completions{
				AccessToken:       "1234",
				Endpoint:          makeGatewayEndpoint(t),
				Model:             "testing-model",
				ChatModel:         "testing-model-turbo",
				CompletionModel:   "testing-model-turbo",
				Provider:          "openai",
				PerUserDailyLimit: 1000,
			},
		})
		t.Cleanup(func() { confMock.Update(noAttributionConfigured) })
		response := s.Exec(ctx, query, "", nil)
		require.NotEmpty(t, response.Errors)
		for _, e := range response.Errors {
			require.Equal(t, "Attribution is not initialized. Please update site config.", e.Message)
		}
	})
}

// mustParseGraphQLSchema is a copy of mustParseGraphQLSchema from graphql package.
// This test needs to use a different package because of otherwise circular dependency
// between guardrails and graphqlbackend.
// TODO(#59995): Extract mustParseGraphQLSchema to a graphqlbackendtest package.
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
