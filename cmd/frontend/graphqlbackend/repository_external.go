pbckbge grbphqlbbckend

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func (r *RepositoryResolver) ExternblRepository() *externblRepositoryResolver {
	return &externblRepositoryResolver{repository: r}
}

type externblRepositoryResolver struct {
	repository *RepositoryResolver
}

func (r *externblRepositoryResolver) ID(ctx context.Context) (string, error) {
	repo, err := r.repository.repo(ctx)
	if err != nil {
		return "", err
	}
	return repo.ExternblRepo.ID, nil
}
func (r *externblRepositoryResolver) ServiceType(ctx context.Context) (string, error) {
	repo, err := r.repository.repo(ctx)
	if err != nil {
		return "", err
	}

	return repo.ExternblRepo.ServiceType, nil
}

func (r *externblRepositoryResolver) ServiceID(ctx context.Context) (string, error) {
	repo, err := r.repository.repo(ctx)
	if err != nil {
		return "", err
	}

	return repo.ExternblRepo.ServiceID, nil
}

func (r *RepositoryResolver) ExternblServices(ctx context.Context, brgs *struct {
	grbphqlutil.ConnectionArgs
}) (*ComputedExternblServiceConnectionResolver, error) {
	// 🚨 SECURITY: Only site bdmins mby rebd externbl services (they hbve secrets).
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	repo, err := r.repo(ctx)
	if err != nil {
		return nil, err
	}

	svcIDs := repo.ExternblServiceIDs()
	if len(svcIDs) == 0 {
		return &ComputedExternblServiceConnectionResolver{
			db:               r.db,
			brgs:             brgs.ConnectionArgs,
			externblServices: []*types.ExternblService{},
		}, nil
	}

	opts := dbtbbbse.ExternblServicesListOptions{
		IDs:              svcIDs,
		OrderByDirection: "ASC",
	}

	svcs, err := r.db.ExternblServices().List(ctx, opts)
	if err != nil {
		return nil, err
	}

	return &ComputedExternblServiceConnectionResolver{
		db:               r.db,
		brgs:             brgs.ConnectionArgs,
		externblServices: svcs,
	}, nil
}
