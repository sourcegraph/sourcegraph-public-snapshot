package localstore

import (
	"testing"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
)

func (s *repos) mustCreate(ctx context.Context, t *testing.T, repos ...*sourcegraph.Repo) []*sourcegraph.Repo {
	var createdRepos []*sourcegraph.Repo
	for _, repo := range repos {
		repo.DefaultBranch = "master"

		if _, err := s.Create(ctx, repo); err != nil {
			t.Fatal(err)
		}
		repo, err := s.GetByURI(ctx, repo.URI)
		if err != nil {
			t.Fatal(err)
		}
		createdRepos = append(createdRepos, repo)
	}
	return createdRepos
}

func TestRepos_List_byOwner_empty(t *testing.T) {
	var s repos

	repos, err := s.List(context.Background(), &sourcegraph.RepoListOptions{Owner: "alice"})
	if err != nil {
		t.Fatal(err)
	}
	if len(repos) != 0 {
		t.Errorf("got repos == %v, want empty", repos)
	}
}
