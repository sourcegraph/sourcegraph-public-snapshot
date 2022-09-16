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

	gqlerrors "github.com/graph-gophers/graphql-go/errors"
	"github.com/inconshreveable/log15"
	sglog "github.com/sourcegraph/log"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestMain(m *testing.M) {
	flag.Parse()
	if !testing.Verbose() {
		log15.Root().SetHandler(log15.DiscardHandler())
		log.SetOutput(io.Discard)
		logtest.InitWithLevel(m, sglog.LevelNone)
	} else {
		logtest.Init(m)
	}
	os.Exit(m.Run())
}

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
	db := database.NewMockDB()
	repos := database.NewMockRepoStore()
	repos.GetByNameFunc.SetDefaultReturn(&types.Repo{ID: 2, Name: "github.com/gorilla/mux"}, nil)
	db.ReposFunc.SetDefaultReturn(repos)
	RunTests(t, []*Test{
		{
			Schema: mustParseGraphQLSchema(t, db),
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
	db := database.NewMockDB()
	// This test exists purely to remove some non determinism in our tests
	// run. The To* resolvers are stored in a map in our graphql
	// implementation => the order we call them is non deterministic =>
	// codecov coverage reports are noisy.
	resolvers := []any{
		&FileMatchResolver{db: db},
		&NamespaceResolver{},
		&NodeResolver{},
		&RepositoryResolver{db: db, logger: logtest.Scoped(t)},
		&CommitSearchResultResolver{},
		&gitRevSpec{},
		&settingsSubject{},
		&statusMessageResolver{db: db},
	}
	for _, r := range resolvers {
		typ := reflect.TypeOf(r)
		t.Run(typ.Name(), func(t *testing.T) {
			for i := 0; i < typ.NumMethod(); i++ {
				if name := typ.Method(i).Name; strings.HasPrefix(name, "To") {
					reflect.ValueOf(r).MethodByName(name).Call(nil)
				}
			}
		})
	}

	t.Run("GitTreeEntryResolver", func(t *testing.T) {
		blobStat, err := os.Stat("graphqlbackend_test.go")
		if err != nil {
			t.Fatalf("unexpected error opening file: %s", err)
		}
		blobEntry := &GitTreeEntryResolver{db: db, stat: blobStat}
		if _, isBlob := blobEntry.ToGitBlob(); !isBlob {
			t.Errorf("expected blobEntry to be blob")
		}
		if _, isTree := blobEntry.ToGitTree(); isTree {
			t.Errorf("expected blobEntry to be blob, but is tree")
		}

		treeStat, err := os.Stat(".")
		if err != nil {
			t.Fatalf("unexpected error opening directory: %s", err)
		}
		treeEntry := &GitTreeEntryResolver{db: db, stat: treeStat}
		if _, isBlob := treeEntry.ToGitBlob(); isBlob {
			t.Errorf("expected treeEntry to be tree, but is blob")
		}
		if _, isTree := treeEntry.ToGitTree(); !isTree {
			t.Errorf("expected treeEntry to be tree")
		}
	})
}

func TestAffiliatedRepositories(t *testing.T) {
	resetMocks()
	rcache.SetupForTest(t)
	users := database.NewMockUserStore()
	users.TagsFunc.SetDefaultReturn(map[string]bool{}, nil)
	users.GetByIDFunc.SetDefaultHook(func(_ context.Context, userID int32) (*types.User, error) {
		return &types.User{ID: userID, SiteAdmin: userID == 2}, nil
	})
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1, SiteAdmin: true}, nil)

	externalServices := database.NewMockExternalServiceStore()
	externalServices.ListFunc.SetDefaultReturn(
		[]*types.ExternalService{
			{
				ID:          1,
				Kind:        extsvc.KindGitHub,
				DisplayName: "github",
				Config:      extsvc.NewEmptyConfig(),
			},
			{
				ID:          2,
				Kind:        extsvc.KindGitLab,
				DisplayName: "gitlab",
				Config:      extsvc.NewEmptyConfig(),
			},
			{
				ID:     3,
				Kind:   extsvc.KindBitbucketCloud, // unsupported, should be ignored
				Config: extsvc.NewEmptyConfig(),
			},
		},
		nil,
	)
	externalServices.GetByIDFunc.SetDefaultHook(func(_ context.Context, id int64) (*types.ExternalService, error) {
		switch id {
		case 1:
			return &types.ExternalService{
				ID:          1,
				Kind:        extsvc.KindGitHub,
				DisplayName: "github",
				Config:      extsvc.NewEmptyConfig(),
			}, nil
		case 2:
			return &types.ExternalService{
				ID:          2,
				Kind:        extsvc.KindGitLab,
				DisplayName: "gitlab",
				Config:      extsvc.NewEmptyConfig(),
			}, nil
		}
		return nil, nil
	})

	db := database.NewMockDB()
	db.UsersFunc.SetDefaultReturn(users)
	db.ExternalServicesFunc.SetDefaultReturn(externalServices)

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
			Schema:  mustParseGraphQLSchema(t, db),
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
			Schema:  mustParseGraphQLSchema(t, db),
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
					Path:          []any{"affiliatedRepositories"},
					Message:       "must be authenticated as user with id 1",
					ResolverError: &backend.InsufficientAuthorizationError{Message: fmt.Sprintf("must be authenticated as user with id %d", 1)},
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

	// When one code host fails, return its errors and also the nodes from the other code host.
	RunTests(t, []*Test{
		{
			Context: ctx,
			Schema:  mustParseGraphQLSchema(t, db),
			Query: `
			{
				affiliatedRepositories(
					namespace: "VXNlcjox"
				) {
					codeHostErrors
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
						"codeHostErrors": ["Error from gitlab: unexpected response from GitLab API (/api/v4/projects?archived=no&membership=true&per_page=40): HTTP error status 401"],
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
			Schema:  mustParseGraphQLSchema(t, db),
			Query: `
			{
				affiliatedRepositories(
					namespace: "VXNlcjox"
				) {
					codeHostErrors
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
					Path:          []any{"affiliatedRepositories", "codeHostErrors"},
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
