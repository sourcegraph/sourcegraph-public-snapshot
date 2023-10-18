package database

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/confdefaults"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
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
	SiteCreateIfUpToDate(ctx context.Context, lastID *int32, authorUserID int32, contents string, isOverride bool) (*SiteConfig, error)

	// SiteGetLatest returns the site config that was most recently saved to the database.
	// This returns nil, nil if there is not yet a site config in the database.
	//
	// 🚨 SECURITY: This method does NOT verify the user is an admin. The caller is
	// responsible for ensuring this or that the response never makes it to a user.
	SiteGetLatest(ctx context.Context) (*SiteConfig, error)

	// ListSiteConfigs will list the configs of type "site".
	//
	// 🚨 SECURITY: This method does NOT verify the user is an admin. The caller is
	// responsible for ensuring this or that the response never makes it to a user.
	ListSiteConfigs(context.Context, *PaginationArgs) ([]*SiteConfig, error)

	// GetSiteConfig will return the total count of all configs of type "site".
	//
	// 🚨 SECURITY: This method does NOT verify the user is an admin. The caller is
	// responsible for ensuring this or that the response never makes it to a user.
	GetSiteConfigCount(context.Context) (int, error)

	Transact(ctx context.Context) (ConfStore, error)
	Done(error) error
	basestore.ShareableStore
}

// ErrNewerEdit is returned by SiteCreateIfUpToDate when a newer edit has already been applied and
// the edit has been rejected.
var ErrNewerEdit = errors.New("someone else has already applied a newer edit")

// ConfStoreWith instantiates and returns a new ConfStore using
// the other store handle.
func ConfStoreWith(other basestore.ShareableStore) ConfStore {
	return &confStore{
		Store:  basestore.NewWithHandle(other.Handle()),
		logger: log.Scoped("confStore"),
	}
}

type confStore struct {
	*basestore.Store
	logger log.Logger
}

// SiteConfig contains the contents of a site config along with associated metadata.
type SiteConfig struct {
	ID               int32  // the unique ID of this config
	AuthorUserID     int32  // the user id of the author that updated this config
	Contents         string // the raw JSON content (with comments and trailing commas allowed)
	RedactedContents string // the raw JSON content but with sensitive fields redacted

	CreatedAt time.Time // the date when this config was created
	UpdatedAt time.Time // the date when this config was updated
}

var siteConfigColumns = []*sqlf.Query{
	sqlf.Sprintf("critical_and_site_config.id"),
	sqlf.Sprintf("critical_and_site_config.author_user_id"),
	sqlf.Sprintf("critical_and_site_config.contents"),
	sqlf.Sprintf("critical_and_site_config.redacted_contents"),
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
	return &confStore{
		Store:  txBase,
		logger: s.logger,
	}, nil
}

func (s *confStore) SiteCreateIfUpToDate(ctx context.Context, lastID *int32, authorUserID int32, contents string, isOverride bool) (_ *SiteConfig, err error) {
	tx, err := s.transact(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	newLastID, err := tx.addDefault(ctx, authorUserID, confdefaults.Default.Site)
	if err != nil {
		return nil, err
	}
	if newLastID != nil {
		lastID = newLastID
	}
	return tx.createIfUpToDate(ctx, lastID, authorUserID, contents, isOverride)
}

func (s *confStore) SiteGetLatest(ctx context.Context) (_ *SiteConfig, err error) {
	tx, err := s.transact(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	// If an actor is associated with this context then we will be able to write the user id to the
	// actor_user_id column. But if it is not associated with an actor, then user id is 0 and NULL
	// will be written to the database instead.
	_, err = tx.addDefault(ctx, actor.FromContext(ctx).UID, confdefaults.Default.Site)
	if err != nil {
		return nil, err
	}

	return tx.getLatest(ctx)
}

const listSiteConfigsFmtStr = `
SELECT
	id,
	author_user_id,
	contents,
	redacted_contents,
	created_at,
	updated_at
FROM (
	SELECT
		*,
		LAG(redacted_contents) OVER (ORDER BY id) AS prev_redacted_contents
	FROM
		critical_and_site_config) t
WHERE
(%s)
`

func (s *confStore) ListSiteConfigs(ctx context.Context, paginationArgs *PaginationArgs) ([]*SiteConfig, error) {
	where := []*sqlf.Query{
		sqlf.Sprintf("(prev_redacted_contents IS NULL OR redacted_contents != prev_redacted_contents)"),
		sqlf.Sprintf("redacted_contents IS NOT NULL"),
		sqlf.Sprintf(`type = 'site'`),
	}

	// This will fetch all site configs.
	if paginationArgs == nil {
		query := sqlf.Sprintf(listSiteConfigsFmtStr, sqlf.Join(where, "AND"))
		rows, err := s.Query(ctx, query)
		return scanSiteConfigs(rows, err)
	}

	args := paginationArgs.SQL()

	if args.Where != nil {
		where = append(where, args.Where)
	}

	query := sqlf.Sprintf(listSiteConfigsFmtStr, sqlf.Join(where, "AND"))
	query = args.AppendOrderToQuery(query)
	query = args.AppendLimitToQuery(query)

	rows, err := s.Query(ctx, query)
	return scanSiteConfigs(rows, err)
}

const getSiteConfigCount = `
SELECT
	COUNT(*)
FROM (
	SELECT
		*,
		LAG(redacted_contents) OVER (ORDER BY id) AS prev_redacted_contents
	FROM
		critical_and_site_config) t
WHERE (prev_redacted_contents IS NULL
	OR redacted_contents != prev_redacted_contents)
AND redacted_contents IS NOT NULL
AND type = 'site'
`

func (s *confStore) GetSiteConfigCount(ctx context.Context) (int, error) {
	q := sqlf.Sprintf(getSiteConfigCount)

	var count int
	err := s.QueryRow(ctx, q).Scan(&count)
	return count, err
}

func (s *confStore) addDefault(ctx context.Context, authorUserID int32, contents string) (newLastID *int32, _ error) {
	latest, err := s.getLatest(ctx)
	if err != nil {
		return nil, err
	}
	if latest != nil {
		// We have an existing config!
		return nil, nil
	}

	latest, err = s.createIfUpToDate(ctx, nil, authorUserID, contents, true)
	if err != nil {
		return nil, err
	}
	return &latest.ID, nil
}

const createSiteConfigFmtStr = `
INSERT INTO critical_and_site_config (type, author_user_id, contents, redacted_contents)
VALUES ('site', %s, %s, %s)
RETURNING %s -- siteConfigColumns
`

func (s *confStore) createIfUpToDate(ctx context.Context, lastID *int32, authorUserID int32, contents string, isOverride bool) (*SiteConfig, error) {
	// Validate config for syntax and by the JSON Schema.
	var problems []string
	var err error
	if isOverride {
		var problemStruct conf.Problems
		problemStruct, err = conf.Validate(conftypes.RawUnified{Site: contents})
		problems = problemStruct.Messages()
	} else {
		problems, err = conf.ValidateSite(contents)
	}
	if err != nil {
		return nil, errors.Errorf("failed to validate site configuration: %w", err)
	} else if len(problems) > 0 {
		return nil, errors.Errorf("site configuration is invalid: %s", strings.Join(problems, ","))
	}

	latest, err := s.getLatest(ctx)
	if err != nil {
		return nil, err
	}
	if latest != nil && lastID != nil && latest.ID != *lastID {
		return nil, ErrNewerEdit
	}

	redactedConf, err := conf.RedactAndHashSecrets(conftypes.RawUnified{Site: contents})
	var redactedContents string
	if err != nil {
		// Do not fail here. Instead continue writing to DB with an empty value for
		// "redacted_contents".
		s.logger.Warn(
			"failed to redact secrets during site config creation (secrets are safely stored but diff generation in site config history will not work)",
			log.Error(err),
		)
	} else {
		redactedContents = redactedConf.Site
	}

	q := sqlf.Sprintf(
		createSiteConfigFmtStr,
		dbutil.NullInt32Column(authorUserID),
		contents,
		redactedContents,
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
		&dbutil.NullInt32{N: &s.AuthorUserID},
		&s.Contents,
		&dbutil.NullString{S: &s.RedactedContents},
		&s.CreatedAt,
		&s.UpdatedAt,
	)
	return &s, err
}

var scanSiteConfigs = basestore.NewSliceScanner(scanSiteConfigRow)
