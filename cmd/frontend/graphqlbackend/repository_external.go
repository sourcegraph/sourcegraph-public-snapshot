package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func (r *RepositoryResolver) ExternalRepository() *externalRepositoryResolver {
	return &externalRepositoryResolver{repository: r}
}

type externalRepositoryResolver struct {
	repository *RepositoryResolver
}

func (r *externalRepositoryResolver) ID(ctx context.Context) (string, error) {
	repo, err := r.repository.repo(ctx)
	if err != nil {
		return "", err
	}
	return repo.ExternalRepo.ID, nil
}
func (r *externalRepositoryResolver) ServiceType(ctx context.Context) (string, error) {
	repo, err := r.repository.repo(ctx)
	if err != nil {
		return "", err
	}

	return repo.ExternalRepo.ServiceType, nil
}

func (r *externalRepositoryResolver) ServiceID(ctx context.Context) (string, error) {
	repo, err := r.repository.repo(ctx)
	if err != nil {
		return "", err
	}

	return repo.ExternalRepo.ServiceID, nil
}

func (r *RepositoryResolver) ExternalServices(ctx context.Context, args *struct {
	graphqlutil.ConnectionArgs
}) (*ComputedExternalServiceConnectionResolver, error) {
	// 🚨 SECURITY: Only site admins may read external services (they have secrets).
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	repo, err := r.repo(ctx)
	if err != nil {
		return nil, err
	}

	svcIDs := repo.ExternalServiceIDs()
	if len(svcIDs) == 0 {
		return &ComputedExternalServiceConnectionResolver{
			db:               r.db,
			args:             args.ConnectionArgs,
			externalServices: []*types.ExternalService{},
		}, nil
	}

	opts := database.ExternalServicesListOptions{
		IDs:              svcIDs,
		OrderByDirection: "ASC",
	}

	svcs, err := r.db.ExternalServices().List(ctx, opts)
	if err != nil {
		return nil, err
	}

	return &ComputedExternalServiceConnectionResolver{
		db:               r.db,
		args:             args.ConnectionArgs,
		externalServices: svcs,
	}, nil
}
