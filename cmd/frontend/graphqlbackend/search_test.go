package graphqlbackend

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
)

func TestSearch(t *testing.T) {
	type Results struct {
		Results     []interface{}
		ResultCount int
	}
	tcs := []struct {
		name                         string
		searchQuery                  string
		reposMinimalListMock         func(v0 context.Context, v1 db.ReposListOptions) ([]*db.MinimalRepo, error)
		repoRevsMock                 func(spec string, opt *git.ResolveRevisionOptions) (api.CommitID, error)
		externalServicesListMock     func(opt db.ExternalServicesListOptions) ([]*types.ExternalService, error)
		phabricatorGetRepoByNameMock func(repo api.RepoName) (*types.PhabricatorRepo, error)
		wantResults                  Results
	}{
		{
			name:        "empty query against no repos gets no results",
			searchQuery: "",
			reposMinimalListMock: func(v0 context.Context, v1 db.ReposListOptions) ([]*db.MinimalRepo, error) {
				return nil, nil
			},
			repoRevsMock: func(spec string, opt *git.ResolveRevisionOptions) (api.CommitID, error) {
				return api.CommitID(""), nil
			},
			externalServicesListMock: func(opt db.ExternalServicesListOptions) ([]*types.ExternalService, error) {
				return nil, nil
			},
			phabricatorGetRepoByNameMock: func(repo api.RepoName) (*types.PhabricatorRepo, error) {
				return nil, nil
			},
			wantResults: Results{
				Results:     nil,
				ResultCount: 0,
			},
		},
		{
			name:        "empty query against empty repo gets no results",
			searchQuery: "",
			reposMinimalListMock: func(v0 context.Context, v1 db.ReposListOptions) ([]*db.MinimalRepo, error) {
				return []*db.MinimalRepo{
					{
						Name: "test",
					},
				}, nil
			},
			repoRevsMock: func(spec string, opt *git.ResolveRevisionOptions) (api.CommitID, error) {
				return api.CommitID(""), nil
			},
			externalServicesListMock: func(opt db.ExternalServicesListOptions) ([]*types.ExternalService, error) {
				return nil, nil
			},
			phabricatorGetRepoByNameMock: func(repo api.RepoName) (*types.PhabricatorRepo, error) {
				return nil, nil
			},
			wantResults: Results{
				Results:     nil,
				ResultCount: 0,
			},
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			conf.Mock(&conf.Unified{})
			defer conf.Mock(nil)
			vars := map[string]interface{}{"query": tc.searchQuery}
			db.Mocks.Repos.MinimalList = tc.reposMinimalListMock
			sr := &schemaResolver{}
			schema, err := graphql.ParseSchema(Schema, sr, graphql.Tracer(prometheusTracer{}))
			if err != nil {
				t.Fatal(err)
			}
			db.Mocks.ExternalServices.List = tc.externalServicesListMock
			db.Mocks.Phabricator.GetByName = tc.phabricatorGetRepoByNameMock
			git.Mocks.ResolveRevision = tc.repoRevsMock
			result := schema.Exec(context.Background(), testSearchGQLQuery, "", vars)
			if len(result.Errors) > 0 {
				t.Fatalf("graphQL query returned errors: %+v", result.Errors)
			}
			var search struct {
				Results Results
			}
			if err := json.Unmarshal(result.Data, &search); err != nil {
				t.Fatalf("parsing JSON response: %v", err)
			}
			gotResults := search.Results
			if !reflect.DeepEqual(gotResults, tc.wantResults) {
				t.Fatalf("results = %+v, want %+v", gotResults, tc.wantResults)
			}
		})
	}
}

type NopRecentSearches struct{}

func (m *NopRecentSearches) Log(ctx context.Context, s string) error { return nil }
func (m *NopRecentSearches) Top(ctx context.Context, n int32) ([]string, []int32, error) {
	return nil, nil, nil
}
func (m *NopRecentSearches) List(ctx context.Context) ([]string, error)   { return nil, nil }
func (m *NopRecentSearches) Cleanup(ctx context.Context, limit int) error { return nil }

var testSearchGQLQuery = `
		fragment FileMatchFields on FileMatch {
			repository {
				name
				url
			}
			file {
				name
				path
				url
				commit {
					oid
				}
			}
			lineMatches {
				preview
				lineNumber
				offsetAndLengths
				limitHit
			}
		}

		fragment CommitSearchResultFields on CommitSearchResult {
			messagePreview {
				value
				highlights{
					line
					character
					length
				}
			}
			diffPreview {
				value
				highlights {
					line
					character
					length
				}
			}
			label {
				html
			}
			url
			matches {
				url
				body {
					html
					text
				}
				highlights {
					character
					line
					length
				}
			}
			commit {
				repository {
					name
				}
				oid
				url
				subject
				author {
					date
					person {
						displayName
					}
				}
			}
		}

		fragment RepositoryFields on Repository {
			name
			url
			externalURLs {
				serviceType
				url
			}
			label {
				html
			}
		}

		query ($query: String!) {
			site {
				buildVersion
			}
			search(query: $query) {
				results {
					results{
						__typename
						... on FileMatch {
						...FileMatchFields
					}
						... on CommitSearchResult {
						...CommitSearchResultFields
					}
						... on Repository {
						...RepositoryFields
					}
					}
					limitHit
					cloning {
						name
					}
					missing {
						name
					}
					timedout {
						name
					}
					resultCount
					elapsedMilliseconds
				}
			}
		}
`

func testStringResult(result *searchSuggestionResolver) string {
	var name string
	switch r := result.result.(type) {
	case *repositoryResolver:
		name = "repo:" + string(r.repo.Name)
	case *gitTreeEntryResolver:
		name = "file:" + r.path
	default:
		panic("never here")
	}
	if result.score == 0 {
		return "<removed>"
	}
	return name
}
