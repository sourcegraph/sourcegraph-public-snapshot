package oobmigration

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
)

// Migration stores metadata and tracks progress of an out-of-band migration routine.
// These fields mirror the out_of_band_migrations table in the database. For docs see
// the [schema](https://github.com/sourcegraph/sourcegraph/blob/main/internal/database/schema.md#table-publicout_of_band_migrations).
type Migration struct {
	ID             int
	Team           string
	Component      string
	Description    string
	Introduced     Version
	Deprecated     *Version
	Progress       float64
	Created        time.Time
	LastUpdated    *time.Time
	NonDestructive bool
	ApplyReverse   bool
	Errors         []MigrationError
	// Metadata can be used to store custom JSON data
	Metadata json.RawMessage
}

// Complete returns true if the migration has 0 un-migrated record in whichever
// direction is indicated by the ApplyReverse flag.
func (m Migration) Complete() bool {
	if m.Progress == 1 && !m.ApplyReverse {
		return true
	}

	if m.Progress == 0 && m.ApplyReverse {
		return true
	}

	return false
}

// MigrationError pairs an error message and the time the error occurred.
type MigrationError struct {
	Message string
	Created time.Time
}

// scanMigrations scans a slice of migrations from the return value of `*Store.query`.
func scanMigrations(rows *sql.Rows, queryErr error) (_ []Migration, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	var values []Migration
	for rows.Next() {
		var message string
		var created *time.Time
		var deprecatedMajor, deprecatedMinor *int
		value := Migration{Errors: []MigrationError{}}

		if err := rows.Scan(
			&value.ID,
			&value.Team,
			&value.Component,
			&value.Description,
			&value.Introduced.Major,
			&value.Introduced.Minor,
			&deprecatedMajor,
			&deprecatedMinor,
			&value.Progress,
			&value.Created,
			&value.LastUpdated,
			&value.NonDestructive,
			&value.ApplyReverse,
			&value.Metadata,
			&dbutil.NullString{S: &message},
			&created,
		); err != nil {
			return nil, err
		}

		if message != "" {
			value.Errors = append(value.Errors, MigrationError{
				Message: message,
				Created: *created,
			})
		}

		if deprecatedMajor != nil && deprecatedMinor != nil {
			value.Deprecated = &Version{
				Major: *deprecatedMajor,
				Minor: *deprecatedMinor,
			}
		}

		if n := len(values); n > 0 && values[n-1].ID == value.ID {
			values[n-1].Errors = append(values[n-1].Errors, value.Errors...)
		} else {
			values = append(values, value)
		}
	}

	return values, nil
}

// Store is the interface over the out-of-band migrations tables.
type Store struct {
	*basestore.Store[schemas.Production]
}

// NewStoreWithDB creates a new Store with the given database connection.
func NewStoreWithDB(db database.DB) *Store {
	return &Store{Store: basestore.NewWithHandle(db.Handle())}
}

var _ basestore.ShareableStore[schemas.Production] = &Store{}

// With creates a new store with the underlying database handle from the given store.
// This method should be used when two distinct store instances need to perform an
// operation within the same shared transaction.
//
// This method wraps the basestore.With method.
func (s *Store) With(other basestore.ShareableStore[schemas.Production]) *Store {
	return &Store{Store: s.Store.With(other)}
}

// Transact returns a new store whose methods operate within the context of a new transaction
// or a new savepoint. This method will return an error if the underlying connection cannot be
// interface upgraded to a TxBeginner.
//
// This method wraps the basestore.Transact method.
func (s *Store) Transact(ctx context.Context) (*Store, error) {
	txBase, err := s.Store.Transact(ctx)
	return &Store{Store: txBase}, err
}

// GetByID retrieves a migration by its identifier. If the migration does not exist, a false
// valued flag is returned.
func (s *Store) GetByID(ctx context.Context, id int) (_ Migration, _ bool, err error) {
	migrations, err := scanMigrations(s.Store.Query(ctx, sqlf.Sprintf(getByIDQuery, id)))
	if err != nil {
		return Migration{}, false, err
	}

	if len(migrations) == 0 {
		return Migration{}, false, nil
	}

	return migrations[0], true, nil
}

const getByIDQuery = `
-- source: internal/oobmigration/store.go:GetByID
SELECT
	m.id,
	m.team,
	m.component,
	m.description,
	m.introduced_version_major,
	m.introduced_version_minor,
	m.deprecated_version_major,
	m.deprecated_version_minor,
	m.progress,
	m.created,
	m.last_updated,
	m.non_destructive,
	m.apply_reverse,
	m.metadata,
	e.message,
	e.created
FROM out_of_band_migrations m
LEFT JOIN out_of_band_migrations_errors e ON e.migration_id = m.id
WHERE m.id = %s
ORDER BY e.created desc
`

// ReturnEnterpriseMigrations is set by the enterprise application to enable the
// inclusion of enterprise-only migration records in the output of oobmigration.List.
var ReturnEnterpriseMigrations = false

// List returns the complete list of out-of-band migrations.
func (s *Store) List(ctx context.Context) (_ []Migration, err error) {
	var conds []*sqlf.Query
	if !ReturnEnterpriseMigrations {
		conds = append(conds, sqlf.Sprintf("NOT m.is_enterprise"))
	}

	if len(conds) == 0 {
		conds = append(conds, sqlf.Sprintf("TRUE"))
	}

	return scanMigrations(s.Store.Query(ctx, sqlf.Sprintf(listQuery, sqlf.Join(conds, "AND"))))
}

const listQuery = `
-- source: internal/oobmigration/store.go:List
SELECT
	m.id,
	m.team,
	m.component,
	m.description,
	m.introduced_version_major,
	m.introduced_version_minor,
	m.deprecated_version_major,
	m.deprecated_version_minor,
	m.progress,
	m.created,
	m.last_updated,
	m.non_destructive,
	m.apply_reverse,
	m.metadata,
	e.message,
	e.created
FROM out_of_band_migrations m
LEFT JOIN out_of_band_migrations_errors e ON e.migration_id = m.id
WHERE %s
ORDER BY m.id desc, e.created desc
`

// UpdateDirection updates the direction for the given migration.
func (s *Store) UpdateDirection(ctx context.Context, id int, applyReverse bool) error {
	return s.Store.Exec(ctx, sqlf.Sprintf(updateDirectionQuery, applyReverse, id))
}

const updateDirectionQuery = `
-- source: internal/oobmigration/store.go:UpdateDirection
UPDATE out_of_band_migrations SET apply_reverse = %s WHERE id = %s
`

// UpdateProgress updates the progress for the given migration.
func (s *Store) UpdateProgress(ctx context.Context, id int, progress float64) error {
	return s.updateProgress(ctx, id, progress, time.Now())
}

func (s *Store) updateProgress(ctx context.Context, id int, progress float64, now time.Time) error {
	return s.Store.Exec(ctx, sqlf.Sprintf(updateProgressQuery, progress, now, id, progress))
}

const updateProgressQuery = `
-- source: internal/oobmigration/store.go:UpdateProgress
UPDATE out_of_band_migrations SET progress = %s, last_updated = %s WHERE id = %s AND progress != %s
`

// UpdateMetadata updates the metadata for the given migration.
func (s *Store) UpdateMetadata(ctx context.Context, id int, meta json.RawMessage) error {
	return s.updateMetadata(ctx, id, meta, time.Now())
}

func (s *Store) updateMetadata(ctx context.Context, id int, meta json.RawMessage, now time.Time) error {
	return s.Store.Exec(ctx, sqlf.Sprintf(updateMetadataQuery, meta, now, id, meta))
}

const updateMetadataQuery = `
-- source: internal/oobmigration/store.go:UpdateProgress
UPDATE out_of_band_migrations SET metadata = %s, last_updated = %s WHERE id = %s AND metadata != %s
`

// MaxMigrationErrors is the maximum number of errors we'll track for a single migration before
// pruning older entries.
const MaxMigrationErrors = 100

// AddError associates the given error message with the given migration. While there are more
// than MaxMigrationErrors errors for this, the oldest error entries will be pruned to keep the
// error list relevant and short.
func (s *Store) AddError(ctx context.Context, id int, message string) (err error) {
	return s.addError(ctx, id, message, time.Now())
}

func (s *Store) addError(ctx context.Context, id int, message string, now time.Time) (err error) {
	tx, err := s.Store.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	if err := tx.Exec(ctx, sqlf.Sprintf(addErrorQuery, id, message, now)); err != nil {
		return err
	}

	if err := tx.Exec(ctx, sqlf.Sprintf(addErrorUpdateTimeQuery, now, id)); err != nil {
		return err
	}

	if err := tx.Exec(ctx, sqlf.Sprintf(addErrorPruneQuery, id, MaxMigrationErrors)); err != nil {
		return err
	}

	return nil
}

const addErrorQuery = `
-- source: internal/oobmigration/store.go:AddError
INSERT INTO out_of_band_migrations_errors (migration_id, message, created) VALUES (%s, %s, %s)
`

const addErrorUpdateTimeQuery = `
-- source: internal/oobmigration/store.go:AddError
UPDATE out_of_band_migrations SET last_updated = %s where id = %s
`

const addErrorPruneQuery = `
-- source: internal/oobmigration/store.go:AddError
DELETE FROM out_of_band_migrations_errors WHERE id IN (
	SELECT id FROM out_of_band_migrations_errors WHERE migration_id = %s ORDER BY created DESC OFFSET %s
)
`
