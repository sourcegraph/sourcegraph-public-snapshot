package graphqlbackend

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

// TestSearch_authzPostFilter tests the validity of search results when authzPostFilter is true.
//
// 🚨 SECURITY: this test ensure the correctness of permissions with authzPostFilter is true.
func TestSearch_authzPostFilter(t *testing.T) {
	db.Mocks.ExternalServices.List = func(opt db.ExternalServicesListOptions) ([]*types.ExternalService, error) {
		return nil, nil
	}
	db.Mocks.Phabricator.GetByName = func(repo api.RepoName) (*types.PhabricatorRepo, error) {
		return nil, nil
	}

	tcs := []struct {
		name string

		// Mock doResultsAttempt, which fetches raw, unauthz'd results
		doResultsAttemptMock func(ctx context.Context, forceOnlyResultType string, cancel func()) (res *SearchResultsResolver, err error)

		// Mock Repos.GetByName function, which is used to verify permissions
		reposGetByNameMock func(ctx context.Context, name api.RepoName) (*types.Repo, error)

		wantResults            Results
		wantResultStrings      []string
		forbiddenResultStrings []string
	}{
		{
			name: "Filter out unauthorized repositories",
			doResultsAttemptMock: func(ctx context.Context, forceOnlyResultType string, cancel func()) (res *SearchResultsResolver, err error) {
				allRepos := []*types.Repo{
					{ID: 1, Name: "repo1"},
					{ID: 2, Name: "repo2"},
					{ID: 3, Name: "repo3"},
					{ID: 4, Name: "repo4"},
					{ID: 5, Name: "repo5"},
					{ID: 6, Name: "repo6"},
				}
				allRepoNames := map[api.RepoName]struct{}{}
				allRepoResolvers := make([]SearchResultResolver, len(allRepos))
				for i, r := range allRepos {
					allRepoNames[r.Name] = struct{}{}
					allRepoResolvers[i] = NewRepositoryResolver(r)
				}
				return &SearchResultsResolver{
					SearchResults: []SearchResultResolver{
						NewRepositoryResolver(&types.Repo{Name: "repo1"}),
						NewRepositoryResolver(&types.Repo{Name: "repo2"}),
						&FileMatchResolver{Repo: &types.Repo{Name: "repo3"}},
						&FileMatchResolver{Repo: &types.Repo{Name: "repo4"}},
						&commitSearchResultResolver{label: "COMMIT_SEARCH_SENTINEL"},
						&codemodResultResolver{
							commit: &GitCommitResolver{repo: &RepositoryResolver{repo: &types.Repo{Name: "repo5"}}},
						},
						&codemodResultResolver{
							commit: &GitCommitResolver{repo: &RepositoryResolver{repo: &types.Repo{Name: "repo6"}}},
						},
					},
					searchResultsCommon: searchResultsCommon{
						maxResultsCount: 500,
						resultCount:     500,
						indexed:         allRepos,
						cloning:         allRepos,
						missing:         allRepos,
						timedout:        allRepos,
						partial:         allRepoNames,
					},
				}, nil
			},
			reposGetByNameMock: func(ctx context.Context, name api.RepoName) (*types.Repo, error) {
				switch name {
				case "repo1", "repo3", "repo5":
					return &types.Repo{Name: name}, nil
				default:
					return nil, errors.New("unauthorized")
				}
			},
			wantResultStrings:      []string{"repo1", "repo3", "repo5"},
			forbiddenResultStrings: []string{"repo2", "repo4", "COMMIT_SEARCH_SENTINEL", "repo6"},
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			conf.Mock(&conf.Unified{
				SiteConfiguration: schema.SiteConfiguration{
					ExperimentalFeatures: &schema.ExperimentalFeatures{
						AuthzPostSearchFilter: true,
					},
				},
			})
			defer conf.Mock(nil)
			vars := map[string]interface{}{"query": "DOES NOT MATTER", "version": "V1"}
			db.Mocks.Repos.GetByName = tc.reposGetByNameMock
			defer func() { db.Mocks.Repos.GetByName = nil }()
			mockDoResultsAttempt = tc.doResultsAttemptMock
			defer func() { mockDoResultsAttempt = nil }()

			sr := &schemaResolver{}
			schema, err := graphql.ParseSchema(Schema, sr, graphql.Tracer(prometheusTracer{}))
			if err != nil {
				t.Fatal(err)
			}
			ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 123})
			result := schema.Exec(ctx, testSearchAuthzGQLQuery, "", vars)
			if len(result.Errors) > 0 {
				t.Fatalf("graphQL query returned errors: %+v", result.Errors)
			}
			rawResults := string(result.Data)
			for _, forbiddenString := range tc.forbiddenResultStrings {
				if strings.Contains(rawResults, forbiddenString) {
					var r interface{}
					_ = json.Unmarshal([]byte(rawResults), &r)
					rb, _ := json.MarshalIndent(r, "", "  ")
					t.Errorf("results contain forbidden string %q, results: %s", forbiddenString, string(rb))
				}
			}
			for _, wantString := range tc.wantResultStrings {
				if !strings.Contains(rawResults, wantString) {
					var r interface{}
					_ = json.Unmarshal([]byte(rawResults), &r)
					rb, _ := json.MarshalIndent(r, "", "  ")
					t.Errorf("results do not contain wanted string %q, results: %s", wantString, string(rb))
				}
			}
		})
	}
}

type RepoResult struct {
	Name string `json:"name"`
}

func (r *RepoResult) UnmarshalJSON(data []byte) error {
	fields := map[string]interface{}{}
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	var ok bool
	r.Name, ok = fields["name"].(string)
	if !ok {
		return fmt.Errorf("\"name\" field %v was not of type bool, was type %T", fields["name"], fields["name"])
	}
	return nil
}

type Results struct {
	Results     *ResultsList
	ResultCount int `json:"resultCount"`
}

type ResultsList struct {
	Results []interface{}
}

func (r *ResultsList) UnmarshalJSON(data []byte) error {
	var entries []map[string]interface{}
	if err := json.Unmarshal(data, &entries); err != nil {
		return err
	}
	for _, entry := range entries {
		switch entry["__typename"] {
		case "Repository":
			marshaledEntry, err := json.Marshal(entry)
			if err != nil {
				return err
			}
			var repo RepoResult
			if err := json.Unmarshal(marshaledEntry, &repo); err != nil {
				return err
			}
			r.Results = append(r.Results, &repo)
		default:
			return fmt.Errorf("unrecognized __typename in entry %v", entry)
		}
	}

	return nil
}

var testSearchAuthzGQLQuery = `
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

		fragment CodemodFields on CodemodResult {
          commit {
            repository{
              name
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

		query ($query: String!, $version: SearchVersion!, $patternType: SearchPatternType) {
			site {
				buildVersion
			}
			search(query: $query, version: $version, patternType: $patternType) {
				results {
					results {
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
						... on CodemodResult {
							...CodemodFields
						}
					}
					limitHit
					repositories {
						name
					}
					repositoriesSearched {
						name
					}
					indexedRepositoriesSearched {
						name
					}
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
