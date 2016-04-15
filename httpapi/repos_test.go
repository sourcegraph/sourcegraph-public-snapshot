package httpapi

import (
	"net/http"
	"reflect"
	"testing"
	"time"

	authpkg "sourcegraph.com/sourcegraph/sourcegraph/auth"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sqs/pbtypes"
)

func TestRepo(t *testing.T) {
	c, mock := newTest()

	wantRepo := &sourcegraph.Repo{URI: "r/r"}

	calledGet := mock.Repos.MockGet(t, "r/r")

	var repo *sourcegraph.Repo
	if err := c.GetJSON("/repos/r/r", &repo); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(repo, wantRepo) {
		t.Errorf("got %+v, want %+v", repo, wantRepo)
	}
	if !*calledGet {
		t.Error("!calledGet")
	}
}

func TestRepoResolve(t *testing.T) {
	c, mock := newTest()

	want := &sourcegraph.RepoResolution{
		Result: &sourcegraph.RepoResolution_Repo{
			Repo: &sourcegraph.RepoSpec{URI: "r"},
		},
	}

	calledResolve := mock.Repos.MockResolve_Local(t, "r")

	var res *sourcegraph.RepoResolution
	if err := c.GetJSON("/repos/r/-/resolve", &res); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(res, want) {
		t.Errorf("got %+v, want %+v", res, want)
	}
	if !*calledResolve {
		t.Error("!calledResolve")
	}
}

// Test that if the repo hasn't been modified since the client's
// If-Modified-Since, HTTP 304 Not Modified is returned.
func TestRepo_caching_notModified(t *testing.T) {
	c, mock := newTest()

	mtime := time.Now().UTC()
	ts := pbtypes.NewTimestamp(mtime)

	calledGet := mock.Repos.MockGet_Return(t, &sourcegraph.Repo{
		URI:       "r/r",
		UpdatedAt: &ts,
	})

	req, _ := http.NewRequest("GET", "/repos/r/r", nil)
	req.Header.Set("if-modified-since", mtime.Add(2*time.Second).Format(http.TimeFormat))

	resp, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if want := http.StatusNotModified; resp.StatusCode != want {
		t.Errorf("got HTTP status %d, want %d", resp.StatusCode, want)
	}
	if !*calledGet {
		t.Error("!calledGet")
	}
}

// Test that if the repo was modified after the client's
// If-Modified-Since, it is returned.
func TestRepo_caching_modifiedSince(t *testing.T) {
	c, mock := newTest()

	mtime := time.Now().UTC()
	ts := pbtypes.NewTimestamp(mtime)

	calledGet := mock.Repos.MockGet_Return(t, &sourcegraph.Repo{
		URI:       "r/r",
		UpdatedAt: &ts,
	})

	req, _ := http.NewRequest("GET", "/repos/r/r", nil)
	req.Header.Set("if-modified-since", mtime.Add(-2*time.Second).Format(http.TimeFormat))

	resp, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if want := http.StatusOK; resp.StatusCode != want {
		t.Errorf("got HTTP status %d, want %d", resp.StatusCode, want)
	}
	if !*calledGet {
		t.Error("!calledGet")
	}
}

func TestRepos_admin(t *testing.T) {
	c, mock := newTest()
	mock.Ctx = authpkg.WithActor(mock.Ctx, authpkg.Actor{UID: 1, Login: "test", Admin: true})

	wantRepos := &sourcegraph.RepoList{
		Repos: []*sourcegraph.Repo{{URI: "r/r"}},
	}

	calledList := mock.Repos.MockList(t, "r/r")

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

func TestRepos_nonadmin(t *testing.T) {
	c, mock := newTest()

	wantRepos := &sourcegraph.RepoList{}

	calledList := mock.Repos.MockList(t)

	var repos *sourcegraph.RepoList
	if err := c.GetJSON("/repos", &repos); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(repos, wantRepos) {
		t.Errorf("got %+v, want %+v", repos, wantRepos)
	}
	if *calledList {
		t.Error("calledList")
	}
}

func TestRepoCreate(t *testing.T) {
	c, mock := newTest()

	want := &sourcegraph.Repo{URI: "r"}

	var calledCreate bool
	mock.Repos.Create_ = func(ctx context.Context, op *sourcegraph.ReposCreateOp) (*sourcegraph.Repo, error) {
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
