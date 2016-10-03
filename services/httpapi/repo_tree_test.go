package httpapi

import (
	"net/http"
	"reflect"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend"
)

func TestRepoTree_file(t *testing.T) {
	c := newTest()

	want := &sourcegraph.TreeEntry{
		BasicTreeEntry: &sourcegraph.BasicTreeEntry{
			Name:     "f",
			Type:     sourcegraph.FileEntry,
			Contents: []byte("c"),
		},
	}

	calledReposResolve := backend.Mocks.Repos.MockResolve_Local(t, "r", 1)
	calledGet := backend.Mocks.RepoTree.MockGet_Return_NoCheck(t, want)
	calledReposResolveRev := backend.Mocks.Repos.MockResolveRev_NoCheck(t, "c")
	calledAnnotationsList := backend.Mocks.Annotations.MockList(t, nil)
	backend.Mocks.Repos.MockGet_Return(t, &sourcegraph.Repo{ID: 1})

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
	c := newTest()

	want := &sourcegraph.TreeEntry{
		BasicTreeEntry: &sourcegraph.BasicTreeEntry{
			Name: "d",
			Type: sourcegraph.DirEntry,
			Entries: []*sourcegraph.BasicTreeEntry{
				{Name: "f", Type: sourcegraph.FileEntry},
			},
		},
	}

	calledReposResolve := backend.Mocks.Repos.MockResolve_Local(t, "r", 1)
	calledGet := backend.Mocks.RepoTree.MockGet_Return_NoCheck(t, want)
	calledReposResolveRev := backend.Mocks.Repos.MockResolveRev_NoCheck(t, "c")
	backend.Mocks.Repos.MockGet_Return(t, &sourcegraph.Repo{ID: 1})

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
	c := newTest()

	calledReposResolve := backend.Mocks.Repos.MockResolve_Local(t, "r", 1)
	calledGet := backend.Mocks.RepoTree.MockGet_NotFound(t)
	calledReposResolveRev := backend.Mocks.Repos.MockResolveRev_NoCheck(t, "c")

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
