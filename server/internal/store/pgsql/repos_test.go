package pgsql

import (
	"testing"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

func (s *Repos) mustCreate(ctx context.Context, t *testing.T, repos ...*sourcegraph.Repo) []*sourcegraph.Repo {
	var createdRepos []*sourcegraph.Repo
	for _, repo := range repos {
		// All pgsql repos must be mirrors, so it's more efficient
		// to set that here than in EVERY test.
		repo.Mirror = true
		repo.HTTPCloneURL = "http://example.com/dummy.git"

		createdRepo, err := s.Create(ctx, repo)
		if err != nil {
			t.Fatal(err)
		}
		createdRepos = append(createdRepos, createdRepo)
	}
	return createdRepos
}

func TestRepos_List_byOwner_empty(t *testing.T) {
	var s Repos

	testUserSpec := sourcegraph.UserSpec{Login: "alice"}

	repos, err := s.List(nil, &sourcegraph.RepoListOptions{Owner: testUserSpec.SpecString()})
	if err != nil {
		t.Fatal(err)
	}
	if len(repos) != 0 {
		t.Errorf("got repos == %v, want empty", repos)
	}
}
