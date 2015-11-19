package httpapi

import (
	"net/http"
	"reflect"
	"testing"
	"time"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
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

// Test that if the repo hasn't been modified since the client's
// If-Modified-Since, HTTP 304 Not Modified is returned.
func TestRepo_caching_notModified(t *testing.T) {
	c, mock := newTest()

	mtime := time.Now().UTC()

	calledGet := mock.Repos.MockGet_Return(t, &sourcegraph.Repo{
		URI:       "r/r",
		UpdatedAt: pbtypes.NewTimestamp(mtime),
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

	calledGet := mock.Repos.MockGet_Return(t, &sourcegraph.Repo{
		URI:       "r/r",
		UpdatedAt: pbtypes.NewTimestamp(mtime),
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

func TestRepos(t *testing.T) {
	c, mock := newTest()

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
