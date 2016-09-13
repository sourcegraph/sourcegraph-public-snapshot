package httpapi

import (
	"net/http"
	"reflect"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
)

func TestRepoTree_file(t *testing.T) {
	c, mock := newTest()

	want := &sourcegraph.TreeEntry{
		BasicTreeEntry: &sourcegraph.BasicTreeEntry{
			Name:     "f",
			Type:     sourcegraph.FileEntry,
			Contents: []byte("c"),
		},
	}

	calledReposResolve := mock.Repos.MockResolve_Local(t, "r", 1)
	calledGet := mock.RepoTree.MockGet_Return_NoCheck(t, want)
	calledReposResolveRev := mock.Repos.MockResolveRev_NoCheck(t, "c")
	calledAnnotationsList := mock.Annotations.MockList(t, nil)
	mock.Repos.MockGet_Return(t, &sourcegraph.Repo{ID: 1})

	var entry *sourcegraph.TreeEntry
	if err := c.GetJSON("/repos/r/-/tree/f", &entry); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(entry, want) {
		t.Errorf("got %+v, want %+v", entry, want)
	}
	if !*calledReposResolve {
		t.Error("!calledReposResolve")
	}
	if !*calledGet {
		t.Error("!calledGet")
	}
	if !*calledReposResolveRev {
		t.Error("!calledReposResolveRev")
	}
	if !*calledAnnotationsList {
		t.Error("!calledAnnotationsList")
	}
}

func TestRepoTree_dir(t *testing.T) {
	c, mock := newTest()

	want := &sourcegraph.TreeEntry{
		BasicTreeEntry: &sourcegraph.BasicTreeEntry{
			Name: "d",
			Type: sourcegraph.DirEntry,
			Entries: []*sourcegraph.BasicTreeEntry{
				{Name: "f", Type: sourcegraph.FileEntry},
			},
		},
	}

	calledReposResolve := mock.Repos.MockResolve_Local(t, "r", 1)
	calledGet := mock.RepoTree.MockGet_Return_NoCheck(t, want)
	calledReposResolveRev := mock.Repos.MockResolveRev_NoCheck(t, "c")
	mock.Repos.MockGet_Return(t, &sourcegraph.Repo{ID: 1})

	var entry *sourcegraph.TreeEntry
	if err := c.GetJSON("/repos/r/-/tree/f", &entry); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(entry, want) {
		t.Errorf("got %+v, want %+v", entry, want)
	}
	if !*calledReposResolve {
		t.Error("!calledReposResolve")
	}
	if !*calledGet {
		t.Error("!calledGet")
	}
	if !*calledReposResolveRev {
		t.Error("!calledReposResolveRev")
	}
}

func TestRepoTree_notFound(t *testing.T) {
	c, mock := newTest()

	calledReposResolve := mock.Repos.MockResolve_Local(t, "r", 1)
	calledGet := mock.RepoTree.MockGet_NotFound(t)
	calledReposResolveRev := mock.Repos.MockResolveRev_NoCheck(t, "c")

	resp, err := c.Get("/repos/r/-/tree/f")
	if err != nil {
		t.Fatal(err)
	}
	if want := http.StatusNotFound; resp.StatusCode != want {
		t.Errorf("got HTTP %d, want %d", resp.StatusCode, want)
	}
	if !*calledReposResolve {
		t.Error("!calledReposResolve")
	}
	if !*calledGet {
		t.Error("!calledGet")
	}
	if !*calledReposResolveRev {
		t.Error("!calledReposResolveRev")
	}
}
