package httpapi

import (
	"reflect"
	"testing"

	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend"
)

func TestRepos(t *testing.T) {
	c := newTest()

	wantRepos := &sourcegraph.RepoList{
		Repos: []*sourcegraph.Repo{{URI: "r/r"}},
	}

	calledList := backend.Mocks.Repos.MockList(t, "r/r")

	var repos *sourcegraph.RepoList
	if err := c.GetJSON("/repos", &repos); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(repos, wantRepos) {
		t.Errorf("got %+v, want %+v", repos, wantRepos)
	}
	if !*calledList {
		t.Error("!calledList")
	}
}

func TestRepoCreate(t *testing.T) {
	c := newTest()

	want := &sourcegraph.Repo{URI: "r"}

	var calledCreate bool
	backend.Mocks.Repos.Create = func(ctx context.Context, op *sourcegraph.ReposCreateOp) (*sourcegraph.Repo, error) {
		if op.GetNew().URI != want.URI {
			t.Errorf("got URI %q, want %q", op.GetNew().URI, want.URI)
		}
		calledCreate = true
		return want, nil
	}

	op := sourcegraph.ReposCreateOp{
		Op: &sourcegraph.ReposCreateOp_New{
			New: &sourcegraph.ReposCreateOp_NewRepo{
				URI: "r",
			},
		},
	}

	var repo *sourcegraph.Repo
	if err := c.DoJSON("POST", "/repos", &op, &repo); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(repo, want) {
		t.Errorf("got %+v, want %+v", repo, want)
	}
	if !calledCreate {
		t.Error("!calledCreate")
	}
}
