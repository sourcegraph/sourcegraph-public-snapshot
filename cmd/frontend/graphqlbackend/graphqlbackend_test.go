package graphqlbackend

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"

	"github.com/graph-gophers/graphql-go/gqltesting"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
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
	gqltesting.RunTests(t, []*gqltesting.Test{
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
		&versionContextResolver{},
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
		log.SetOutput(ioutil.Discard)
	}
	os.Exit(m.Run())
}

func TestAffiliatedRepositories(t *testing.T) {
	resetMocks()
	database.Mocks.Users.HasTag = func(ctx context.Context, userID int32, tag string) (bool, error) { return true, nil }
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
			SiteAdmin: userID == 1,
		}, nil
	}
	database.Mocks.Users.GetByCurrentAuthUser = func(ctx context.Context) (*types.User, error) {
		return &types.User{ID: 1, SiteAdmin: true}, nil
	}
	cf = httpcli.NewFactory(
		nil,
		func(c *http.Client) error {
			c.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
				buf := &bytes.Buffer{}
				enc := json.NewEncoder(buf)
				switch r.URL.Path {
				case "/api/graphql": //github
					enc.Encode(repoResponse{
						Data: data{
							Viewer: viewer{
								Repositories: repositories{
									Nodes: []githubRepository{
										{
											NameWithOwner: "test-user/test",
											IsPrivate:     false,
										},
									},
								},
							},
						},
					})
				case "/api/v4/projects": //gitlab
					enc.Encode([]gitlabRepository{
						{
							PathWithNamespace: "test-user2/test2",
							Visibility:        "public",
						},
					})
				default:
					t.Fatalf("unexpected path: %s", r.URL.Path)
				}
				return &http.Response{
					Body:       ioutil.NopCloser(buf),
					StatusCode: http.StatusOK,
				}, nil
			})
			return nil
		},
	)

	ctx := context.Background()
	ctx = actor.WithActor(ctx, &actor.Actor{
		UID: 1,
	})

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Context: ctx,
			Schema:  mustParseGraphQLSchema(t),
			Query: `
			{
				affiliatedRepositories(
					user: "VXNlcjox"
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
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

// copied from the github client, just the fields we need
type githubRepository struct {
	NameWithOwner string
	IsPrivate     bool
}

type repoResponse struct {
	Data data `json:"data"`
}
type data struct {
	Viewer viewer `json:"viewer"`
}

type viewer struct {
	Repositories repositories `json:"repositories"`
}

type repositories struct {
	Nodes []githubRepository `json:"nodes"`
}

type gitlabRepository struct {
	Visibility        string `json:"visibility"`
	ID                int    `json:"id"`
	PathWithNamespace string `json:"path_with_namespace"`
}
