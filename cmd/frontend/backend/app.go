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
	ExternalServicesCounts(ctx context.Context) (int, int, error)
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
		Kinds: []string{extsvc.VariantOther.AsKind(), extsvc.VariantLocalGit.AsKind()},
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

		switch svc.Kind {
		case extsvc.VariantLocalGit.AsKind():
			localExternalServices = append(localExternalServices, svc)
		case extsvc.VariantOther.AsKind():
			var otherConfig schema.OtherExternalServiceConnection
			if err = jsonc.Unmarshal(serviceConfig, &otherConfig); err != nil {
				return nil, errors.Wrap(err, "failed to unmarshal service config JSON")
			}

			if len(otherConfig.Repos) == 1 && otherConfig.Repos[0] == "src-serve-local" {
				localExternalServices = append(localExternalServices, svc)
			}
		}

	}

	return localExternalServices, nil
}

// Return the count of remote external services and the count of local external services
func (a *appExternalServices) ExternalServicesCounts(ctx context.Context) (int, int, error) {
	localServices, err := a.LocalExternalServices(ctx)
	if err != nil {
		return 0, 0, err
	}

	localServicesCount := len(localServices)

	totalServicesCount, err := a.db.ExternalServices().Count(ctx, database.ExternalServicesListOptions{})
	if err != nil {
		return 0, 0, err
	}

	if totalServicesCount < localServicesCount {
		return 0, 0, errors.Newf("One or more external services counts are incorrect: local external services should be a subset of all external services. "+
			"total external services count: %d. local external services count: %d.", totalServicesCount, localServicesCount)
	}

	return totalServicesCount - localServicesCount, localServicesCount, nil
}
