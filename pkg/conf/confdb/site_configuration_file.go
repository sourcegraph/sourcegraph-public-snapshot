package confdb

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	multierror "github.com/hashicorp/go-multierror"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/jsonx"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

// CoreSiteConfigurationFiles provides methods to read and write the
// core/site configuration files to the database.
type CoreSiteConfigurationFiles struct {
	// Conn is a function that returns the connection that is used to connect to the database.
	Conn func() *sql.DB
}

// SiteCreateIfUpToDate saves the given site configuration "contents" to the database iff the
// supplied "lastID" is equal to the one that was most recently saved to the database.
//
// The site configuration that was most recently saved to the database is returned.
// An error is returned if "contents" is invalid JSON.
//
// 🚨 SECURITY: This method does NOT verify the user is an admin. The caller is
// responsible for ensuring this or that the response never makes it to a user.
func (c *CoreSiteConfigurationFiles) SiteCreateIfUpToDate(ctx context.Context, lastID *int32, contents string) (latest *api.SiteConfigurationFile, err error) {
	coreSiteFile, err := c.createIfUpToDate(ctx, kindCore, lastID, contents)
	siteFile := (*api.SiteConfigurationFile)(coreSiteFile)

	return siteFile, err
}

// CoreCreateIfUpToDate saves the given core configuration "contents" to the database iff the
// supplied "lastID" is equal to the one that was most recently saved to the database.
//
// The core configuration that was most recently saved to the database is returned.
// An error is returned if "contents" is invalid JSON.
//
// 🚨 SECURITY: This method does NOT verify the user is an admin. The caller is
// responsible for ensuring this or that the response never makes it to a user.
func (c *CoreSiteConfigurationFiles) CoreCreateIfUpToDate(ctx context.Context, lastID *int32, contents string) (latest *api.CoreConfigurationFile, err error) {
	coreSiteFile, err := c.createIfUpToDate(ctx, kindCore, lastID, contents)
	coreFile := (*api.CoreConfigurationFile)(coreSiteFile)

	return coreFile, err
}

// SiteGetLatest returns the site configuration file that was most recently saved to the database.
// This returns nil, nil if there is not yet a site configuration in the database.
//
// 🚨 SECURITY: This method does NOT verify the user is an admin. The caller is
// responsible for ensuring this or that the response never makes it to a user.
func (c *CoreSiteConfigurationFiles) SiteGetLatest(ctx context.Context) (latest *api.SiteConfigurationFile, err error) {
	coreSiteFile, err := c.getLatest(ctx, kindSite)
	siteFile := (*api.SiteConfigurationFile)(coreSiteFile)

	return siteFile, err
}

// CoreGetLatest returns core site configuration file that was most recently saved to the database.
// This returns nil, nil if there is not yet a core configuration in the database.
//
// 🚨 SECURITY: This method does NOT verify the user is an admin. The caller is
// responsible for ensuring this or that the response never makes it to a user.
func (c *CoreSiteConfigurationFiles) CoreGetLatest(ctx context.Context) (latest *api.CoreConfigurationFile, err error) {
	coreSiteFile, err := c.getLatest(ctx, kindCore)
	coreFile := (*api.CoreConfigurationFile)(coreSiteFile)

	return coreFile, err
}

func (c *CoreSiteConfigurationFiles) createIfUpToDate(ctx context.Context, kind configurationKind, lastID *int32, contents string) (latest *api.CoreSiteConfigurationFile, err error) {
	// Validate JSON syntax before saving.
	if _, errs := jsonx.Parse(contents, jsonx.ParseOptions{Comments: true, TrailingCommas: true}); len(errs) > 0 {
		return nil, fmt.Errorf("invalid settings JSON: %v", errs)
	}

	newFile := api.CoreSiteConfigurationFile{
		Contents: contents,
	}

	tx, err := c.Conn().BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			rollErr := tx.Rollback()
			if rollErr != nil {
				err = multierror.Append(err, rollErr)
			}
			return
		}
		err = tx.Commit()
	}()

	latestFile, err := c.getLatestUnderTx(ctx, tx, kind)
	if err != nil {
		return nil, err
	}

	creatorIsUpToDate := latestFile != nil && lastID != nil && latestFile.ID == *lastID
	if latestFile == nil || creatorIsUpToDate {
		err := tx.QueryRow(
			fmt.Sprintf("INSERT INTO %s(contents) VALUES($1) RETURNING id, created_at, updated_at", tableName(kind)),
			newFile.Contents).Scan(&newFile.ID, &newFile.CreatedAt, &newFile.UpdatedAt)
		if err != nil {
			return nil, err
		}
		latestFile = &newFile
	}

	return latestFile, nil
}

func (c *CoreSiteConfigurationFiles) getLatest(ctx context.Context, kind configurationKind) (*api.CoreSiteConfigurationFile, error) {
	return c.getLatestUnderTx(ctx, c.Conn(), kind)
}

func (c *CoreSiteConfigurationFiles) getLatestUnderTx(ctx context.Context, queryTarget queryable, kind configurationKind) (*api.CoreSiteConfigurationFile, error) {
	q := sqlf.Sprintf(fmt.Sprintf("SELECT s.id, s.contents, s.created_at, s.updated_at FROM %s s ORDER BY id DESC LIMIT 1", tableName(kind)))
	log.Printf("%+v", q)
	rows, err := queryTarget.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return nil, err
	}
	files, err := c.parseQueryRows(ctx, rows)
	if err != nil {
		return nil, err
	}
	if len(files) != 1 {
		// No configuration file has been written yet.
		return nil, nil
	}
	return files[0], nil
}

func (c *CoreSiteConfigurationFiles) parseQueryRows(ctx context.Context, rows *sql.Rows) ([]*api.CoreSiteConfigurationFile, error) {
	files := []*api.CoreSiteConfigurationFile{}
	defer rows.Close()
	for rows.Next() {
		f := api.CoreSiteConfigurationFile{}
		err := rows.Scan(&f.ID, &f.Contents, &f.CreatedAt, &f.UpdatedAt)
		if err != nil {
			return nil, err
		}
		files = append(files, &f)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return files, nil
}

// queryable allows us to reuse the same logic for certain operations both
// inside and outside an explicit transaction.
type queryable interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

type configurationKind int

const (
	kindSite configurationKind = iota
	kindCore
)

func tableName(k configurationKind) string {
	if k == kindSite {
		return "site_configuration_files"
	}
	return "core_configuration_files"
}
