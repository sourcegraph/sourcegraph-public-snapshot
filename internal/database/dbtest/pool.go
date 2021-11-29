package dbtest

import (
	"context"
	"database/sql"
	"fmt"
	"hash/fnv"
	"io"
	"io/fs"
	"net/url"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/internal/database/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

type testDatabasePool struct {
	db *sql.DB
}

const poolSchemaVersion = 1
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
	claimed		bool NOT NULL,
	clean		bool NOT NULL,
	name		text GENERATED ALWAYS AS ('sourcegraph-dbtest-migrated-' || id::text) STORED
);

CREATE TABLE schema_version (
	version int NOT NULL
);

INSERT INTO schema_version (version) VALUES (1);
	
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
	ID            uint64
	MigrationHash uint64
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

func (t *testDatabasePool) GetTemplate(ctx context.Context, u *url.URL, defs ...*dbconn.Database) (_ *TemplateDB, err error) {
	tx, err := t.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	_, err = tx.ExecContext(ctx, "LOCK TABLE template_dbs IN ACCESS EXCLUSIVE MODE")
	if err != nil {
		return nil, err
	}

	hash, err := hashMigrations(defs...)
	if err != nil {
		return nil, err
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
	_, err = t.db.ExecContext(ctx, "CREATE DATABASE"+pq.QuoteIdentifier(tdb.Name))
	if err != nil {
		return nil, errors.Wrap(err, "create template database")
	}

	templateDB, err := dbconn.NewRaw(urlWithDB(u, tdb.Name).String())
	if err != nil {
		return nil, err
	}
	for _, def := range defs {
		done, err := dbconn.DoMigrateDB(templateDB, def)
		if err != nil {
			return nil, err
		}
		defer done()
	}

	return tdb, nil
}

type MigratedDB struct {
	ID       uint64
	Template uint64
	Claimed  bool
	Clean    bool
	Name     string
}

var migratedDBColumns = []*sqlf.Query{
	sqlf.Sprintf("migrated_dbs.id"),
	sqlf.Sprintf("migrated_dbs.template"),
	sqlf.Sprintf("migrated_dbs.claimed"),
	sqlf.Sprintf("migrated_dbs.clean"),
	sqlf.Sprintf("migrated_dbs.name"),
}

const insertMigratedDB = `
INSERT INTO migrated_dbs (template, claimed, clean)
VALUES (%s, %s, %s)
RETURNING %s
`

const getExistingMigratedDB = `
SELECT %s
FROM migrated_dbs
WHERE claimed = false
	AND clean = true
LIMIT 1
`

func (t *testDatabasePool) GetMigratedDB(ctx context.Context, tdb *TemplateDB) (_ *MigratedDB, err error) {
	tx, err := t.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	// Check to see if there is a clean, migrated DB already available
	q := sqlf.Sprintf(
		getExistingMigratedDB,
		sqlf.Join(migratedDBColumns, ","),
	)
	row := tx.QueryRowContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	mdb, err := scanMigratedDB(row)
	if err == nil {
		return mdb, nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	// Insert a new row, returning the generated name
	q = sqlf.Sprintf(
		insertMigratedDB,
		tdb.ID,
		true,
		false,
		sqlf.Join(migratedDBColumns, ","),
	)
	row = tx.QueryRowContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	mdb, err = scanMigratedDB(row)
	if err != nil {
		return nil, err
	}

	// Create the new database outside of the transaction because databases
	// cannot be created in a transaction.
	_, err = t.db.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE %s TEMPLATE %s", pq.QuoteIdentifier(mdb.Name), pq.QuoteIdentifier(tdb.Name)))
	if err != nil {
		return nil, err
	}

	// No need to migrate the new database since it was created from a template
	return mdb, nil
}

const unclaimCleanMigratedDB = `
UPDATE migrated_dbs
SET (claimed, clean) = (false, true)
WHERE id = %s
`

func (t *testDatabasePool) UnclaimCleanMigratedDB(ctx context.Context, mdb *MigratedDB) error {
	q := sqlf.Sprintf(
		unclaimCleanMigratedDB,
		mdb.ID,
	)

	_, err := t.db.ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	return err
}

const uninsertMigratedDB = `
DELETE FROM migrated_dbs
WHERE id = %s
`

func (t *testDatabasePool) DeleteMigratedDB(ctx context.Context, mdb *MigratedDB) error {
	q := sqlf.Sprintf(uninsertMigratedDB, mdb.ID)
	_, err := t.db.ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return err
	}

	_, err = t.db.ExecContext(ctx, "DROP DATABASE "+pq.QuoteIdentifier(mdb.Name))
	return err
}

func hashMigrations(defs ...*dbconn.Database) (uint64, error) {
	hash := fnv.New64()
	for _, def := range defs {
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
	return hash.Sum64(), nil
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

func scanMigratedDB(scanner dbutil.Scanner) (*MigratedDB, error) {
	var t MigratedDB
	err := scanner.Scan(
		&t.ID,
		&t.Template,
		&t.Claimed,
		&t.Clean,
		&t.Name,
	)
	return &t, err
}
