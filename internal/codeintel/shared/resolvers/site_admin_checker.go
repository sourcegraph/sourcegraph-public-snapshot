package sharedresolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

type SiteAdminChecker interface {
	CheckCurrentUserIsSiteAdmin(ctx context.Context) error
}

type siteAdminChecker struct {
	db database.DB
}

func NewSiteAdminChecker(db database.DB) SiteAdminChecker {
	return &siteAdminChecker{
		db: db,
	}
}

func (c *siteAdminChecker) CheckCurrentUserIsSiteAdmin(ctx context.Context) error {
	return auth.CheckCurrentUserIsSiteAdmin(ctx, c.db)
}
