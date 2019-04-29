package graphqlbackend

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
)

func TestSearch(t *testing.T) {
	type Results struct {
		Results     []interface{}
		ResultCount int
	}
	tcs := []struct {
		name          string
		searchQuery   string
		reposListMock func(v0 context.Context, v1 db.ReposListOptions) ([]*types.Repo, error)
		wantResults   Results
	}{
		{
			name:        "empty query against no repos gets no results",
			searchQuery: "",
			reposListMock: func(v0 context.Context, v1 db.ReposListOptions) ([]*types.Repo, error) {
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
			db.Mocks.Repos.List = tc.reposListMock
			result := GraphQLSchema.Exec(context.Background(), testSearchGQLQuery, "", vars)
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

// serveLocalGitRepos serves git repositories in the given dir that have been set up according to
// https://theartofmachinery.com/2016/07/02/git_over_http.html.
func serveLocalGitRepos(dir string) error {
	s := http.Server{
		Addr: ":8080",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// fixme: serve directory
		}),
	}
	return s.ListenAndServe()
}
