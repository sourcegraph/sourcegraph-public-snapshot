package dbtest

import (
	"context"
	"database/sql"
	"hash/fnv"
	"io"
	"io/fs"
	"net/url"
	"time"

	"github.com/jackc/pgconn"
	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"

	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/test"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// testDatabasePool handles creating and reusing migrated database instances
type testDatabasePool struct {
	*sql.DB
}

func newTestDatabasePool(db *sql.DB) *testDatabasePool {
	return &testDatabasePool{
		DB: db,
	}
}

const poolSchemaVersion = 2
const poolSchema = `
BEGIN;

CREATE TABLE template_dbs (
	id				bigserial PRIMARY KEY,
	migration_hash  bigint NOT NULL,
	name			text GENERATED ALWAYS AS ('sourcegraph-dbtest-template-' || id::text) STORED,
	created_at		timestamptz DEFAULT now(),
	last_used_at	timestamptz DEFAULT now()
);

CREATE TABLE migrated_dbs (
	id			bigserial PRIMARY KEY,
	template	bigint NOT NULL REFERENCES template_dbs(id) ON DELETE RESTRICT, -- restrict to avoid dangling dbs
	available	bool NOT NULL,
	name		text GENERATED ALWAYS AS ('sourcegraph-dbtest-migrated-' || id::text) STORED
);

CREATE TABLE schema_version (
	version int NOT NULL
);

INSERT INTO schema_version (version) VALUES (2);

COMMIT;
`

func poolSchemaUpToDate(db *sql.DB) bool {
	row := db.QueryRow("SELECT version FROM schema_version")
	var v int
	err := row.Scan(&v)
	if err != nil {
		return false
	}
	return v == poolSchemaVersion
}

func migratePoolDB(db *sql.DB) error {
	_, err := db.Exec(poolSchema)
	return err
}

type TemplateDB struct {
	ID            int64
	MigrationHash int64
	Name          string
	CreatedAt     time.Time
	LastUsedAt    time.Time
}

var templateDBColumns = []*sqlf.Query{
	sqlf.Sprintf("template_dbs.id"),
	sqlf.Sprintf("template_dbs.migration_hash"),
	sqlf.Sprintf("template_dbs.name"),
	sqlf.Sprintf("template_dbs.created_at"),
	sqlf.Sprintf("template_dbs.last_used_at"),
}

func scanTemplateDB(scanner dbutil.Scanner) (*TemplateDB, error) {
	var t TemplateDB
	err := scanner.Scan(
		&t.ID,
		&t.MigrationHash,
		&t.Name,
		&t.CreatedAt,
		&t.LastUsedAt,
	)
	return &t, err
}

func scanTemplateDBs(rows *sql.Rows) ([]*TemplateDB, error) {
	var tdbs []*TemplateDB
	for rows.Next() {
		tdb, err := scanTemplateDB(rows)
		if err != nil {
			return nil, err
		}
		tdbs = append(tdbs, tdb)
	}
	return tdbs, nil
}

const getTemplateDB = `
UPDATE template_dbs
SET last_used_at = now()
WHERE migration_hash = %s
RETURNING %s
`

const insertTemplateDB = `
INSERT INTO template_dbs (migration_hash)
VALUES (%s)
RETURNING %s
`

// GetTemplate will return a template database that has been migrated with the given migrations.
// The given migrations are hashed and used to identify template databases that have already been
// migrated. If no template database exists with the same hash as the given migrations, a new template
// database is created and the migrations are run.
func (t *testDatabasePool) GetTemplate(ctx context.Context, u *url.URL, schemas ...*schemas.Schema) (_ *TemplateDB, err error) {
	// Create a transaction so the exclusive lock is dropped at the end of this function.
	tx, err := t.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
		err = tx.Commit()
	}()

	// Create an exclusive lock because we want exactly one template database per hash,
	// and that's difficult to guarantee _and_ guarantee that we don't create the row
	// until the template database is created and fully migrated.
	_, err = tx.ExecContext(ctx, "LOCK TABLE template_dbs IN ACCESS EXCLUSIVE MODE")
	if err != nil {
		return nil, errors.Wrap(err, "lock template_dbs")
	}

	hash, err := hashSchema(schemas...)
	if err != nil {
		return nil, errors.Wrap(err, "hash schemas")
	}

	// Check if the template database has already been created, and return that if it has
	q := sqlf.Sprintf(
		getTemplateDB,
		hash,
		sqlf.Join(templateDBColumns, ","),
	)
	row := tx.QueryRowContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	tdb, err := scanTemplateDB(row)
	if err == nil {
		return tdb, nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, errors.Wrap(err, "check if template exists")
	}

	// If the template database has not been created, insert the row to get the
	// generated name, then create the template database below.
	q = sqlf.Sprintf(
		insertTemplateDB,
		hash,
		sqlf.Join(templateDBColumns, ","),
	)
	row = tx.QueryRowContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	tdb, err = scanTemplateDB(row)
	if err != nil {
		return nil, errors.Wrap(err, "insert template row")
	}

	// Create the database outside the transaction (use t.db) because databases
	// cannot be created inside a transaciton. This is safe because the whole
	// template_dbs table is locked in the transaction above, so this
	// will never happen concurrently.
	if err := createDB(ctx, t.DB, tdb.Name, ""); err != nil {
		return nil, errors.Wrap(err, "create template database")
	}

	db, err := connections.NewTestDB(urlWithDB(u, tdb.Name).String(), schemas...)
	if err != nil {
		return nil, errors.Wrap(err, "migrate template DB")
	}
	if err := db.Close(); err != nil {
		return nil, errors.Wrap(err, "close template DB")
	}

	return tdb, nil
}

const lockTemplateDBQuery = `
SELECT id
FROM migrated_dbs
WHERE id = %s
FOR UPDATE
`

const deleteTemplateDBQuery = `
DELETE FROM migrated_dbs
WHERE id = %s
`

func deleteTemplateDB(ctx context.Context, db *sql.DB, tx *sql.Tx, tdb *TemplateDB) (err error) {
	// Lock the row for delete
	q := sqlf.Sprintf(lockTemplateDBQuery, tdb.ID)
	_, err = tx.ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return errors.Wrap(err, "lock template db row")
	}

	// Delete the database outside of the transaction (dbs can't be created or removed
	// within a transaction)
	_, err = db.ExecContext(ctx, "DROP DATABASE "+pq.QuoteIdentifier(tdb.Name))
	if err != nil {
		return errors.Wrap(err, "drop template db")
	}

	// Remove the row in the transaction that locked it
	q = sqlf.Sprintf(deleteTemplateDBQuery, tdb.ID)
	_, err = tx.ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return errors.Wrap(err, "remove template db row")
	}
	return nil
}

type MigratedDB struct {
	ID        int64
	Template  int64
	Available bool
	Name      string
}

var migratedDBColumns = []*sqlf.Query{
	sqlf.Sprintf("migrated_dbs.id"),
	sqlf.Sprintf("migrated_dbs.template"),
	sqlf.Sprintf("migrated_dbs.available"),
	sqlf.Sprintf("migrated_dbs.name"),
}

func scanMigratedDB(scanner dbutil.Scanner) (*MigratedDB, error) {
	var t MigratedDB
	err := scanner.Scan(
		&t.ID,
		&t.Template,
		&t.Available,
		&t.Name,
	)
	return &t, err
}

func scanMigratedDBs(rows *sql.Rows) ([]*MigratedDB, error) {
	var mdbs []*MigratedDB
	for rows.Next() {
		mdb, err := scanMigratedDB(rows)
		if err != nil {
			return nil, err
		}
		mdbs = append(mdbs, mdb)
	}
	return mdbs, nil
}

const insertMigratedDB = `
INSERT INTO migrated_dbs (template, available)
VALUES (%s, %s)
RETURNING %s
`

const getExistingMigratedDB = `
UPDATE migrated_dbs
SET available = false
WHERE id = (
	SELECT id
	FROM migrated_dbs
	WHERE available = true
	LIMIT 1
	FOR UPDATE
)
RETURNING %s
`

// GetMigratedDB returns a clean, available, migrated db that is cloned from the given templated db. If an available,
// clean database already exists for the given template, that is made unavavailable and returned. If it does not, a new
// database is created from the given template and returned.
func (t *testDatabasePool) GetMigratedDB(ctx context.Context, reuse bool, tdb *TemplateDB) (_ *MigratedDB, err error) {
	// Run this in a transaction so if creating the database
	// fails, creating the row is rolled back
	tx, err := t.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
		err = tx.Commit()
	}()

	// Only reuse a database if the caller says it's okay. Even a "clean" database that
	// has had all transactions rolled back will have updated sequences, and some tests
	// might depend on IDs starting at 1 (even though they probably shouldn't).
	if reuse {
		// Check to see if there is a clean, migrated DB already available
		q := sqlf.Sprintf(getExistingMigratedDB, sqlf.Join(migratedDBColumns, ","))
		row := tx.QueryRowContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
		mdb, err := scanMigratedDB(row)
		if err == nil {
			return mdb, nil
		} else if !errors.Is(err, sql.ErrNoRows) {
			return nil, errors.Wrap(err, "get existing migrated db")
		}
	}

	// Insert a new row, returning the generated name
	q := sqlf.Sprintf(
		insertMigratedDB,
		tdb.ID,
		false,
		sqlf.Join(migratedDBColumns, ","),
	)
	row := tx.QueryRowContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	mdb, err := scanMigratedDB(row)
	if err != nil {
		return nil, errors.Wrap(err, "insert new migrated db")
	}

	// Create the new database outside of the transaction because databases cannot be created in a transaction.
	if err := createDB(ctx, t.DB, mdb.Name, tdb.Name); err != nil {
		return nil, errors.Wrap(err, "create migrated db")
	}

	// No need to migrate the new database since it was created from a template
	return mdb, nil
}

const returnCleanMigratedDB = `
UPDATE migrated_dbs
SET available = true
WHERE id = %s
`

// PutMigratedDB marks a clean database as available, allowing it to be returned by a
// call to GetMigratedDB. A migrated db should never be made available if it was written to, and should
// be deleted instead. This should really only be called if the database was only used in a transaction
// and that transaction was rolled back (as in NewFastTx).
func (t *testDatabasePool) PutMigratedDB(ctx context.Context, mdb *MigratedDB) error {
	q := sqlf.Sprintf(returnCleanMigratedDB, mdb.ID)
	_, err := t.DB.ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return errors.Wrap(err, "make db available")
	}
	return nil
}

const lockMigratedDBQuery = `
SELECT id
FROM migrated_dbs
WHERE id = %s
FOR UPDATE
`

const deleteMigratedDBQuery = `
DELETE FROM migrated_dbs
WHERE id = %s
`

// DeleteMigratedDB deletes a database and untracks it in migrated_dbs. This should
// only be called by the caller who called GetMigratedDB
func (t *testDatabasePool) DeleteMigratedDB(ctx context.Context, mdb *MigratedDB) (err error) {
	tx, err := t.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
		err = tx.Commit()
	}()

	return deleteMigratedDB(ctx, t.DB, tx, mdb)
}

func deleteMigratedDB(ctx context.Context, db *sql.DB, tx *sql.Tx, mdb *MigratedDB) (err error) {
	// Lock the row for delete
	q := sqlf.Sprintf(lockMigratedDBQuery, mdb.ID)
	_, err = tx.ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return errors.Wrap(err, "lock migrated db row")
	}

	// Delete the database outside of the transaction (dbs can't be created or removed
	// within a transaction)
	_, err = db.ExecContext(ctx, "DROP DATABASE "+pq.QuoteIdentifier(mdb.Name))
	if err != nil {
		return errors.Wrap(err, "drop migrated db")
	}

	// Remove the row in the transaction that locked it
	q = sqlf.Sprintf(deleteMigratedDBQuery, mdb.ID)
	_, err = tx.ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return errors.Wrap(err, "delete migrated db row")
	}
	return nil
}

const listOldTemplateDBs = `
SELECT %s
FROM template_dbs
WHERE
	migration_hash != %s
	AND last_used_at < NOW() - INTERVAL '1 day'
FOR UPDATE
`

func (t *testDatabasePool) CleanUpOldDBs(ctx context.Context, except ...*schemas.Schema) (err error) {
	hash, err := hashSchema(except...)
	if err != nil {
		return errors.Wrap(err, "hash schema")
	}

	tx, err := t.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
		err = tx.Commit()
	}()

	// List any old template databases that don't have the same
	// hash as the given database definitions
	q := sqlf.Sprintf(listOldTemplateDBs, sqlf.Join(templateDBColumns, ","), hash)
	rows, err := tx.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return errors.Wrap(err, "list old template DBs")
	}
	defer rows.Close()

	oldTDBs, err := scanTemplateDBs(rows)
	if err != nil {
		return err
	}

	var errs *errors.MultiError
	for _, tdb := range oldTDBs {
		mdbs, err := listMigratedDBs(ctx, tx, tdb.ID)
		if err != nil {
			errs = errors.Append(errs, err)
			continue
		}

		for _, mdb := range mdbs {
			err = deleteMigratedDB(ctx, t.DB, tx, mdb)
			if err != nil {
				errs = errors.Append(errs, err)
				continue
			}
		}

		err = deleteTemplateDB(ctx, t.DB, tx, tdb)
		if err != nil {
			errs = errors.Append(errs, err)
		}
	}

	return errs.ErrorOrNil()
}

const listMigratedDBsQuery = `
SELECT %s
FROM migrated_dbs
WHERE template = %s
FOR UPDATE
`

func listMigratedDBs(ctx context.Context, tx *sql.Tx, template int64) ([]*MigratedDB, error) {
	q := sqlf.Sprintf(listMigratedDBsQuery, sqlf.Join(migratedDBColumns, ","), template)
	rows, err := tx.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return nil, errors.Wrap(err, "list migrated dbs")
	}
	defer rows.Close()
	return scanMigratedDBs(rows)
}

// createDB creates a db that should not exist. If it does already exist, it deletes
// it and recreates it.
func createDB(ctx context.Context, db *sql.DB, name string, template string) error {
	var withTemplate string
	if template != "" {
		withTemplate = " TEMPLATE " + pq.QuoteIdentifier(template)
	}
	_, err := db.Exec("CREATE DATABASE " + pq.QuoteIdentifier(name) + withTemplate)
	if err == nil {
		return nil
	}

	// If the database already exists, handle it "gracefully" by dropping
	// it and recreating it.
	var e pgconn.PgError
	if errors.As(err, e) && e.Code == "42P04" { // code for database already exists
		_, err := db.Exec("DROP DATABASE " + pq.QuoteIdentifier(name))
		if err != nil {
			return errors.Wrapf(err, "dropping database %q after failed create", name)
		}

		_, err = db.Exec("CREATE DATABASE " + pq.QuoteIdentifier(name) + withTemplate)
		if err != nil {
			return errors.Wrapf(err, "creating database %q after drop", name)
		}
		return nil
	}

	return errors.Wrapf(err, "create database %q", name)
}

// hashSchema deterministically hashes all the migrations in the given
// schema description. This is used to determine whether a new template
// database should be created for the given set of migrations.
func hashSchema(schemas ...*schemas.Schema) (int64, error) {
	hash := fnv.New64()
	for _, def := range schemas {
		root, err := def.FS.Open(".")
		if err != nil {
			return 0, err
		}

		rootDir, ok := root.(fs.ReadDirFile)
		if !ok {
			return 0, errors.New("root of migration filesystem is not a directory")
		}

		dirEntries, err := rootDir.ReadDir(0)
		if err != nil {
			return 0, err
		}

		for _, entry := range dirEntries {
			f, err := def.FS.Open(entry.Name())
			if err != nil {
				return 0, err
			}
			_, err = io.Copy(hash, f)
			if err != nil {
				return 0, err
			}
		}
	}
	return int64(hash.Sum64()), nil
}
