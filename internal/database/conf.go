package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/jsonx"

	"github.com/sourcegraph/sourcegraph/internal/conf/confdefaults"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// ConfStore is a store that interacts with the config tables.
//
// Only the frontend should use this store.  All other users should go through
// the conf package and NOT interact with the database on their own.
type ConfStore interface {
	// SiteCreateIfUpToDate saves the given site config "contents" to the database iff the
	// supplied "lastID" is equal to the one that was most recently saved to the database.
	//
	// The site config that was most recently saved to the database is returned.
	// An error is returned if "contents" is invalid JSON.
	//
	// 🚨 SECURITY: This method does NOT verify the user is an admin. The caller is
	// responsible for ensuring this or that the response never makes it to a user.
	SiteCreateIfUpToDate(ctx context.Context, lastID *int32, contents string) (*SiteConfig, error)

	// SiteGetLatest returns the site config that was most recently saved to the database.
	// This returns nil, nil if there is not yet a site config in the database.
	//
	// 🚨 SECURITY: This method does NOT verify the user is an admin. The caller is
	// responsible for ensuring this or that the response never makes it to a user.
	SiteGetLatest(ctx context.Context) (*SiteConfig, error)

	Transact(ctx context.Context) (ConfStore, error)
	Done(error) error
	basestore.ShareableStore
}

// ErrNewerEdit is returned by SiteCreateIfUpToDate when a newer edit has already been applied and
// the edit has been rejected.
var ErrNewerEdit = errors.New("someone else has already applied a newer edit")

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

func (s *confStore) Transact(ctx context.Context) (ConfStore, error) {
	return s.transact(ctx)
}

func (s *confStore) transact(ctx context.Context) (*confStore, error) {
	txBase, err := s.Store.Transact(ctx)
	if err != nil {
		return nil, err
	}
	return &confStore{Store: txBase}, nil
}

func (s *confStore) SiteCreateIfUpToDate(ctx context.Context, lastID *int32, contents string) (_ *SiteConfig, err error) {
	tx, err := s.transact(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	newLastID, err := tx.addDefault(ctx, confdefaults.Default.Site)
	if err != nil {
		return nil, err
	}
	if newLastID != nil {
		lastID = newLastID
	}
	return tx.createIfUpToDate(ctx, lastID, contents)
}

func (s *confStore) SiteGetLatest(ctx context.Context) (_ *SiteConfig, err error) {
	tx, err := s.transact(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	_, err = tx.addDefault(ctx, confdefaults.Default.Site)
	if err != nil {
		return nil, err
	}

	return tx.getLatest(ctx)
}

func (s *confStore) addDefault(ctx context.Context, contents string) (newLastID *int32, _ error) {
	latest, err := s.getLatest(ctx)
	if err != nil {
		return nil, err
	}
	if latest != nil {
		// We have an existing config!
		return nil, nil
	}

	latest, err = s.createIfUpToDate(ctx, nil, contents)
	if err != nil {
		return nil, err
	}
	return &latest.ID, nil
}

const createSiteConfigFmtStr = `
INSERT INTO critical_and_site_config (type, contents)
VALUES ('site', %s)
RETURNING %s -- siteConfigColumns
`

func (s *confStore) createIfUpToDate(ctx context.Context, lastID *int32, contents string) (*SiteConfig, error) {
	// Validate JSON syntax before saving.
	if _, errs := jsonx.Parse(contents, jsonx.ParseOptions{Comments: true, TrailingCommas: true}); len(errs) > 0 {
		return nil, errors.Errorf("invalid settings JSON: %v", errs)
	}

	latest, err := s.getLatest(ctx)
	if err != nil {
		return nil, err
	}
	if latest != nil && lastID != nil && latest.ID != *lastID {
		return nil, ErrNewerEdit
	}

	q := sqlf.Sprintf(
		createSiteConfigFmtStr,
		contents,
		sqlf.Join(siteConfigColumns, ","),
	)
	row := s.QueryRow(ctx, q)
	return scanSiteConfigRow(row)
}

const getLatestFmtStr = `
SELECT %s -- siteConfigRows
FROM critical_and_site_config
WHERE type='site'
ORDER BY id DESC
LIMIT 1
`

func (s *confStore) getLatest(ctx context.Context) (*SiteConfig, error) {
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
