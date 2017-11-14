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
		Query: &patternInfo{
			FileMatchLimit: 300,
			Pattern:        "foo",
		},
		Repositories: []*repositoryRevision{{Repo: "foo/one"}, {Repo: "foo/two"}, {Repo: "foo/empty"}, {Repo: "foo/cloning"}, {Repo: "foo/missing"}},
	}
	results, err := (&rootResolver{}).SearchRepos(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	if len(results.Results()) != 2 {
		t.Errorf("expected two results, got %d", len(results.Results()))
	}
	if !reflect.DeepEqual(results.Cloning(), []string{"foo/cloning"}) {
		t.Errorf("unexpected missing: %v", results.Cloning())
	}
	if !reflect.DeepEqual(results.Missing(), []string{"foo/missing"}) {
		t.Errorf("unexpected missing: %v", results.Missing())
	}
}
