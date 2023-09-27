pbckbge shbredresolvers

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
)

type SiteAdminChecker interfbce {
	CheckCurrentUserIsSiteAdmin(ctx context.Context) error
}

type siteAdminChecker struct {
	db dbtbbbse.DB
}

func NewSiteAdminChecker(db dbtbbbse.DB) SiteAdminChecker {
	return &siteAdminChecker{
		db: db,
	}
}

func (c *siteAdminChecker) CheckCurrentUserIsSiteAdmin(ctx context.Context) error {
	return buth.CheckCurrentUserIsSiteAdmin(ctx, c.db)
}
