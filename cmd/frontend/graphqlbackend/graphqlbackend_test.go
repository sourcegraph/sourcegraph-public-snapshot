package graphqlbackend

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/cockroachdb/errors"
	gqlerrors "github.com/graph-gophers/graphql-go/errors"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func BenchmarkPrometheusFieldName(b *testing.B) {
	tests := [][3]string{
		{"Query", "settingsSubject", "settingsSubject"},
		{"SearchResultMatch", "highlights", "highlights"},
		{"TreeEntry", "isSingleChild", "isSingleChild"},
		{"NoMatch", "NotMatch", "other"},
	}
	for i, t := range tests {
		typeName, fieldName, want := t[0], t[1], t[2]
		b.Run(fmt.Sprintf("test-%v", i), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				got := prometheusFieldName(typeName, fieldName)
				if got != want {
					b.Fatalf("got %q want %q", got, want)
				}
			}
		})
	}
}

func TestRepository(t *testing.T) {
	resetMocks()
	database.Mocks.Repos.MockGetByName(t, "github.com/gorilla/mux", 2)
	RunTests(t, []*Test{
		{
			Schema: mustParseGraphQLSchema(t),
			Query: `
				{
					repository(name: "github.com/gorilla/mux") {
						name
					}
				}
			`,
			ExpectedResult: `
				{
					"repository": {
						"name": "github.com/gorilla/mux"
					}
				}
			`,
		},
	})
}

func TestResolverTo(t *testing.T) {
	db := new(dbtesting.MockDB)
	// This test exists purely to remove some non determinism in our tests
	// run. The To* resolvers are stored in a map in our graphql
	// implementation => the order we call them is non deterministic =>
	// codecov coverage reports are noisy.
	resolvers := []interface{}{
		&FileMatchResolver{db: db},
		&GitTreeEntryResolver{db: db},
		&NamespaceResolver{},
		&NodeResolver{},
		&RepositoryResolver{db: db},
		&CommitSearchResultResolver{},
		&gitRevSpec{},
		&repositorySuggestionResolver{},
		&symbolSuggestionResolver{},
		&languageSuggestionResolver{},
		&settingsSubject{},
		&statusMessageResolver{db: db},
	}
	for _, r := range resolvers {
		typ := reflect.TypeOf(r)
		for i := 0; i < typ.NumMethod(); i++ {
			if name := typ.Method(i).Name; strings.HasPrefix(name, "To") {
				reflect.ValueOf(r).MethodByName(name).Call(nil)
			}
		}
	}
}

func TestMain(m *testing.M) {
	flag.Parse()
	if !testing.Verbose() {
		log15.Root().SetHandler(log15.LvlFilterHandler(log15.LvlError, log15.Root().GetHandler()))
		log.SetOutput(io.Discard)
	}
	os.Exit(m.Run())
}

func TestAffiliatedRepositories(t *testing.T) {
	resetMocks()
	rcache.SetupForTest(t)
	database.Mocks.Users.Tags = func(ctx context.Context, userID int32) (map[string]bool, error) {
		return map[string]bool{}, nil
	}
	database.Mocks.ExternalServices.List = func(opt database.ExternalServicesListOptions) ([]*types.ExternalService, error) {
		return []*types.ExternalService{
			{
				ID:          1,
				Kind:        extsvc.KindGitHub,
				DisplayName: "github",
			},
			{
				ID:          2,
				Kind:        extsvc.KindGitLab,
				DisplayName: "gitlab",
			},
			{
				ID:   3,
				Kind: extsvc.KindBitbucketCloud, // unsupported, should be ignored
			},
		}, nil
	}
	database.Mocks.ExternalServices.GetByID = func(id int64) (*types.ExternalService, error) {
		switch id {
		case 1:
			return &types.ExternalService{
				ID:          1,
				Kind:        extsvc.KindGitHub,
				DisplayName: "github",
			}, nil
		case 2:
			return &types.ExternalService{
				ID:          2,
				Kind:        extsvc.KindGitLab,
				DisplayName: "gitlab",
			}, nil
		}
		return nil, nil
	}
	database.Mocks.Users.GetByID = func(ctx context.Context, userID int32) (*types.User, error) {
		return &types.User{
			ID:        userID,
			SiteAdmin: userID == 2,
		}, nil
	}
	database.Mocks.Users.GetByCurrentAuthUser = func(ctx context.Context) (*types.User, error) {
		return &types.User{ID: 1, SiteAdmin: true}, nil
	}

	// Map from path rou
	httpResponder := map[string]roundTripFunc{
		// github
		"/api/v3/user/repos": func(r *http.Request) (*http.Response, error) {
			buf := &bytes.Buffer{}
			enc := json.NewEncoder(buf)
			page := r.URL.Query().Get("page")
			if page == "1" {
				if err := enc.Encode([]githubRepository{
					{
						FullName: "test-user/test",
						Private:  false,
					},
				}); err != nil {
					t.Fatal(err)
				}
			}
			// Stop on the second page
			if page == "2" {
				if err := enc.Encode([]githubRepository{}); err != nil {
					t.Fatal(err)
				}
			}
			return &http.Response{
				Body:       io.NopCloser(buf),
				StatusCode: http.StatusOK,
			}, nil
		},

		// gitlab
		"/api/v4/projects": func(r *http.Request) (*http.Response, error) {
			buf := &bytes.Buffer{}
			enc := json.NewEncoder(buf)
			if err := enc.Encode([]gitlabRepository{
				{
					PathWithNamespace: "test-user2/test2",
					Visibility:        "public",
				},
			}); err != nil {
				t.Fatal(err)
			}
			return &http.Response{
				Body:       io.NopCloser(buf),
				StatusCode: http.StatusOK,
			}, nil
		},
	}

	cf = httpcli.NewFactory(
		nil,
		func(c *http.Client) error {
			c.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
				fn := httpResponder[r.URL.Path]
				if fn == nil {
					t.Fatalf("unexpected path: %s", r.URL.Path)
				}
				return fn(r)
			})
			return nil
		},
	)

	ctx := context.Background()
	ctx = actor.WithActor(ctx, &actor.Actor{
		UID: 1,
	})

	RunTests(t, []*Test{
		{
			Context: ctx,
			Schema:  mustParseGraphQLSchema(t),
			Query: `
			{
				affiliatedRepositories(
					namespace: "VXNlcjox"
				) {
					nodes {
						name,
						private,
						codeHost {
							displayName
						}
					}
				}
			}
			`,
			ExpectedResult: `
				{
					"affiliatedRepositories": {
						"nodes": [
							{
								"name": "test-user/test",
								"private": false,
								"codeHost": {
									"displayName": "github"
								}
							},
							{
								"name": "test-user2/test2",
								"private": false,
								"codeHost": {
									"displayName": "gitlab"
								}
							}
						]
					}
				}
			`,
		},
	})

	// Confirm that a site admin cannot list someone else's repos
	ctx = actor.WithActor(ctx, &actor.Actor{
		UID: 2,
	})

	RunTests(t, []*Test{
		{
			Context: ctx,
			Schema:  mustParseGraphQLSchema(t),
			Query: `
			{
				affiliatedRepositories(
					namespace: "VXNlcjox"
				) {
					nodes {
						name,
						private,
						codeHost {
							displayName
						}
					}
				}
			}
			`,
			ExpectedResult: `null`,
			ExpectedErrors: []*gqlerrors.QueryError{
				{
					Path:          []interface{}{"affiliatedRepositories"},
					Message:       "Must be authenticated as user with id 1",
					ResolverError: &backend.InsufficientAuthorizationError{Message: fmt.Sprintf("Must be authenticated as user with id %d", 1)},
				},
			},
		},
	})

	// One code host failing should not break everything
	ctx = actor.WithActor(ctx, &actor.Actor{
		UID: 1,
	})

	// Make gitlab break
	httpResponder["/api/v4/projects"] = func(request *http.Request) (*http.Response, error) {
		buf := &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		if err := enc.Encode([]gitlabRepository{
			{
				PathWithNamespace: "test-user2/test2",
				Visibility:        "public",
			},
		}); err != nil {
			t.Fatal(err)
		}
		return &http.Response{
			Body:       nil,
			StatusCode: http.StatusUnauthorized,
		}, nil
	}

	RunTests(t, []*Test{
		{
			Context: ctx,
			Schema:  mustParseGraphQLSchema(t),
			Query: `
			{
				affiliatedRepositories(
					namespace: "VXNlcjox"
				) {
					nodes {
						name,
						private,
						codeHost {
							displayName
						}
					}
				}
			}
			`,
			ExpectedResult: `
				{
					"affiliatedRepositories": {
						"nodes": [
							{
								"name": "test-user/test",
								"private": false,
								"codeHost": {
									"displayName": "github"
								}
							}
						]
					}
				}
			`,
		},
	})

	// Both code hosts failing is an error
	// Make github break too
	httpResponder["/api/v3/user/repos"] = func(request *http.Request) (*http.Response, error) {
		buf := &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		if err := enc.Encode([]gitlabRepository{
			{
				PathWithNamespace: "test-user2/test2",
				Visibility:        "public",
			},
		}); err != nil {
			t.Fatal(err)
		}
		return &http.Response{
			Body:       nil,
			StatusCode: http.StatusUnauthorized,
		}, nil
	}

	RunTests(t, []*Test{
		{
			Context: ctx,
			Schema:  mustParseGraphQLSchema(t),
			Query: `
			{
				affiliatedRepositories(
					namespace: "VXNlcjox"
				) {
					nodes {
						name,
						private,
						codeHost {
							displayName
						}
					}
				}
			}
			`,
			ExpectedResult: `null`,
			ExpectedErrors: []*gqlerrors.QueryError{
				{
					Path:          []interface{}{"affiliatedRepositories", "nodes"},
					Message:       "failed to fetch from any code host",
					ResolverError: errors.New("failed to fetch from any code host"),
				},
			},
		},
	})
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

// copied from the github client, just the fields we need
type githubRepository struct {
	FullName string `json:"full_name"`
	Private  bool   `json:"private"`
}

type gitlabRepository struct {
	Visibility        string `json:"visibility"`
	ID                int    `json:"id"`
	PathWithNamespace string `json:"path_with_namespace"`
}
