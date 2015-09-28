package vcsclient

import (
	"encoding/json"
	"net/http"
	"reflect"
	"testing"

	"sourcegraph.com/sourcegraph/go-vcs/vcs"
)

func TestRepository_CloneOrUpdate(t *testing.T) {
	setup()
	defer teardown()

	repoPath := "a.b/c"
	repo_, _ := vcsclient.Repository(repoPath)
	repo := repo_.(*repository)

	cloneURL := "git://a.b/c"
	opt := &CloneInfo{
		VCS:        "git",
		CloneURL:   cloneURL,
		RemoteOpts: vcs.RemoteOpts{SSH: &vcs.SSHConfig{PrivateKey: []byte("abc")}},
	}

	var called bool
	mux.HandleFunc(urlPath(t, RouteRepo, repo, nil), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "POST")

		body, _ := json.Marshal(opt)
		testBody(t, r, string(body)+"\n")
	})

	err := repo.CloneOrUpdate(opt)
	if err != nil {
		t.Errorf("Repository.CloneOrUpdate returned error: %v", err)
	}

	if !called {
		t.Fatal("!called")
	}
}

func TestRepository_ResolveBranch(t *testing.T) {
	setup()
	defer teardown()

	repoPath := "a.b/c"
	repo_, _ := vcsclient.Repository(repoPath)
	repo := repo_.(*repository)

	want := vcs.CommitID("abcd")

	var called bool
	mux.HandleFunc(urlPath(t, RouteRepoBranch, repo, map[string]string{"RepoPath": repoPath, "Branch": "mybranch"}), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "GET")

		http.Redirect(w, r, urlPath(t, RouteRepoCommit, repo, map[string]string{"CommitID": "abcd"}), http.StatusFound)
	})

	commitID, err := repo.ResolveBranch("mybranch")
	if err != nil {
		t.Errorf("Repository.ResolveBranch returned error: %v", err)
	}

	if !called {
		t.Fatal("!called")
	}

	if commitID != want {
		t.Errorf("Repository.ResolveBranch returned %+v, want %+v", commitID, want)
	}
}

func TestRepository_ResolveRevision(t *testing.T) {
	setup()
	defer teardown()

	repoPath := "a.b/c"
	repo_, _ := vcsclient.Repository(repoPath)
	repo := repo_.(*repository)

	want := vcs.CommitID("abcd")

	var called bool
	mux.HandleFunc(urlPath(t, RouteRepoRevision, repo, map[string]string{"RepoPath": repoPath, "RevSpec": "myrevspec"}), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "GET")

		http.Redirect(w, r, urlPath(t, RouteRepoCommit, repo, map[string]string{"CommitID": "abcd"}), http.StatusFound)
	})

	commitID, err := repo.ResolveRevision("myrevspec")
	if err != nil {
		t.Errorf("Repository.ResolveRevision returned error: %v", err)
	}

	if !called {
		t.Fatal("!called")
	}

	if commitID != want {
		t.Errorf("Repository.ResolveRevision returned %+v, want %+v", commitID, want)
	}
}

func TestRepository_ResolveTag(t *testing.T) {
	setup()
	defer teardown()

	repoPath := "a.b/c"
	repo_, _ := vcsclient.Repository(repoPath)
	repo := repo_.(*repository)

	want := vcs.CommitID("abcd")

	var called bool
	mux.HandleFunc(urlPath(t, RouteRepoTag, repo, map[string]string{"RepoPath": repoPath, "Tag": "mytag"}), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "GET")

		http.Redirect(w, r, urlPath(t, RouteRepoCommit, repo, map[string]string{"CommitID": "abcd"}), http.StatusFound)
	})

	commitID, err := repo.ResolveTag("mytag")
	if err != nil {
		t.Errorf("Repository.ResolveTag returned error: %v", err)
	}

	if !called {
		t.Fatal("!called")
	}

	if commitID != want {
		t.Errorf("Repository.ResolveTag returned %+v, want %+v", commitID, want)
	}
}

func TestRepository_Branches(t *testing.T) {
	setup()
	defer teardown()

	repoPath := "a.b/c"
	repo_, _ := vcsclient.Repository(repoPath)
	repo := repo_.(*repository)

	want := []*vcs.Branch{{Name: "mybranch", Head: "abcd"}}

	var called bool
	mux.HandleFunc(urlPath(t, RouteRepoBranches, repo, nil), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "GET")

		writeJSON(w, want)
	})

	branches, err := repo.Branches(vcs.BranchesOptions{})
	if err != nil {
		t.Errorf("Repository.Branches returned error: %v", err)
	}

	if !called {
		t.Fatal("!called")
	}

	if !reflect.DeepEqual(branches, want) {
		t.Errorf("Repository.Branches returned %+v, want %+v", branches, want)
	}
}

func TestRepository_Tags(t *testing.T) {
	setup()
	defer teardown()

	repoPath := "a.b/c"
	repo_, _ := vcsclient.Repository(repoPath)
	repo := repo_.(*repository)

	want := []*vcs.Tag{{Name: "mytag", CommitID: "abcd"}}

	var called bool
	mux.HandleFunc(urlPath(t, RouteRepoTags, repo, nil), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "GET")

		writeJSON(w, want)
	})

	tags, err := repo.Tags()
	if err != nil {
		t.Errorf("Repository.Tags returned error: %v", err)
	}

	if !called {
		t.Fatal("!called")
	}

	if !reflect.DeepEqual(tags, want) {
		t.Errorf("Repository.Tags returned %+v, want %+v", tags, want)
	}
}

func TestRepository_Commits(t *testing.T) {
	setup()
	defer teardown()

	repoPath := "a.b/c"
	repo_, _ := vcsclient.Repository(repoPath)
	repo := repo_.(*repository)

	want := []*vcs.Commit{{ID: "abcd"}}

	var called bool
	mux.HandleFunc(urlPath(t, RouteRepoCommits, repo, nil), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "GET")
		testFormValues(t, r, values{"Head": "abcd", "Base": "wxyz", "N": "2", "Skip": "3"})

		w.Header().Set(TotalCommitsHeader, "123")
		writeJSON(w, want)
	})

	commits, total, err := repo.Commits(vcs.CommitsOptions{Head: "abcd", Base: "wxyz", N: 2, Skip: 3})
	if err != nil {
		t.Errorf("Repository.Commits returned error: %v", err)
	}

	if !called {
		t.Fatal("!called")
	}

	if want := uint(123); total != want {
		t.Errorf("Repository.Commits: got total %d, want %d", total, want)
	}

	if !reflect.DeepEqual(commits, want) {
		t.Errorf("Repository.Commits returned %+v, want %+v", commits, want)
	}
}

func TestRepository_GetCommit(t *testing.T) {
	setup()
	defer teardown()

	repoPath := "a.b/c"
	repo_, _ := vcsclient.Repository(repoPath)
	repo := repo_.(*repository)

	want := &vcs.Commit{ID: "abcd"}

	var called bool
	mux.HandleFunc(urlPath(t, RouteRepoCommit, repo, map[string]string{"CommitID": "abcd"}), func(w http.ResponseWriter, r *http.Request) {
		called = true
		testMethod(t, r, "GET")

		writeJSON(w, want)
	})

	commit, err := repo.GetCommit("abcd")
	if err != nil {
		t.Errorf("Repository.GetCommit returned error: %v", err)
	}

	if !called {
		t.Fatal("!called")
	}

	if !reflect.DeepEqual(commit, want) {
		t.Errorf("Repository.GetCommit returned %+v, want %+v", commit, want)
	}
}
