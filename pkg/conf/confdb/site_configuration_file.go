package confdb

import (
	"context"
	"database/sql"
	"fmt"

	multierror "github.com/hashicorp/go-multierror"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/jsonx"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

// SiteConfigurationFiles provides methods to read and write site configuration
// files to the database.
type SiteConfigurationFiles struct {
	// Conn is a function that returns the connection that is used to connect to the database.
	Conn func() *sql.DB
}

// CreateIfUpToDate saves the given site configuration "contents" to the database iff the
// supplied "lastID" is equal to the one that was most recently saved to the database.
//
// The site configuration that was most recently saved to the database is returned.
// An error is returned if "contents" is invalid JSON.
//
// 🚨 SECURITY: This method does NOT verify the user is an admin. The caller is
// responsible for ensuring this or that the response never makes it to a user.
func (s *SiteConfigurationFiles) CreateIfUpToDate(ctx context.Context, lastID *int32, contents string) (latest *api.SiteConfigurationFile, err error) {
	// Validate JSON syntax before saving.
	if _, errs := jsonx.Parse(contents, jsonx.ParseOptions{Comments: true, TrailingCommas: true}); len(errs) > 0 {
		return nil, fmt.Errorf("invalid settings JSON: %v", errs)
	}

	newFile := api.SiteConfigurationFile{
		Contents: contents,
	}

	tx, err := s.Conn().BeginTx(ctx, nil)
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

	latestFile, err := s.getLatest(ctx, tx)
	if err != nil {
		return nil, err
	}

	creatorIsUpToDate := latestFile != nil && lastID != nil && latestFile.ID == *lastID
	if latestFile == nil || creatorIsUpToDate {
		err := tx.QueryRow(
			"INSERT INTO site_configuration_files(contents) VALUES($1) RETURNING id, created_at, updated_at",
			newFile.Contents).Scan(&newFile.ID, &newFile.CreatedAt, &newFile.UpdatedAt)
		if err != nil {
			return nil, err
		}
		latestFile = &newFile
	}

	return latestFile, nil
}

// GetLatest returns the site configuration file that was most recently saved to the database.
// This returns nil, nil if there is not yet a site configuration in the database.
//
// 🚨 SECURITY: This method does NOT verify the user is an admin. The caller is
// responsible for ensuring this or that the response never makes it to a user.
func (s *SiteConfigurationFiles) GetLatest(ctx context.Context) (*api.SiteConfigurationFile, error) {
	return s.getLatest(ctx, s.Conn())
}

func (s *SiteConfigurationFiles) getLatest(ctx context.Context, queryTarget queryable) (*api.SiteConfigurationFile, error) {
	q := sqlf.Sprintf(`SELECT s.id, s.contents, s.created_at, s.updated_at FROM site_configuration_files s ORDER BY id DESC LIMIT 1`)
	rows, err := queryTarget.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return nil, err
	}
	files, err := s.parseQueryRows(ctx, rows)
	if err != nil {
		return nil, err
	}
	if len(files) != 1 {
		// No site configuration file has been written yet.
		return nil, nil
	}
	return files[0], nil
}

func (s *SiteConfigurationFiles) parseQueryRows(ctx context.Context, rows *sql.Rows) ([]*api.SiteConfigurationFile, error) {
	files := []*api.SiteConfigurationFile{}
	defer rows.Close()
	for rows.Next() {
		f := api.SiteConfigurationFile{}
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
