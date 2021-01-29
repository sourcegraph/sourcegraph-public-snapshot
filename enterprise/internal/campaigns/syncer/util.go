package syncer

import (
	"context"

	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

func loadExternalService(ctx context.Context, esStore ExternalServiceStore, repo *types.Repo) (*types.ExternalService, error) {
	var externalService *types.ExternalService
	args := database.ExternalServicesListOptions{IDs: repo.ExternalServiceIDs()}

	es, err := esStore.List(ctx, args)
	if err != nil {
		return nil, err
	}

	for _, e := range es {
		cfg, err := e.Configuration()
		if err != nil {
			return nil, err
		}

		switch cfg := cfg.(type) {
		case *schema.GitHubConnection:
			if cfg.Token != "" {
				externalService = e
			}
		case *schema.BitbucketServerConnection:
			if cfg.Token != "" {
				externalService = e
			}
		case *schema.GitLabConnection:
			if cfg.Token != "" {
				externalService = e
			}
		}
		if externalService != nil {
			break
		}
	}

	if externalService == nil {
		return nil, errors.Errorf("no external services found for repo %q", repo.Name)
	}

	return externalService, nil
}
