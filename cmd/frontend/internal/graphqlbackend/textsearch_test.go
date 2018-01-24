package graphqlbackend

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/searchquery"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

func TestSearchRepos(t *testing.T) {
	mockSearchRepo = func(ctx context.Context, repo *types.Repo, rev string, info *patternInfo) (matches []*fileMatch, limitHit bool, err error) {
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
		repos: makeRepositoryRevisions("foo/one", "foo/two", "foo/empty", "foo/cloning", "foo/missing"),
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
		t.Errorf("unexpected missing: %v", common.cloning)
	}
	if !reflect.DeepEqual(common.missing, []api.RepoURI{"foo/missing"}) {
		t.Errorf("unexpected missing: %v", common.missing)
	}
}

func makeRepositoryRevisions(repos ...api.RepoURI) []*repositoryRevisions {
	r := make([]*repositoryRevisions, len(repos))
	for i, uri := range repos {
		r[i] = &repositoryRevisions{repo: &types.Repo{URI: uri}}
	}
	return r
}
