package app_test

import (
	"testing"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/app/internal/apptest"
	"src.sourcegraph.com/sourcegraph/app/router"
)

func TestReposByOwner(t *testing.T) {
	c, mock := apptest.New()

	user := &sourcegraph.User{Login: "u"}
	repos := &sourcegraph.RepoList{Repos: []*sourcegraph.Repo{{URI: "r1"}, {URI: "r2"}}}

	var calledReposList bool
	calledUsersGet := mock.Users.MockGet_Return(t, user)
	mock.Repos.List_ = func(ctx context.Context, opt *sourcegraph.RepoListOptions) (*sourcegraph.RepoList, error) {
		calledReposList = true
		return repos, nil
	}

	resp, err := c.GetOK(router.Rel.URLToUserSubroute(router.User, "u").String())
	if err != nil {
		t.Fatal(err)
	}

	dom, err := parseHTML(resp)
	if err != nil {
		t.Fatal(err)
	}
	personOwnedRepos := dom.Find(".person-repos li")
	if want := len(repos.Repos); personOwnedRepos.Size() != want {
		t.Errorf(".person-repos: got %d, want %d", personOwnedRepos.Size(), want)
	}

	if !*calledUsersGet {
		t.Errorf("!calledUsersGet")
	}
	if !calledReposList {
		t.Errorf("!calledReposList")
	}
}
