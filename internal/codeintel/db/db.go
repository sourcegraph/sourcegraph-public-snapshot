package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/types"
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
	// Transact returns a Database whose methods operate within the context of a transaction.
	// This method will return an error if the underlying DB cannot be interface upgraded
	// to a TxBeginner.
	Transact(ctx context.Context) (DB, error)

	// Done commits underlying the transaction on a nil error value and performs a rollback
	// otherwise. If an error occurs during commit or rollback of the transaction, the error
	// is added to the resulting error value. If the Database does not wrap a transaction the
	// original error value is returned unchanged.
	Done(err error) error

	// GetUploadByID returns an upload by its identifier and boolean flag indicating its existence.
	GetUploadByID(ctx context.Context, id int) (Upload, bool, error)

	// GetUploadsByRepo returns a list of uploads for a particular repo and the total count of records matching the given conditions.
	GetUploadsByRepo(ctx context.Context, repositoryID int, state, term string, visibleAtTip bool, limit, offset int) ([]Upload, int, error)

	// Enqueue inserts a new upload with a "queued" state and returns its identifier.
	Enqueue(ctx context.Context, commit, root, tracingContext string, repositoryID int, indexerName string) (int, error)

	// Dequeue selects the oldest queued upload and locks it with a transaction. If there is such an upload, the
	// upload is returned along with a JobHandle instance which wraps the transaction. This handle must be closed.
	// If there is no such unlocked upload, a zero-value upload and nil-job handle will be returned along with a
	// false-valued flag.  This method must not be called from within a transaction.
	Dequeue(ctx context.Context) (Upload, JobHandle, bool, error)

	// GetStates returns the states for the uploads with the given identifiers.
	GetStates(ctx context.Context, ids []int) (map[int]string, error)

	// DeleteUploadByID deletes an upload by its identifier. If the upload was visible at the tip of its repository's default branch,
	// the visibility of all uploads for that repository are recalculated. The given function is expected to return the newest commit
	// on the default branch when invoked.
	DeleteUploadByID(ctx context.Context, id int, getTipCommit GetTipCommitFn) (bool, error)

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

	// UpdateDumpsVisibleFromTip recalculates the visible_at_tip flag of all dumps of the given repository.
	UpdateDumpsVisibleFromTip(ctx context.Context, repositoryID int, tipCommit string) (err error)

	// DeleteOverlapapingDumps deletes all completed uploads for the given repository with the same
	// commit, root, and indexer. This is necessary to perform during conversions before changing
	// the state of a processing upload to completed as there is a unique index on these four columns.
	DeleteOverlappingDumps(ctx context.Context, repositoryID int, commit, root, indexer string) error

	// GetPackage returns the dump that provides the package with the given scheme, name, and version and a flag indicating its existence.
	GetPackage(ctx context.Context, scheme, name, version string) (Dump, bool, error)

	// UpdatePackages bulk upserts package data.
	UpdatePackages(ctx context.Context, packages []types.Package) error

	// SameRepoPager returns a ReferencePager for dumps that belong to the given repository and commit and reference the package with the
	// given scheme, name, and version.
	SameRepoPager(ctx context.Context, repositoryID int, commit, scheme, name, version string, limit int) (int, ReferencePager, error)

	// UpdatePackageReferences bulk inserts package reference data.
	UpdatePackageReferences(ctx context.Context, packageReferences []types.PackageReference) error

	// PackageReferencePager returns a ReferencePager for dumps that belong to a remote repository (distinct from the given repository id)
	// and reference the package with the given scheme, name, and version. All resulting dumps are visible at the tip of their repository's
	// default branch.
	PackageReferencePager(ctx context.Context, scheme, name, version string, repositoryID, limit int) (int, ReferencePager, error)

	// HasCommit determines if the given commit is known for the given repository.
	HasCommit(ctx context.Context, repositoryID int, commit string) (bool, error)

	// UpdateCommits upserts commits/parent-commit relations for the given repository ID.
	UpdateCommits(ctx context.Context, repositoryID int, commits map[string][]string) error

	// RepoName returns the name for the repo with the given identifier. This is the only method
	// in this package that touches any table that does not start with `lsif_`.
	RepoName(ctx context.Context, repositoryID int) (string, error)
}

// GetTipCommitFn returns the head commit for the given repository.
type GetTipCommitFn func(repositoryID int) (string, error)

type dbImpl struct {
	db dbutil.DB
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

// query performs QueryContext on the underlying connection.
func (db *dbImpl) query(ctx context.Context, query *sqlf.Query) (*sql.Rows, error) {
	return db.db.QueryContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...)
}

// exec performs a query and throws away the result.
func (db *dbImpl) exec(ctx context.Context, query *sqlf.Query) error {
	rows, err := db.query(ctx, query)
	if err != nil {
		return err
	}
	return rows.Close()
}
