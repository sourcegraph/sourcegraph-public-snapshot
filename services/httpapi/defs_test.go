package httpapi

import (
	"net/http"
	"reflect"
	"testing"
	"time"

	"golang.org/x/net/context"

	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sqs/pbtypes"
)

func TestDefs(t *testing.T) {
	c, mock := newTest()

	wantDefs := &sourcegraph.DefList{
		Defs: []*sourcegraph.Def{{Def: graph.Def{Name: "d", Data: pbtypes.RawMessage("{}")}}},
	}

	calledList := mock.Defs.MockList(t, &sourcegraph.Def{Def: graph.Def{Name: "d", Data: pbtypes.RawMessage("{}")}})

	var defs *sourcegraph.DefList
	if err := c.GetJSON("/defs", &defs); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(defs, wantDefs) {
		t.Errorf("got defs %+v, want %+v", defs, wantDefs)
	}
	if !*calledList {
		t.Error("!calledList")
	}
}

// Test that if the builds corresponding to the requested defs are
// older than the the client's If-Modified-Since, HTTP 304 Not
// Modified is returned (because it's not possible for the defs to
// have changed without a newer build existing).
func TestDefs_caching_notModified(t *testing.T) {
	c, mock := newTest()

	mtime := time.Now().UTC()
	mtimeTS := pbtypes.NewTimestamp(mtime)

	calledReposResolve := mock.Repos.MockResolve_Local(t, "r", 1)
	calledList := mock.Defs.MockList(t)
	mock.Builds.List_ = func(ctx context.Context, op *sourcegraph.BuildListOptions) (*sourcegraph.BuildList, error) {
		return &sourcegraph.BuildList{Builds: []*sourcegraph.Build{
			{EndedAt: &mtimeTS},
		}}, nil
	}

	req, _ := http.NewRequest("GET", "/defs?RepoRevs=r@"+strings.Repeat("a", 40), nil)
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
	if *calledList {
		t.Error("Defs.List was called, but it should not have been (because the client's cache already holds the newest data)")
	}
}

// Test that if there are builds newer than the client's
// If-Modified-Since, the defs are returned.
func TestDefs_caching_modifiedSince(t *testing.T) {
	c, mock := newTest()

	mtime := time.Now().UTC()
	mtimeTS := pbtypes.NewTimestamp(mtime)

	calledReposResolve := mock.Repos.MockResolve_Local(t, "r", 1)
	mock.Builds.List_ = func(ctx context.Context, op *sourcegraph.BuildListOptions) (*sourcegraph.BuildList, error) {
		return &sourcegraph.BuildList{Builds: []*sourcegraph.Build{
			{EndedAt: &mtimeTS},
		}}, nil
	}
	calledList := mock.Defs.MockList(t)

	req, _ := http.NewRequest("GET", "/defs?RepoRevs=r@"+strings.Repeat("a", 40), nil)
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
	if !*calledList {
		t.Error("!calledList")
	}
}
