package httpapi

import (
	"reflect"
	"testing"

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
