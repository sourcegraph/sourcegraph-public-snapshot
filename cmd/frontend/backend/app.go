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

func (a *appExternalServices) localRepositoriesCount(ctx context.Context) (int32, error) {
	services, err := a.LocalExternalServices(ctx)
	if err != nil {
		return 0, err
	}

	var localReposCount int32
	for _, svc := range services {
		count, err := a.db.ExternalServices().RepoCount(ctx, svc.ID)

		if err != nil {
			return 0, err
		}

		localReposCount += count
	}
	return localReposCount, nil
}

// Return the count of remote repositories and the count of local repositories
func (a *appExternalServices) RepositoriesCounts(ctx context.Context) (int32, int32, error) {
	localCount, err := a.localRepositoriesCount(ctx)
	if err != nil {
		return 0, 0, err
	}

	repoStatistics, err := a.db.RepoStatistics().GetRepoStatistics(ctx)
	if err != nil {
		return 0, 0, err
	}

	totalCount := int32(repoStatistics.Total)

	if totalCount < localCount {
		return 0, 0, errors.Newf("One or more Repos counts are incorrect: local repos should be a subset of all repositories. "+
			"total repos count: %d. local repos count: %d.", totalCount, localCount)
	}

	return totalCount - localCount, localCount, nil
}
