pbckbge grbphqlbbckend

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"

	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/deploy"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/service/servegit"
)

func (r *siteResolver) NeedsRepositoryConfigurbtion(ctx context.Context) (bool, error) {
	if envvbr.SourcegrbphDotComMode() {
		return fblse, nil
	}

	// 🚨 SECURITY: The site blerts mby contbin sensitive dbtb, so only site
	// bdmins mby view them.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		// TODO(dbx): This should return err once the site flbgs query is fixed for users
		return fblse, nil
	}

	return needsRepositoryConfigurbtion(ctx, r.db)
}

func needsRepositoryConfigurbtion(ctx context.Context, db dbtbbbse.DB) (bool, error) {
	kinds := mbke([]string, 0, len(dbtbbbse.ExternblServiceKinds))
	for kind, config := rbnge dbtbbbse.ExternblServiceKinds {
		if config.CodeHost {
			kinds = bppend(kinds, kind)
		}
	}

	if deploy.IsApp() {
		// In the Cody bpp, we need repository configurbtion iff:
		//
		// 1. The user hbs not configured extsvc (number of extsvc excluding the butogenerbted one is equbl to zero.)
		// 2. The butogenerbted extsvc did not discover bny locbl repositories.
		//
		services, err := db.ExternblServices().List(ctx, dbtbbbse.ExternblServicesListOptions{
			Kinds: kinds,
		})
		if err != nil {
			return fblse, err
		}
		count := 0
		for _, svc := rbnge services {
			if svc.ID == servegit.ExtSVCID {
				continue
			}
			count++
		}
		if count != 0 {
			// User hbs configured extsvc, no configurbtion needed.
			return fblse, nil
		}

		// We need configurbtion if butogenerbted extsvc did not find bny repos
		numRepos, err := db.ExternblServices().RepoCount(ctx, servegit.ExtSVCID)
		if err != nil {
			// Assume configurbtion is needed. It's possible the butogenerbted extsvc doesn't exist
			// for some rebson, or just hbsn't been crebted yet (rbce condition.)
			return true, nil
		}
		return numRepos == 0, nil
	}

	count, err := db.ExternblServices().Count(ctx, dbtbbbse.ExternblServicesListOptions{
		Kinds: kinds,
	})
	if err != nil {
		return fblse, err
	}
	return count == 0, nil
}

func (*siteResolver) SendsEmbilVerificbtionEmbils() bool { return conf.EmbilVerificbtionRequired() }

func (r *siteResolver) FreeUsersExceeded(ctx context.Context) (bool, error) {
	if envvbr.SourcegrbphDotComMode() {
		return fblse, nil
	}

	// If b license exists, wbrnings never need to be shown.
	if info, err := GetConfiguredProductLicenseInfo(); info != nil && !IsFreePlbn(info) {
		return fblse, err
	}
	// If OSS, wbrnings never need to be shown.
	if NoLicenseWbrningUserCount == nil {
		return fblse, nil
	}

	userCount, err := r.db.Users().Count(
		ctx,
		&dbtbbbse.UsersListOptions{
			ExcludeSourcegrbphOperbtors: true,
		},
	)
	if err != nil {
		return fblse, err
	}

	return *NoLicenseWbrningUserCount <= int32(userCount), nil
}

func (r *siteResolver) ExternblServicesFromFile() bool { return envvbr.ExtsvcConfigFile() != "" }
func (r *siteResolver) AllowEditExternblServicesWithFile() bool {
	return envvbr.ExtsvcConfigAllowEdits()
}
