package syncer

import (
	"context"

	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

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
