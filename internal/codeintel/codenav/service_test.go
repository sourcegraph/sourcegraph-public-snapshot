package codenav

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	internaltypes "github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type AllPresentFakeRepoStore struct{}

var _ minimalRepoStore = AllPresentFakeRepoStore{}

func (s AllPresentFakeRepoStore) Get(_ context.Context, id api.RepoID) (*internaltypes.Repo, error) {
	return &internaltypes.Repo{ID: id, Name: api.RepoName(fmt.Sprintf("r%d", id))}, nil
}

func (s AllPresentFakeRepoStore) GetReposSetByIDs(ctx context.Context, ids ...api.RepoID) (map[api.RepoID]*internaltypes.Repo, error) {
	out := map[api.RepoID]*internaltypes.Repo{}
	for _, id := range ids {
		r, _ := s.Get(ctx, id) // Get doesn't error so this is OK
		out[id] = r
	}
	return out, nil
}

type FakeMinimalRepoStore struct {
	data map[api.RepoID]*internaltypes.Repo
}

var _ minimalRepoStore = FakeMinimalRepoStore{}

func (f FakeMinimalRepoStore) Get(ctx context.Context, id api.RepoID) (*internaltypes.Repo, error) {
	if r, ok := f.data[id]; ok {
		return r, nil
	}
	return nil, &database.RepoNotFoundErr{ID: id}
}

func (f FakeMinimalRepoStore) GetReposSetByIDs(ctx context.Context, ids ...api.RepoID) (map[api.RepoID]*internaltypes.Repo, error) {
	out := map[api.RepoID]*internaltypes.Repo{}
	for _, id := range ids {
		r, err := f.Get(ctx, id)
		if err != nil {
			if errors.Is(err, &database.RepoNotFoundErr{}) {
				continue
			}
			return nil, err
		}
		out[id] = r
	}
	return out, nil
}
