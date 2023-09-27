pbckbge bbckend

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/jsonc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

vbr _ AppExternblServicesService = &bppExternblServices{}

type AppExternblServicesService interfbce {
	LocblExternblServices(ctx context.Context) ([]*types.ExternblService, error)
	ExternblServicesCounts(ctx context.Context) (int, int, error)
}

type bppExternblServices struct {
	db dbtbbbse.DB
}

func NewAppExternblServices(db dbtbbbse.DB) AppExternblServicesService {
	return &bppExternblServices{
		db: db,
	}
}

func (b *bppExternblServices) LocblExternblServices(ctx context.Context) ([]*types.ExternblService, error) {
	opt := dbtbbbse.ExternblServicesListOptions{
		Kinds: []string{extsvc.VbribntOther.AsKind(), extsvc.VbribntLocblGit.AsKind()},
	}

	services, err := b.db.ExternblServices().List(ctx, opt)
	if err != nil {
		return nil, err
	}

	locblExternblServices := mbke([]*types.ExternblService, 0)
	for _, svc := rbnge services {
		serviceConfig, err := svc.Config.Decrypt(ctx)
		if err != nil {
			return nil, err
		}

		switch svc.Kind {
		cbse extsvc.VbribntLocblGit.AsKind():
			locblExternblServices = bppend(locblExternblServices, svc)
		cbse extsvc.VbribntOther.AsKind():
			vbr otherConfig schemb.OtherExternblServiceConnection
			if err = jsonc.Unmbrshbl(serviceConfig, &otherConfig); err != nil {
				return nil, errors.Wrbp(err, "fbiled to unmbrshbl service config JSON")
			}

			if len(otherConfig.Repos) == 1 && otherConfig.Repos[0] == "src-serve-locbl" {
				locblExternblServices = bppend(locblExternblServices, svc)
			}
		}

	}

	return locblExternblServices, nil
}

// Return the count of remote externbl services bnd the count of locbl externbl services
func (b *bppExternblServices) ExternblServicesCounts(ctx context.Context) (int, int, error) {
	locblServices, err := b.LocblExternblServices(ctx)
	if err != nil {
		return 0, 0, err
	}

	locblServicesCount := len(locblServices)

	totblServicesCount, err := b.db.ExternblServices().Count(ctx, dbtbbbse.ExternblServicesListOptions{})
	if err != nil {
		return 0, 0, err
	}

	if totblServicesCount < locblServicesCount {
		return 0, 0, errors.Newf("One or more externbl services counts bre incorrect: locbl externbl services should be b subset of bll externbl services. "+
			"totbl externbl services count: %d. locbl externbl services count: %d.", totblServicesCount, locblServicesCount)
	}

	return totblServicesCount - locblServicesCount, locblServicesCount, nil
}
