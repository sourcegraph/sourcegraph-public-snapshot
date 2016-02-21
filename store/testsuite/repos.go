package testsuite

import (
	"testing"

	"golang.org/x/net/context"

	"sort"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store"
)

func Repos_Update_Visibility(ctx context.Context, t *testing.T, s store.Repos) {
	// Add a repo.
	if err := s.Create(ctx, &sourcegraph.Repo{URI: "a/b", VCS: "git"}); err != nil {
		t.Fatal(err)
	}

	// Verify visibility is public by default.
	repo, err := s.Get(ctx, "a/b")
	if err != nil {
		t.Fatal(err)
	}
	if want := false; repo.Private != want {
		t.Errorf("got private %v, want %v", repo.Private, want)
	}

	// Verify visibility gets updated to private.
	if err := s.Update(ctx, &store.RepoUpdate{ReposUpdateOp: &sourcegraph.ReposUpdateOp{Repo: sourcegraph.RepoSpec{URI: "a/b"}, IsPrivate: true}}); err != nil {
		t.Fatal(err)
	}

	repo, err = s.Get(ctx, "a/b")
	if err != nil {
		t.Fatal(err)
	}
	if want := true; repo.Private != want {
		t.Errorf("got private %v, want %v", repo.Private, want)
	}

	// Verify visibility gets updated to public.
	if err := s.Update(ctx, &store.RepoUpdate{ReposUpdateOp: &sourcegraph.ReposUpdateOp{Repo: sourcegraph.RepoSpec{URI: "a/b"}, IsPublic: true}}); err != nil {
		t.Fatal(err)
	}
	repo, err = s.Get(ctx, "a/b")
	if err != nil {
		t.Fatal(err)
	}
	if want := false; repo.Private != want {
		t.Errorf("got private %v, want %v", repo.Private, want)
	}

	// Verify bad arguments return error.
	if err := s.Update(ctx, &store.RepoUpdate{ReposUpdateOp: &sourcegraph.ReposUpdateOp{Repo: sourcegraph.RepoSpec{URI: "a/b"}, IsPrivate: true, IsPublic: true}}); err == nil {
		t.Errorf("got nil error, want bad args to fail")
	}
}

func repoURIs(repos []*sourcegraph.Repo) []string {
	var uris []string
	for _, repo := range repos {
		uris = append(uris, repo.URI)
	}
	sort.Strings(uris)
	return uris
}

func isRepoNotFound(err error) bool {
	_, ok := err.(*store.RepoNotFoundError)
	return ok
}
