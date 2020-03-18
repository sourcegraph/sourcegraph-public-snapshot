package graphqlbackend

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

func TestSearch(t *testing.T) {
	type Results struct {
		Results     []interface{}
		ResultCount int
	}
	tcs := []struct {
		name                         string
		searchQuery                  string
		searchVersion                string
		reposListMock                func(v0 context.Context, v1 db.ReposListOptions) ([]*types.Repo, error)
		repoRevsMock                 func(spec string, opt *git.ResolveRevisionOptions) (api.CommitID, error)
		externalServicesListMock     func(opt db.ExternalServicesListOptions) ([]*types.ExternalService, error)
		phabricatorGetRepoByNameMock func(repo api.RepoName) (*types.PhabricatorRepo, error)
		wantResults                  Results
	}{
		{
			name:        "empty query against no repos gets no results",
			searchQuery: "",
			reposListMock: func(v0 context.Context, v1 db.ReposListOptions) ([]*types.Repo, error) {
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
			searchVersion: "V1",
		},
		{
			name:        "empty query against empty repo gets no results",
			searchQuery: "",
			reposListMock: func(v0 context.Context, v1 db.ReposListOptions) ([]*types.Repo, error) {
				return []*types.Repo{{Name: "test"}},

					nil
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
			searchVersion: "V1",
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			conf.Mock(&conf.Unified{})
			defer conf.Mock(nil)
			vars := map[string]interface{}{"query": tc.searchQuery, "version": tc.searchVersion}
			db.Mocks.Repos.List = tc.reposListMock
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
					resultCount
					elapsedMilliseconds
				}
			}
		}
`

func testStringResult(result *searchSuggestionResolver) string {
	var name string
	switch r := result.result.(type) {
	case *RepositoryResolver:
		name = "repo:" + string(r.repo.Name)
	case *GitTreeEntryResolver:
		name = "file:" + r.Path()
	case *languageResolver:
		name = "lang:" + r.name
	default:
		panic("never here")
	}
	if result.score == 0 {
		return "<removed>"
	}
	return name
}

func Test_defaultRepositories(t *testing.T) {
	tcs := []struct {
		name             string
		defaultsInDb     []string
		indexedRepoNames map[string]bool
		want             []string
	}{
		{
			name:             "none in db => none returned",
			defaultsInDb:     nil,
			indexedRepoNames: nil,
			want:             nil,
		},
		{
			name:             "two in db, one indexed => indexed repo returned",
			defaultsInDb:     []string{"unindexedrepo", "indexedrepo"},
			indexedRepoNames: map[string]bool{"indexedrepo": true},
			want:             []string{"indexedrepo"},
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			var drs []*types.Repo
			for i, name := range tc.defaultsInDb {
				r := &types.Repo{
					ID:   api.RepoID(i),
					Name: api.RepoName(name),
				}
				drs = append(drs, r)
			}
			getRawDefaultRepos := func(ctx context.Context) ([]*types.Repo, error) {
				return drs, nil
			}
			indexedRepos := func(ctx context.Context, revs []*search.RepositoryRevisions) (indexed, unindexed []*search.RepositoryRevisions, err error) {
				for _, r := range drs {
					r2 := &search.RepositoryRevisions{
						Repo: r,
					}
					if tc.indexedRepoNames[string(r.Name)] {
						indexed = append(indexed, r2)
					} else {
						unindexed = append(unindexed, r2)
					}
				}
				return indexed, unindexed, nil
			}
			ctx := context.Background()
			drs, err := defaultRepositories(ctx, getRawDefaultRepos, indexedRepos)
			if err != nil {
				t.Fatal(err)
			}
			var drNames []string
			for _, dr := range drs {
				drNames = append(drNames, string(dr.Name))
			}
			if !reflect.DeepEqual(drNames, tc.want) {
				t.Errorf("names of default repos = %v, want %v", drNames, tc.want)
			}
		})
	}
}

func Test_detectSearchType(t *testing.T) {
	typeRegexp := "regexp"
	typeLiteral := "literal"
	testCases := []struct {
		name        string
		version     string
		patternType *string
		input       string
		want        query.SearchType
	}{
		{"V1, no pattern type", "V1", nil, "", query.SearchTypeRegex},
		{"V2, no pattern type", "V2", nil, "", query.SearchTypeLiteral},
		{"V2, no pattern type, input does not produce parse error", "V2", nil, "/-/godoc", query.SearchTypeLiteral},
		{"V1, regexp pattern type", "V1", &typeRegexp, "", query.SearchTypeRegex},
		{"V2, regexp pattern type", "V2", &typeRegexp, "", query.SearchTypeRegex},
		{"V1, literal pattern type", "V1", &typeLiteral, "", query.SearchTypeLiteral},
		{"V2, override regexp pattern type", "V2", &typeLiteral, "patterntype:regexp", query.SearchTypeRegex},
		{"V2, override regex variant pattern type", "V2", &typeLiteral, "patterntype:regex", query.SearchTypeRegex},
		{"V2, override regex variant pattern type with double quotes", "V2", &typeLiteral, `patterntype:"regex"`, query.SearchTypeRegex},
		{"V2, override regex variant pattern type with single quotes", "V2", &typeLiteral, `patterntype:'regex'`, query.SearchTypeRegex},
		{"V1, override literal pattern type", "V1", &typeRegexp, "patterntype:literal", query.SearchTypeLiteral},
		{"V1, override literal pattern type, with case-insensitive query", "V1", &typeRegexp, "pAtTErNTypE:literal", query.SearchTypeLiteral},
	}

	for _, test := range testCases {
		t.Run(test.name, func(*testing.T) {
			got, err := detectSearchType(test.version, test.patternType, test.input)
			if err != nil {
				t.Fatal(err)
			}
			if got != test.want {
				t.Errorf("failed %v, got %v, expected %v", test.name, got, test.want)
			}
		})
	}
}

func Test_exactlyOneRepo(t *testing.T) {
	cases := []struct {
		repoFilters []string
		want        bool
	}{
		{
			repoFilters: []string{`^github\.com/sourcegraph/zoekt$`},
			want:        true,
		},
		{
			repoFilters: []string{`^.*$`},
			want:        false,
		},

		{
			repoFilters: []string{`^github\.com/sourcegraph/zoekt`},
			want:        false,
		},
		{
			repoFilters: []string{`^github\.com/sourcegraph/zoekt$`, `github\.com/sourcegraph/sourcegraph`},
			want:        false,
		},
	}
	for _, c := range cases {
		t.Run("exactly one repo", func(t *testing.T) {
			if got := exactlyOneRepo(c.repoFilters); got != c.want {
				t.Errorf("got %t, want %t", got, c.want)
			}
		})
	}
}

func Test_QuoteSuggestions(t *testing.T) {
	t.Run("regex error", func(t *testing.T) {
		raw := "*"
		_, _, err := query.Process(raw, query.SearchTypeRegex)
		if err == nil {
			t.Fatalf("error returned from query.Process(%q) is nil", raw)
		}
		alert := alertForQuery(raw, err)
		if !strings.Contains(strings.ToLower(alert.title), "regexp") {
			t.Errorf("title is '%s', want it to contain 'regexp'", alert.title)
		}
		if !strings.Contains(alert.description, "regular expression") {
			t.Errorf("description is '%s', want it to contain 'regular expression'", alert.description)
		}
	})

	t.Run("type error that is not a regex error should show a suggestion", func(t *testing.T) {
		raw := "-foobar"
		_, _, alert := query.Process(raw, query.SearchTypeRegex)
		if alert == nil {
			t.Fatalf("alert returned from query.Process(%q) is nil", raw)
		}
	})

	t.Run("query parse error", func(t *testing.T) {
		raw := ":"
		_, _, err := query.Process(raw, query.SearchTypeRegex)
		if err == nil {
			t.Fatalf("error returned from query.Process(%q) is nil", raw)
		}
		alert := alertForQuery(raw, err)
		if strings.Contains(strings.ToLower(alert.title), "regexp") {
			t.Errorf("title is '%s', want it not to contain 'regexp'", alert.title)
		}
		if strings.Contains(alert.description, "regular expression") {
			t.Errorf("description is '%s', want it not to contain 'regular expression'", alert.description)
		}
	})

	t.Run("negated file field with an invalid regex", func(t *testing.T) {
		raw := "-f:(a"
		_, _, err := query.Process(raw, query.SearchTypeRegex)
		if err == nil {
			t.Fatal("query.Process failed to detect the invalid regex in the f: field")
		}
		alert := alertForQuery(raw, err)
		if len(alert.proposedQueries) != 1 {
			t.Fatalf("got %d proposed queries (%v), want exactly 1", len(alert.proposedQueries), alert.proposedQueries)
		}
	})
}
