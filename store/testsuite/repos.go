package testsuite

import (
	"reflect"
	"testing"
	"time"

	"golang.org/x/net/context"

	"sort"

	"sourcegraph.com/sqs/pbtypes"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store"
)

// Repos_Get_nonexistent tests the behavior of Repos.Get when called
// on a repo that does not exist.
func Repos_Get_nonexistent(ctx context.Context, t *testing.T, s store.Repos, nonexistentRepo string) {
	repo, err := s.Get(ctx, nonexistentRepo)
	if !isRepoNotFound(err) {
		t.Fatal(err)
	}
	if repo != nil {
		t.Error("repo != nil")
	}
}

// Repos_List_query tests the behavior of Repos.List when called with
// a query.
func Repos_List_query(ctx context.Context, t *testing.T, s store.Repos) {
	// Add some repos.
	if err := s.Create(ctx, &sourcegraph.Repo{URI: "abc/def", Name: "def", VCS: "git"}); err != nil {
		t.Fatal(err)
	}
	if err := s.Create(ctx, &sourcegraph.Repo{URI: "def/ghi", Name: "ghi", VCS: "git"}); err != nil {
		t.Fatal(err)
	}
	if err := s.Create(ctx, &sourcegraph.Repo{URI: "jkl/mno/pqr", Name: "pqr", VCS: "git"}); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		query string
		want  []string
	}{
		{"de", []string{"abc/def", "def/ghi"}},
		{"def", []string{"abc/def", "def/ghi"}},
		{"ABC/DEF", []string{"abc/def"}},
		{"xyz", nil},
	}
	for _, test := range tests {
		repos, err := s.List(ctx, &sourcegraph.RepoListOptions{Query: test.query})
		if err != nil {
			t.Fatal(err)
		}
		if got := repoURIs(repos); !reflect.DeepEqual(got, test.want) {
			t.Errorf("%q: got repos %v, want %v", test.query, got, test.want)
		}
	}
}

// Repos_List_URIs tests the behavior of Repos.List when called with
// URIs.
func Repos_List_URIs(ctx context.Context, t *testing.T, s store.Repos) {
	// Add some repos.
	if err := s.Create(ctx, &sourcegraph.Repo{URI: "a/b", VCS: "git"}); err != nil {
		t.Fatal(err)
	}
	if err := s.Create(ctx, &sourcegraph.Repo{URI: "c/d", VCS: "git"}); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		uris []string
		want []string
	}{
		{[]string{"a/b"}, []string{"a/b"}},
		{[]string{"x/y"}, nil},
		{[]string{"a/b", "c/d"}, []string{"a/b", "c/d"}},
		{[]string{"a/b", "x/y", "c/d"}, []string{"a/b", "c/d"}},
	}
	for _, test := range tests {
		repos, err := s.List(ctx, &sourcegraph.RepoListOptions{URIs: test.uris})
		if err != nil {
			t.Fatal(err)
		}
		if got := repoURIs(repos); !reflect.DeepEqual(got, test.want) {
			t.Errorf("%v: got repos %v, want %v", test.uris, got, test.want)
		}
	}
}

func Repos_Create(ctx context.Context, t *testing.T, s store.Repos) {
	tm := time.Now().Round(time.Second)
	ts := pbtypes.NewTimestamp(tm)

	// Add a repo.
	if err := s.Create(ctx, &sourcegraph.Repo{URI: "a/b", CreatedAt: &ts, VCS: "git"}); err != nil {
		t.Fatal(err)
	}

	repo, err := s.Get(ctx, "a/b")
	if err != nil {
		t.Fatal(err)
	}
	if repo.CreatedAt == nil {
		t.Fatal("got CreatedAt nil")
	}
	if want := ts.Time(); !repo.CreatedAt.Time().Equal(want) {
		t.Errorf("got CreatedAt %q, want %q", repo.CreatedAt.Time(), want)
	}
}

func Repos_Create_dupe(ctx context.Context, t *testing.T, s store.Repos) {
	tm := time.Now().Round(time.Second)
	ts := pbtypes.NewTimestamp(tm)

	// Add a repo.
	if err := s.Create(ctx, &sourcegraph.Repo{URI: "a/b", CreatedAt: &ts, VCS: "git"}); err != nil {
		t.Fatal(err)
	}

	// Add another repo with the same name.
	if err := s.Create(ctx, &sourcegraph.Repo{URI: "a/b", CreatedAt: &ts, VCS: "git"}); err == nil {
		t.Fatalf("got err == nil, want an error when creating a duplicate repo")
	}
}

// Repos_Update_Description tests the behavior of Repos.Update to
// update a repo's description.
func Repos_Update_Description(ctx context.Context, t *testing.T, s store.Repos) {
	// Add a repo.
	if err := s.Create(ctx, &sourcegraph.Repo{URI: "a/b", VCS: "git"}); err != nil {
		t.Fatal(err)
	}

	repo, err := s.Get(ctx, "a/b")
	if err != nil {
		t.Fatal(err)
	}
	if want := ""; repo.Description != want {
		t.Errorf("got description %q, want %q", repo.Description, want)
	}

	if err := s.Update(ctx, &store.RepoUpdate{ReposUpdateOp: &sourcegraph.ReposUpdateOp{Repo: sourcegraph.RepoSpec{URI: "a/b"}, Description: "d"}}); err != nil {
		t.Fatal(err)
	}

	repo, err = s.Get(ctx, "a/b")
	if err != nil {
		t.Fatal(err)
	}
	if want := "d"; repo.Description != want {
		t.Errorf("got description %q, want %q", repo.Description, want)
	}
}

func Repos_Update_UpdatedAt(ctx context.Context, t *testing.T, s store.Repos) {
	// Add a repo.
	if err := s.Create(ctx, &sourcegraph.Repo{URI: "a/b", VCS: "git"}); err != nil {
		t.Fatal(err)
	}

	repo, err := s.Get(ctx, "a/b")
	if err != nil {
		t.Fatal(err)
	}
	if repo.UpdatedAt != nil {
		t.Errorf("got UpdatedAt %v, want nil", repo.UpdatedAt.Time())
	}

	// Perform any update.
	newTime := time.Unix(123456, 0)
	if err := s.Update(ctx, &store.RepoUpdate{ReposUpdateOp: &sourcegraph.ReposUpdateOp{Repo: sourcegraph.RepoSpec{URI: "a/b"}}, UpdatedAt: &newTime}); err != nil {
		t.Fatal(err)
	}

	repo, err = s.Get(ctx, "a/b")
	if err != nil {
		t.Fatal(err)
	}
	if repo.UpdatedAt == nil {
		t.Fatal("got UpdatedAt nil, want non-nil")
	}
	if want := newTime; !repo.UpdatedAt.Time().Equal(want) {
		t.Errorf("got UpdatedAt %q, want %q", repo.UpdatedAt.Time(), want)
	}
}

func Repos_Update_PushedAt(ctx context.Context, t *testing.T, s store.Repos) {
	// Add a repo.
	if err := s.Create(ctx, &sourcegraph.Repo{URI: "a/b", VCS: "git"}); err != nil {
		t.Fatal(err)
	}

	repo, err := s.Get(ctx, "a/b")
	if err != nil {
		t.Fatal(err)
	}
	if repo.PushedAt != nil {
		t.Errorf("got PushedAt %v, want nil", repo.PushedAt.Time())
	}

	newTime := time.Unix(123456, 0)
	if err := s.Update(ctx, &store.RepoUpdate{ReposUpdateOp: &sourcegraph.ReposUpdateOp{Repo: sourcegraph.RepoSpec{URI: "a/b"}}, PushedAt: &newTime}); err != nil {
		t.Fatal(err)
	}

	repo, err = s.Get(ctx, "a/b")
	if err != nil {
		t.Fatal(err)
	}
	if repo.PushedAt == nil {
		t.Fatal("got PushedAt nil, want non-nil")
	}
	if repo.UpdatedAt != nil {
		t.Fatal("got UpdatedAt non-nil, want nil")
	}
	if want := newTime; !repo.PushedAt.Time().Equal(want) {
		t.Errorf("got PushedAt %q, want %q", repo.PushedAt.Time(), want)
	}
}

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
