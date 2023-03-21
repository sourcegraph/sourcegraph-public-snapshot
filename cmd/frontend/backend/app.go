package backend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

var _ AppExternalServicesService = &appExternalServices{}

type AppExternalServicesService interface {
	LocalExternalServices(ctx context.Context) ([]*types.ExternalService, error)
	RepositoriesCounts(ctx context.Context) (int32, int32, error)
}

type appExternalServices struct {
	db database.DB
}

func NewAppExternalServices(db database.DB) AppExternalServicesService {
	return &appExternalServices{
		db: db,
	}
}

func (a *appExternalServices) LocalExternalServices(ctx context.Context) ([]*types.ExternalService, error) {
	opt := database.ExternalServicesListOptions{
		Kinds: []string{extsvc.KindOther},
	}

	services, err := a.db.ExternalServices().List(ctx, opt)
	if err != nil {
		return nil, err
	}

	localExternalServices := make([]*types.ExternalService, 0)
	for _, svc := range services {
		serviceConfig, err := svc.Config.Decrypt(ctx)
		if err != nil {
			return nil, err
		}

		var otherConfig schema.OtherExternalServiceConnection
		if err = jsonc.Unmarshal(serviceConfig, &otherConfig); err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal service config JSON")
		}

		if len(otherConfig.Repos) == 1 && otherConfig.Repos[0] == "src-serve-local" {
			localExternalServices = append(localExternalServices, svc)
		}
	}

	return localExternalServices, nil
}

// Return the count of remote repositories and the count of local repositories
func (a *appExternalServices) RepositoriesCounts(ctx context.Context) (int32, int32, error) {
	localServices, err := a.LocalExternalServices(ctx)

	var localReposCount int32

	localIds := make(map[int64]*types.ExternalService)

	if err != nil {
		return 0, 0, err
	}

	// local external services
	// TODO: Obtain each count in one for loop
	for _, svc := range localServices {
		localIds[svc.ID] = svc
		count, err := a.db.ExternalServices().RepoCount(ctx, svc.ID)
		if err == nil {
			// TODO: identify repos sync'd by multiple external services to get accurate repositories set
			localReposCount += count
		}
	}

	var remoteReposCount int32

	// all external services
	services, err := a.db.ExternalServices().List(ctx, database.ExternalServicesListOptions{})

	if err == nil {
		for _, svc := range services {
			if _, ok := localIds[svc.ID]; !ok {
				count, err := a.db.ExternalServices().RepoCount(ctx, svc.ID)
				if err == nil {
					// TODO: identify repos sync'd by multiple external services to get accurate repositories set
					remoteReposCount += count
				}
			}
		}
	}

	return remoteReposCount, localReposCount, nil
}
