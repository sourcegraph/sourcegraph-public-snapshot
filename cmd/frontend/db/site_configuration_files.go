package db

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db/dbconn"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/conf/confdb"
)

type siteConfigurationFiles struct{}

// CreateIfUpToDate saves the given site configuration "contents" to the database iff the
// supplied "lastID" is equal to the one that was most recently saved to the database.
//
// The site configuration that was most recently saved to the database is returned.
// An error is returned if "contents" is invalid JSON.
//
// 🚨 SECURITY: This method does NOT verify the user is an admin. The caller is
// responsible for ensuring this or that the response never makes it to a user.
func (s *siteConfigurationFiles) CreateIfUpToDate(ctx context.Context, lastID *int32, contents string) (latest *api.SiteConfigurationFile, err error) {
	handler := &confdb.SiteConfigurationFiles{Conn: dbconn.Global}
	return handler.CreateIfUpToDate(ctx, lastID, contents)

}

// GetLatest returns the site configuration file that was most recently saved to the database.
// This returns nil, nil if there is not yet a site configuration in the database.
//
// 🚨 SECURITY: This method does NOT verify the user is an admin. The caller is
// responsible for ensuring this or that the response never makes it to a user.
func (s *siteConfigurationFiles) GetLatest(ctx context.Context) (*api.SiteConfigurationFile, error) {
	handler := &confdb.SiteConfigurationFiles{Conn: dbconn.Global}
	return handler.GetLatest(ctx)
}
