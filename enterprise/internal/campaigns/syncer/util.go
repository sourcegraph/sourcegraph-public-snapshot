package syncer

import (
	"context"

	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

// loadExternalService looks up all external services that are connected to the given repo.
// The first external service to have a token configured will be returned then.
// If no external service matching the above criteria is found, an error is returned.
func loadExternalService(ctx context.Context, esStore ExternalServiceStore, repo *types.Repo) (*types.ExternalService, error) {
	es, err := esStore.List(ctx, database.ExternalServicesListOptions{IDs: repo.ExternalServiceIDs()})
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
				return e, nil
			}
		case *schema.BitbucketServerConnection:
			if cfg.Token != "" {
				return e, nil
			}
		case *schema.GitLabConnection:
			if cfg.Token != "" {
				return e, nil
			}
		}
	}

	return nil, errors.Errorf("no external services found for repo %q", repo.Name)
}
