package graphqlbackend

import (
	"context"
	"testing"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

func FuzzTestSearch(f *testing.F) {
	type Results struct {
		Results    []any
		MatchCount int
	}
	tcs := []struct {
		name                         string
		searchQuery                  string
		searchVersion                string
		reposListMock                func(v0 context.Context, v1 database.ReposListOptions) ([]*types.Repo, error)
		repoRevsMock                 func(_ context.Context, _ api.RepoName, spec string, _ gitserver.ResolveRevisionOptions) (api.CommitID, error)
		externalServicesListMock     func(_ context.Context, opt database.ExternalServicesListOptions) ([]*types.ExternalService, error)
		phabricatorGetRepoByNameMock func(_ context.Context, repo api.RepoName) (*types.PhabricatorRepo, error)
		wantResults                  Results
	}{
		{
			name:        "empty query against no repos gets no results",
			searchQuery: "",
			reposListMock: func(v0 context.Context, v1 database.ReposListOptions) ([]*types.Repo, error) {
				return nil, nil
			},
			repoRevsMock: func(_ context.Context, _ api.RepoName, spec string, _ gitserver.ResolveRevisionOptions) (api.CommitID, error) {
				return "", nil
			},
			externalServicesListMock: func(_ context.Context, opt database.ExternalServicesListOptions) ([]*types.ExternalService, error) {
				return nil, nil
			},
			phabricatorGetRepoByNameMock: func(_ context.Context, repo api.RepoName) (*types.PhabricatorRepo, error) {
				return nil, nil
			},
			wantResults: Results{
				Results:    nil,
				MatchCount: 0,
			},
			searchVersion: "V1",
		},
		{
			name:        "empty query against empty repo gets no results",
			searchQuery: "",
			reposListMock: func(v0 context.Context, v1 database.ReposListOptions) ([]*types.Repo, error) {
				return []*types.Repo{{Name: "test"}},

					nil
			},
			repoRevsMock: func(_ context.Context, _ api.RepoName, spec string, _ gitserver.ResolveRevisionOptions) (api.CommitID, error) {
				return "", nil
			},
			externalServicesListMock: func(_ context.Context, opt database.ExternalServicesListOptions) ([]*types.ExternalService, error) {
				return nil, nil
			},
			phabricatorGetRepoByNameMock: func(_ context.Context, repo api.RepoName) (*types.PhabricatorRepo, error) {
				return nil, nil
			},
			wantResults: Results{
				Results:    nil,
				MatchCount: 0,
			},
			searchVersion: "V1",
		},
	}
	ftc := tcs[0]
	f.Add(ftestSearchGQLQuery)

	f.Fuzz(func(t *testing.T, tcq string) {
		conf.Mock(&conf.Unified{})
		defer conf.Mock(nil)
		vars := map[string]any{"query": ftc.searchQuery, "version": ftc.searchVersion}

		MockDecodedViewerFinalSettings = &schema.Settings{}
		defer func() { MockDecodedViewerFinalSettings = nil }()

		repos := database.NewMockRepoStore()
		repos.ListFunc.SetDefaultHook(ftc.reposListMock)

		ext := database.NewMockExternalServiceStore()
		ext.ListFunc.SetDefaultHook(ftc.externalServicesListMock)

		phabricator := database.NewMockPhabricatorStore()
		phabricator.GetByNameFunc.SetDefaultHook(ftc.phabricatorGetRepoByNameMock)

		db := database.NewMockDB()
		db.ReposFunc.SetDefaultReturn(repos)
		db.ExternalServicesFunc.SetDefaultReturn(ext)
		db.PhabricatorFunc.SetDefaultReturn(phabricator)

		gsClient := gitserver.NewMockClient()
		gsClient.ResolveRevisionFunc.SetDefaultHook(ftc.repoRevsMock)

		sr := newSchemaResolver(db, gsClient)
		schema, err := graphql.ParseSchema(mainSchema, sr, graphql.Tracer(&requestTracer{}))
		if err != nil {
			t.Fatal(err)
		}

		result := schema.Exec(context.Background(), tcq, "", vars)
		if len(result.Errors) > 0 {
			t.Log(err)
		}
	})
}

var ftestSearchGQLQuery = `
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
				serviceKind
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
					matchCount
					elapsedMilliseconds
				}
			}
		}
`
