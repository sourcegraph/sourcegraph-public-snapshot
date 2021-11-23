package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

type ConfStore interface {
}

type confStore struct {
	*basestore.Store
}

// SiteConfig contains the contents of a site config along with associated metadata.
type SiteConfig struct {
	ID        int32     // the unique ID of this config
	Contents  string    // the raw JSON content (with comments and trailing commas allowed)
	CreatedAt time.Time // the date when this config was created
	UpdatedAt time.Time // the date when this config was updated
}

var siteConfigColumns = []*sqlf.Query{
	sqlf.Sprintf("critical_and_site_config.id"),
	sqlf.Sprintf("critical_and_site_config.contents"),
	sqlf.Sprintf("critical_and_site_config.created_at"),
	sqlf.Sprintf("critical_and_site_config.updated_at"),
}

const getLatestFmtStr = `
SELECT %s -- siteConfigRows
FROM critical_and_site_config
WHERE type='site'
ORDER BY id DESC
LIMIT 1
`

func (s *confStore) GetLatest(ctx context.Context) (*SiteConfig, error) {
	q := sqlf.Sprintf(
		getLatestFmtStr,
		sqlf.Join(siteConfigColumns, ","),
	)
	row := s.QueryRow(ctx, q)
	config, err := scanSiteConfigRow(row)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		// No config has been written yet
		return nil, nil
	}
	return config, err
}

// scanSiteConfigRow scans a single row from a *sql.Row or *sql.Rows.
// It must be kept in sync with siteConfigColumns
func scanSiteConfigRow(scanner dbutil.Scanner) (*SiteConfig, error) {
	var s SiteConfig
	err := scanner.Scan(
		&s.ID,
		&s.Contents,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
	return &s, err
}
