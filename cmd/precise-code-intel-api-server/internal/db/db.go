package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/internal/db/dbutil"
)

// DB is the interface to Postgres that deals with LSIF-specific tables.
//
//   - lsif_commits
//   - lsif_packages
//   - lsif_references
//   - lsif_uploads
//
// These tables are kept separate from the remainder of Sourcegraph tablespace.
type DB interface {
	// GetUploadByID returns an upload by its identifier and boolean flag indicating its existence.
	GetUploadByID(ctx context.Context, id int) (Upload, bool, error)

	// GetUploadsByRepo returns a list of uploads for a particular repo and the total count of records matching the given conditions.
	GetUploadsByRepo(ctx context.Context, repositoryID int, state, term string, visibleAtTip bool, limit, offset int) ([]Upload, int, error)

	// Enqueue inserts a new upload with a "queued" state, returning its identifier and a TxCloser that must be closed to commit the transaction.
	Enqueue(ctx context.Context, commit, root, tracingContext string, repositoryID int, indexerName string) (int, TxCloser, error)

	// GetStates returns the states for the uploads with the given identifiers.
	GetStates(ctx context.Context, ids []int) (map[int]string, error)

	// DeleteUploadByID deletes an upload by its identifier. If the upload was visible at the tip of its repository's default branch,
	// the visibility of all uploads for that repository are recalculated. The given function is expected to return the newest commit
	// on the default branch when invoked.
	DeleteUploadByID(ctx context.Context, id int, getTipCommit func(repositoryID int) (string, error)) (bool, error)

	// ResetStalled moves all unlocked uploads processing for more than `StalledUploadMaxAge` back to the queued state.
	// This method returns a list of updated upload identifiers.
	ResetStalled(ctx context.Context, now time.Time) ([]int, error)

	// GetDumpByID returns a dump by its identifier and boolean flag indicating its existence.
	GetDumpByID(ctx context.Context, id int) (Dump, bool, error)

	// FindClosestDumps returns the set of dumps that can most accurately answer queries for the given repository, commit, and file.
	FindClosestDumps(ctx context.Context, repositoryID int, commit, file string) ([]Dump, error)

	// DeleteOldestDump deletes the oldest dump that is not currently visible at the tip of its repository's default branch.
	// This method returns the deleted dump's identifier and a flag indicating its (previous) existence.
	DeleteOldestDump(ctx context.Context) (int, bool, error)

	// GetPackage returns the dump that provides the package with the given scheme, name, and version and a flag indicating its existence.
	GetPackage(ctx context.Context, scheme, name, version string) (Dump, bool, error)

	// SameRepoPager returns a ReferencePager for dumps that belong to the given repository and commit and reference the package with the
	// given scheme, name, and version.
	SameRepoPager(ctx context.Context, repositoryID int, commit, scheme, name, version string, limit int) (int, ReferencePager, error)

	// PackageReferencePager returns a ReferencePager for dumps that belong to a remote repository (distinct from the given repository id)
	// and reference the package with the given scheme, name, and version. All resulting dumps are visible at the tip of their repository's
	// default branch.
	PackageReferencePager(ctx context.Context, scheme, name, version string, repositoryID, limit int) (int, ReferencePager, error)
}

type dbImpl struct {
	db *sql.DB
}

var _ DB = &dbImpl{}

// New creates a new instance of DB connected to the given Postgres DSN.
func New(postgresDSN string) (DB, error) {
	db, err := dbutil.NewDB(postgresDSN, "precise-code-intel-api-server")
	if err != nil {
		return nil, err
	}

	return &dbImpl{db: db}, nil
}

// query performs Query on the underlying connection.
func (db *dbImpl) query(ctx context.Context, query *sqlf.Query) (*sql.Rows, error) {
	return db.db.QueryContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...)
}

// queryRow performs QueryRow on the underlying connection.
func (db *dbImpl) queryRow(ctx context.Context, query *sqlf.Query) *sql.Row {
	return db.db.QueryRowContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...)
}

// beginTx performs BeginTx on the underlying connection and wraps the transaction.
func (db *dbImpl) beginTx(ctx context.Context) (*transactionWrapper, error) {
	tx, err := db.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &transactionWrapper{tx}, nil
}
