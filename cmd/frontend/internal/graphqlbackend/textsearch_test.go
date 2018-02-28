package graphqlbackend

import (
	"context"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/google/zoekt/query"
	"github.com/pkg/errors"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/errcode"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitserver"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/searchquery"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

func TestQueryToZoektQuery(t *testing.T) {
	sPtr := func(s string) *string {
		return &s
	}
	cases := []struct {
		Name    string
		Pattern *patternInfo
		Query   string
	}{
		{
			Name: "substr",
			Pattern: &patternInfo{
				IsRegExp:                     true,
				IsCaseSensitive:              false,
				Pattern:                      "foo",
				IncludePatterns:              nil,
				ExcludePattern:               nil,
				PathPatternsAreRegExps:       true,
				PathPatternsAreCaseSensitive: false,
			},
			Query: "foo case:no",
		},
		{
			Name: "regex",
			Pattern: &patternInfo{
				IsRegExp:                     true,
				IsCaseSensitive:              false,
				Pattern:                      "(foo).*?(bar)",
				IncludePatterns:              nil,
				ExcludePattern:               nil,
				PathPatternsAreRegExps:       true,
				PathPatternsAreCaseSensitive: false,
			},
			Query: "(foo).*?(bar) case:no",
		},
		{
			Name: "path",
			Pattern: &patternInfo{
				IsRegExp:                     true,
				IsCaseSensitive:              false,
				Pattern:                      "foo",
				IncludePatterns:              []string{`\.go$`, `\.yaml$`},
				ExcludePattern:               sPtr(`\bvendor\b`),
				PathPatternsAreRegExps:       true,
				PathPatternsAreCaseSensitive: false,
			},
			Query: `foo case:no f:\.go$ f:\.yaml$ -f:\bvendor\b`,
		},
		{
			Name: "case",
			Pattern: &patternInfo{
				IsRegExp:                     true,
				IsCaseSensitive:              true,
				Pattern:                      "foo",
				IncludePatterns:              []string{`\.go$`, `yaml`},
				ExcludePattern:               nil,
				PathPatternsAreRegExps:       true,
				PathPatternsAreCaseSensitive: true,
			},
			Query: `foo case:yes f:\.go$ f:yaml`,
		},
		{
			Name: "casepath",
			Pattern: &patternInfo{
				IsRegExp:                     true,
				IsCaseSensitive:              true,
				Pattern:                      "foo",
				IncludePatterns:              []string{`\.go$`, `\.yaml$`},
				ExcludePattern:               sPtr(`\bvendor\b`),
				PathPatternsAreRegExps:       true,
				PathPatternsAreCaseSensitive: true,
			},
			Query: `foo case:yes f:\.go$ f:\.yaml$ -f:\bvendor\b`,
		},
	}
	for _, tt := range cases {
		t.Run(tt.Name, func(t *testing.T) {
			q, err := query.Parse(tt.Query)
			if err != nil {
				t.Fatalf("failed to parse %q: %v", tt.Query, err)
			}
			got, err := queryToZoektQuery(tt.Pattern)
			if err != nil {
				t.Fatal("queryToZoektQuery failed:", err)
			}
			if !queryEqual(got, q) {
				t.Fatalf("mismatched queries\ngot  %s\nwant %s", got.String(), q.String())
			}
		})
	}
}

func queryEqual(a query.Q, b query.Q) bool {
	sortChildren := func(q query.Q) query.Q {
		switch s := q.(type) {
		case *query.And:
			sort.Slice(s.Children, func(i, j int) bool {
				return s.Children[i].String() < s.Children[j].String()
			})
		case *query.Or:
			sort.Slice(s.Children, func(i, j int) bool {
				return s.Children[i].String() < s.Children[j].String()
			})
		}
		return q
	}
	return query.Map(a, sortChildren).String() == query.Map(b, sortChildren).String()
}

func TestSearchFilesInRepos(t *testing.T) {
	mockSearchFilesInRepo = func(ctx context.Context, repo *types.Repo, gitserverRepo gitserver.Repo, rev string, info *patternInfo, fetchTimeout time.Duration) (matches []*fileMatchResolver, limitHit bool, err error) {
		repoName := repo.URI
		switch repoName {
		case "foo/one":
			return []*fileMatchResolver{
				{
					uri: "git://" + string(repoName) + "?" + rev + "#" + "main.go",
				},
			}, false, nil
		case "foo/two":
			return []*fileMatchResolver{
				{
					uri: "git://" + string(repoName) + "?" + rev + "#" + "main.go",
				},
			}, false, nil
		case "foo/empty":
			return nil, false, nil
		case "foo/cloning":
			return nil, false, vcs.RepoNotExistError{CloneInProgress: true}
		case "foo/missing":
			return nil, false, vcs.RepoNotExistError{}
		case "foo/missing-db":
			return nil, false, &errcode.Mock{Message: "repo not found: foo/missing-db", IsNotFound: true}
		case "foo/timedout":
			return nil, false, context.DeadlineExceeded
		case "foo/no-rev":
			return nil, false, vcs.ErrRevisionNotFound
		default:
			return nil, false, errors.New("Unexpected repo")
		}
	}
	defer func() { mockSearchFilesInRepo = nil }()

	args := &repoSearchArgs{
		query: &patternInfo{
			FileMatchLimit: defaultMaxSearchResults,
			Pattern:        "foo",
		},
		repos: makeRepositoryRevisions("foo/one", "foo/two", "foo/empty", "foo/cloning", "foo/missing", "foo/missing-db", "foo/timedout", "foo/no-rev"),
	}
	query, err := searchquery.ParseAndCheck("foo")
	if err != nil {
		t.Fatal(err)
	}
	results, common, err := searchFilesInRepos(context.Background(), args, *query)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Errorf("expected two results, got %d", len(results))
	}
	if !reflect.DeepEqual(common.cloning, []api.RepoURI{"foo/cloning"}) {
		t.Errorf("unexpected cloning: %v", common.cloning)
	}
	sort.Slice(common.missing, func(i, j int) bool { return common.missing[i] < common.missing[j] }) // to make deterministic
	if !reflect.DeepEqual(common.missing, []api.RepoURI{"foo/missing", "foo/missing-db"}) {
		t.Errorf("unexpected missing: %v", common.missing)
	}
	if !reflect.DeepEqual(common.timedout, []api.RepoURI{"foo/timedout"}) {
		t.Errorf("unexpected timedout: %v", common.timedout)
	}

	// If we specify a rev and it isn't found, we fail the whole search since
	// that should be checked earlier.
	args = &repoSearchArgs{
		query: &patternInfo{
			FileMatchLimit: defaultMaxSearchResults,
			Pattern:        "foo",
		},
		repos: makeRepositoryRevisions("foo/no-rev@dev"),
	}
	_, _, err = searchFilesInRepos(context.Background(), args, *query)
	if errors.Cause(err) != vcs.ErrRevisionNotFound {
		t.Fatalf("searching non-existent rev expected to fail with %v got: %v", vcs.ErrRevisionNotFound, err)
	}
}

func makeRepositoryRevisions(repos ...string) []*repositoryRevisions {
	r := make([]*repositoryRevisions, len(repos))
	for i, urispec := range repos {
		uri, revs := parseRepositoryRevisions(urispec)
		r[i] = &repositoryRevisions{repo: &types.Repo{URI: uri}, revs: revs}
	}
	return r
}
