package db

import (
	"context"
	"database/sql"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db/dbconn"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

type siteConfigurationFiles struct{}

// GetLatest returns the site configuration file that was most recently saved to the database.
//
// 🚨 SECURITY: This method does NOT verify the user is an admin. The caller is
// responsible for ensuring this or that the response never makes it to a user.
func (s *siteConfigurationFiles) GetLatest(ctx context.Context, subject api.ConfigurationSubject) (*api.SiteConfigurationFile, error) {
	return s.getLatest(ctx, dbconn.Global)
}

func (s *siteConfigurationFiles) getLatest(ctx context.Context, queryTarget queryable) (*api.SiteConfigurationFile, error) {
	q := sqlf.Sprintf(`SELECT s.id, s.contents, s.created_at FROM site_configuration_files s ORDER BY id DESC LIMIT 1`)
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

func (s *siteConfigurationFiles) parseQueryRows(ctx context.Context, rows *sql.Rows) ([]*api.SiteConfigurationFile, error) {
	files := []*api.SiteConfigurationFile{}
	defer rows.Close()
	for rows.Next() {
		f := api.SiteConfigurationFile{}
		err := rows.Scan(&f.ID, &f.Contents, &f.CreatedAt)
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
