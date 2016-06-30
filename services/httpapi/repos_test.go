package httpapi

import (
	"net/http"
	"reflect"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sqs/pbtypes"
)

func TestRepo(t *testing.T) {
	c, mock := newTest()

	wantRepo := &sourcegraph.Repo{ID: 1}

	calledReposResolve := mock.Repos.MockResolve_Local(t, "r/r", 1)
	calledGet := mock.Repos.MockGet(t, 1)

	var repo *sourcegraph.Repo
	if err := c.GetJSON("/repos/r/r", &repo); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(repo, wantRepo) {
		t.Errorf("got %+v, want %+v", repo, wantRepo)
	}
	if !*calledReposResolve {
		t.Error("!calledReposResolve")
	}
	if !*calledGet {
		t.Error("!calledGet")
	}
}

func TestRepoResolve_IncludedRepo(t *testing.T) {
	c, mock := newTest()

	want := &repoResolution{
		Data:         sourcegraph.RepoResolution{Repo: 1, CanonicalPath: "r"},
		IncludedRepo: &sourcegraph.Repo{ID: 1},
	}

	calledResolve := mock.Repos.MockResolve_Local(t, "r", 1)
	calledGet := mock.Repos.MockGet(t, 1)

	var res *repoResolution
	if err := c.GetJSON("/repos/r/-/resolve", &res); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(res, want) {
		t.Errorf("got %+v, want %+v", res, want)
	}
	if !*calledResolve {
		t.Error("!calledResolve")
	}
	if !*calledGet {
		t.Error("!calledGet")
	}
}

func TestRepoResolve_IncludedRepo_ignoreErr(t *testing.T) {
	c, mock := newTest()

	want := &repoResolution{
		Data: sourcegraph.RepoResolution{Repo: 1, CanonicalPath: "r"},
	}

	calledResolve := mock.Repos.MockResolve_Local(t, "r", 1)
	var calledReposGet bool
	mock.Repos.Get_ = func(ctx context.Context, repo *sourcegraph.RepoSpec) (*sourcegraph.Repo, error) {
		calledReposGet = true
		return nil, grpc.Errorf(codes.Unknown, "error")
	}

	var res *repoResolution
	if err := c.GetJSON("/repos/r/-/resolve", &res); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(res, want) {
		t.Errorf("got %+v, want %+v", res, want)
	}
	if !*calledResolve {
		t.Error("!calledResolve")
	}
	if !calledReposGet {
		t.Error("!calledReposGet")
	}
}

func TestRepoResolve_Remote(t *testing.T) {
	c, mock := newTest()

	want := &repoResolution{
		Data: sourcegraph.RepoResolution{RemoteRepo: &sourcegraph.RemoteRepo{Name: "r"}},
	}

	calledResolve := mock.Repos.MockResolve_Remote(t, "r", &sourcegraph.RemoteRepo{Name: "r"})

	var res *repoResolution
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

func TestRepoResolve_notFound(t *testing.T) {
	c, mock := newTest()

	calledResolve := mock.Repos.MockResolve_NotFound(t, "r")

	resp, err := c.Get("/repos/r/-/resolve")
	if err != nil {
		t.Fatal(err)
	}
	if want := http.StatusNotFound; resp.StatusCode != want {
		t.Errorf("got HTTP %d, want %d", resp.StatusCode, want)
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

	calledReposResolve := mock.Repos.MockResolve_Local(t, "r/r", 1)
	calledGet := mock.Repos.MockGet_Return(t, &sourcegraph.Repo{
		ID:        1,
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
	if !*calledReposResolve {
		t.Error("!calledReposResolve")
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

	calledReposResolve := mock.Repos.MockResolve_Local(t, "r/r", 1)
	calledGet := mock.Repos.MockGet_Return(t, &sourcegraph.Repo{
		ID:        1,
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
	if !*calledReposResolve {
		t.Error("!calledReposResolve")
	}
	if !*calledGet {
		t.Error("!calledGet")
	}
}

func TestRepos(t *testing.T) {
	c, mock := newTest()

	wantRepos := &sourcegraph.RepoList{
		Repos: []*sourcegraph.Repo{{URI: "r/r"}},
	}

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
