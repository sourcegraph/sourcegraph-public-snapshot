package app_test

import (
	"reflect"
	"testing"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/app/internal/apptest"
	"src.sourcegraph.com/sourcegraph/app/router"
)

func TestSearchForm(t *testing.T) {
	c, _ := apptest.New()

	if _, err := c.GetOK(router.Rel.URLTo(router.SearchForm).String()); err != nil {
		t.Fatal(err)
	}
}

func TestSearchResults(t *testing.T) {
	c, mock := apptest.New()

	mock.Search.Search_ = func(ctx context.Context, opt *sourcegraph.SearchOptions) (*sourcegraph.SearchResults, error) {
		want := &sourcegraph.SearchOptions{
			Repos:       true,
			People:      true,
			Defs:        true,
			Tree:        true,
			Query:       "myquery",
			ListOptions: sourcegraph.ListOptions{PerPage: 15},
		}
		if !reflect.DeepEqual(opt, want) {
			t.Errorf("got Search query %+v, want %+v", opt, want)
		}
		return &sourcegraph.SearchResults{}, nil
	}

	if _, err := c.GetOK(router.Rel.URLToSearch("myquery").String()); err != nil {
		t.Fatal(err)
	}
}

func TestRepoSearch_defaultBranch(t *testing.T) {
	c, mock := apptest.New()

	var calledSearch bool
	calledReposGet := mockRepoGet(mock, "my/repo")
	calledReposGetCommit := mock.Repos.MockGetCommit_ByID_NoCheck(t, "c")
	mock.Builds.GetRepoBuildInfo_ = func(ctx context.Context, op *sourcegraph.BuildsGetRepoBuildInfoOp) (*sourcegraph.RepoBuildInfo, error) {
		return &sourcegraph.RepoBuildInfo{
			LastSuccessful: &sourcegraph.Build{CommitID: "c2"},
		}, nil
	}
	mockEnabledRepoConfig(mock)
	mock.Search.Search_ = func(ctx context.Context, opt *sourcegraph.SearchOptions) (*sourcegraph.SearchResults, error) {
		calledSearch = true
		want := &sourcegraph.SearchOptions{
			Defs:        true,
			Tree:        true,
			Query:       "my/repo myquery",
			ListOptions: opt.ListOptions,
		}
		if !reflect.DeepEqual(opt, want) {
			t.Errorf("got Search query %+v, want %+v", opt, want)
		}
		return &sourcegraph.SearchResults{}, nil
	}

	u, err := router.Rel.URLToRepoSearch("my/repo", "", "myquery")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.GetOK(u.String()); err != nil {
		t.Fatal(err)
	}

	if !*calledReposGet {
		t.Error("!calledReposGet")
	}
	if !*calledReposGetCommit {
		t.Error("!calledReposGetCommit")
	}
	if !calledSearch {
		t.Error("!calledSearch")
	}
}

func TestRepoSearch_specificCommit(t *testing.T) {
	c, mock := apptest.New()

	var calledSearch bool
	calledReposGet := mockRepoGet(mock, "my/repo")
	calledReposGetCommit := mock.Repos.MockGetCommit_ByID_NoCheck(t, "c")
	mock.Builds.GetRepoBuildInfo_ = func(ctx context.Context, op *sourcegraph.BuildsGetRepoBuildInfoOp) (*sourcegraph.RepoBuildInfo, error) {
		return &sourcegraph.RepoBuildInfo{
			LastSuccessful: &sourcegraph.Build{CommitID: "c2"},
		}, nil
	}
	mockEnabledRepoConfig(mock)
	mock.Search.Search_ = func(ctx context.Context, opt *sourcegraph.SearchOptions) (*sourcegraph.SearchResults, error) {
		calledSearch = true
		want := &sourcegraph.SearchOptions{
			Defs: true,
			Tree: true,
			// TODO(sqs): Be sure that if the search results are
			// coming from an older revision than the one explicitly
			// specified in the URL, an alert is displayed to the
			// user.

			// TODO(sqs): should use last build (c2) not c, but I hot-fixed this to fix another issue.
			Query:       "my/repo :c myquery",
			ListOptions: opt.ListOptions,
		}
		if !reflect.DeepEqual(opt, want) {
			t.Errorf("got Search query %+v, want %+v", opt, want)
		}
		return &sourcegraph.SearchResults{}, nil
	}

	u, err := router.Rel.URLToRepoSearch("my/repo", "c", "myquery")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.GetOK(u.String()); err != nil {
		t.Fatal(err)
	}

	if !*calledReposGet {
		t.Error("!calledReposGet")
	}
	if !*calledReposGetCommit {
		t.Error("!calledReposGetCommit")
	}
	if !calledSearch {
		t.Error("!calledSearch")
	}
}
