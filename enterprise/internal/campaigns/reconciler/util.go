package reconciler

import (
	"context"

	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// TODO: copy-pasted
type RepoStore interface {
	Get(ctx context.Context, id api.RepoID) (*types.Repo, error)
}

// TODO: copy-pasted
type ExternalServiceStore interface {
	List(context.Context, database.ExternalServicesListOptions) ([]*types.ExternalService, error)
}

// TODO: copy-pasted
func loadExternalService(ctx context.Context, esStore ExternalServiceStore, repo *types.Repo) (*types.ExternalService, error) {
	args := database.ExternalServicesListOptions{IDs: repo.ExternalServiceIDs()}
	es, err := esStore.List(ctx, args)
	if err != nil {
		return nil, err
	}

	if len(es) == 0 {
		return nil, errors.Errorf("no external services found for repo %q", repo.Name)
	}

	return es[0], nil
}
