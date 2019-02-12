package repos

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/kylelemons/godebug/pretty"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

func TestSyncer_diff(t *testing.T) {
	t.Fatal("Tests not finished yet. TODO")

	for _, tc := range []struct {
		name            string
		sourced, stored []*Repo
		diff            Diff
	}{
		{
			name:    "empty inputs",
			sourced: []*Repo{},
			stored:  []*Repo{},
			diff:    Diff{},
		},
		{
			name: "nil inputs",
			diff: Diff{},
		},
		{
			name: "added",
			diff: Diff{},
		},
	} {
		var s Syncer
		diff := s.diff(tc.sourced, tc.stored)
		if cmp := pretty.Compare(diff, tc.diff); cmp != "" {
			t.Errorf("Diff:\n%s", cmp)
		}
	}
}

type mockSource struct {
	repos []*Repo
	err   error
}

func (s mockSource) ListRepos(ctx context.Context) ([]*Repo, error) {
	return s.repos, s.err
}

type mockStore struct {
	mockSource
	repos []*Repo
	err   error
}

func (s *mockStore) UpsertRepos(ctx context.Context, repos ...*Repo) error {
	for i, r := range s.repos {
		*repos[i] = *r
	}
	return s.err
}

func fakeRepo(id, name string) *Repo {
	return &Repo{
		Name:        name,
		Description: fmt.Sprintf("%q description", name),
		Language:    "fakelang",
		Enabled:     true,
		Archived:    false,
		Fork:        false,
		CreatedAt:   time.Now().UTC(),
		ExternalRepo: api.ExternalRepoSpec{
			ID:          id,
			ServiceType: "github",
			ServiceID:   "http://github.com",
		},
	}
}
