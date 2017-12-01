package graphqlbackend

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

func TestSearchRepos(t *testing.T) {
	mockSearchRepo = func(ctx context.Context, repoName, rev string, info *patternInfo) (matches []*fileMatch, limitHit bool, err error) {
		switch repoName {
		case "foo/one":
			return []*fileMatch{
				{
					uri: "git://" + repoName + "?" + rev + "#" + "main.go",
				},
			}, false, nil
		case "foo/two":
			return []*fileMatch{
				{
					uri: "git://" + repoName + "?" + rev + "#" + "main.go",
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
			FileMatchLimit: 300,
			Pattern:        "foo",
		},
		repos: []*repositoryRevision{{repo: "foo/one"}, {repo: "foo/two"}, {repo: "foo/empty"}, {repo: "foo/cloning"}, {repo: "foo/missing"}},
	}
	results, common, err := searchRepos(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Errorf("expected two results, got %d", len(results))
	}
	if !reflect.DeepEqual(common.cloning, []string{"foo/cloning"}) {
		t.Errorf("unexpected missing: %v", common.cloning)
	}
	if !reflect.DeepEqual(common.missing, []string{"foo/missing"}) {
		t.Errorf("unexpected missing: %v", common.missing)
	}
}
