package vcsclient

import (
	"net/http"
	"reflect"
	"testing"

	"sourcegraph.com/sourcegraph/go-vcs/vcs"
)

func TestRepository_Search(t *testing.T) {
	setup()
	defer teardown()

	repoPath := "a.b/c"
	repo_, _ := vcsclient.Repository(repoPath)
	repo := repo_.(*repository)

	want := []*vcs.SearchResult{{File: "f", StartLine: 1, EndLine: 2, Match: []byte("xyz")}}

	var called bool
	mux.HandleFunc(urlPath(t, RouteRepoSearch, repo, map[string]string{"RepoPath": repoPath, "CommitID": "c"}), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "GET")
		testFormValues(t, r, values{"Query": "q", "QueryType": "t", "ContextLines": "0", "N": "0", "Offset": "0"})

		writeJSON(w, want)
	})

	res, err := repo.Search("c", vcs.SearchOptions{Query: "q", QueryType: "t"})
	if err != nil {
		t.Errorf("Repository.Search returned error: %v", err)
	}

	if !called {
		t.Fatal("!called")
	}

	if !reflect.DeepEqual(res, want) {
		t.Errorf("Repository.Search returned %+v, want %+v", res, want)
	}
}
