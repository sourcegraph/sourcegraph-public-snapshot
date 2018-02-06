package graphqlbackend

import (
	"context"
	"reflect"
	"sort"
	"testing"

	"github.com/pkg/errors"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/errcode"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitserver"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/searchquery"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

func TestSearchRepos(t *testing.T) {
	mockSearchRepo = func(ctx context.Context, repo *types.Repo, gitserverRepo gitserver.Repo, rev string, info *patternInfo) (matches []*fileMatch, limitHit bool, err error) {
		repoName := repo.URI
		switch repoName {
		case "foo/one":
			return []*fileMatch{
				{
					uri: "git://" + string(repoName) + "?" + rev + "#" + "main.go",
				},
			}, false, nil
		case "foo/two":
			return []*fileMatch{
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
	defer func() { mockSearchRepo = nil }()

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
	results, common, err := searchRepos(context.Background(), args, *query)
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
	_, _, err = searchRepos(context.Background(), args, *query)
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
