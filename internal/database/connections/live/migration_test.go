package connections

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"github.com/jackc/pgconn"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/definition"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/drift"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/runner"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/store"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var testSchemas = []string{
	"frontend",
	"codeintel",
	"codeinsights",
}

func TestMigrations(t *testing.T) {
	for _, name := range testSchemas {
		schema, ok := getSchema(name)
		if !ok {
			t.Fatalf("missing schema %s", name)
		}

		t.Run(name, func(t *testing.T) {
			testMigrations(t, name, schema)
			testMigrationIdempotency(t, name, schema)
			testDownMigrationsDoNotCreateDrift(t, name, schema)
		})
	}
}

func getSchema(name string) (*schemas.Schema, bool) {
	for _, schema := range schemas.Schemas {
		if schema.Name == name {
			return schema, true
		}
	}

	return nil, false
}

// 🚨 SECURITY: These tables are NOT governed by Postgres RLS protection to isolate
// tenant data.
// This list should only ever contain tables that are system critical, and NOT tenant-specific.
var tablesWithoutTenant = map[string]map[string]struct{}{
	"frontend": {
		"tenants":                  {}, // The tenant table itself, it cannot link to itself.
		"migration_logs":           {}, // Maintained by migrator and not part of Sourcegraph proper.
		"versions":                 {}, // Maintained by migrator and not part of Sourcegraph proper.
		"critical_and_site_config": {}, // Site config is global to the instance so it does not have a tenant.

		// Excluding lsif_* since we are not targetting code-intel initially
		// and they cause issues since its hard to get a table lock with the
		// many long running transactions against them from worker.
		"lsif_configuration_policies":                           {},
		"lsif_configuration_policies_repository_pattern_lookup": {},
		"lsif_dependency_indexing_jobs":                         {},
		"lsif_dependency_repos":                                 {},
		"lsif_dependency_syncing_jobs":                          {},
		"lsif_dirty_repositories":                               {},
		"lsif_index_configuration":                              {},
		"lsif_indexes":                                          {},
		"lsif_last_index_scan":                                  {},
		"lsif_last_retention_scan":                              {},
		"lsif_nearest_uploads":                                  {},
		"lsif_nearest_uploads_links":                            {},
		"lsif_packages":                                         {},
		"lsif_references":                                       {},
		"lsif_retention_configuration":                          {},
		"lsif_uploads":                                          {},
		"lsif_uploads_audit_logs":                               {},
		"lsif_uploads_reference_counts":                         {},
		"lsif_uploads_visible_at_tip":                           {},
		"lsif_uploads_vulnerability_scan":                       {},
	},
	"codeintel": {
		"tenants":        {}, // The tenant table itself, it cannot link to itself.
		"migration_logs": {}, // Maintained by migrator and not part of Sourcegraph proper.
	},
	"codeinsights": {
		"tenants":        {}, // The tenant table itself, it cannot link to itself.
		"migration_logs": {}, // Maintained by migrator and not part of Sourcegraph proper.
	},
}

func testMigrations(t *testing.T, name string, schema *schemas.Schema) {
	t.Helper()

	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := dbtest.NewRawDB(logger, t)
	storeFactory := newStoreFactory(observation.TestContextTB(t))
	migrationRunner := runnerFromDB(logger, storeFactory, db, schema)
	all := schema.Definitions.All()

	t.Run("up", func(t *testing.T) {
		options := runner.Options{
			Operations: []runner.MigrationOperation{
				{
					SchemaName: name,
					Type:       runner.MigrationOperationTypeUpgrade,
				},
			},
		}
		if err := migrationRunner.Run(ctx, options); err != nil {
			t.Fatalf("failed to perform initial upgrade: %s", err)
		}

		t.Run("verify tenant isolation config", func(t *testing.T) {
			// Get the list of all tables
			tables, err := getAllTables(db)
			if err != nil {
				t.Fatalf("Failed to retrieve tables: %v", err)
			}

			for _, table := range tables {
				if _, ok := tablesWithoutTenant[name][table]; ok {
					continue
				}
				hasTenantID, err := tableHasTenantIDColumn(db, table)
				if err != nil {
					t.Errorf("Failed to check tenant_id column for table %s: %v", table, err)
				}
				if !hasTenantID {
					t.Errorf("Table %s does not have a tenant_id column. In the migration that adds it, make sure to include \n\ntenant_id integer REFERENCES tenants(id) ON UPDATE CASCADE ON DELETE CASCADE;\n\n", table)
				}
			}
		})
	})
	t.Run("down", func(t *testing.T) {
		// Run down to the root "squashed commits" migration. For this, we need to select
		// the _last_ nonIdempotent migration in the prefix of the migration definitions.
		// This will ensure that we downgrade _to_ the squashed migrations, but do not try
		// to re-run them on the way back up.

		var target int
		for offset := range len(all) {
			// This is the last definition _or_ the next migration is idempotent
			if offset+1 >= len(all) || !all[offset+1].NonIdempotent {
				target = all[offset].ID
				break
			}
		}
		if target == 0 {
			t.Fatalf("failed to locate last squashed migration definition")
		}

		options := runner.Options{
			Operations: []runner.MigrationOperation{
				{
					SchemaName:     name,
					Type:           runner.MigrationOperationTypeTargetedDown,
					TargetVersions: []int{target},
				},
			},
		}
		if err := migrationRunner.Run(ctx, options); err != nil {
			t.Fatalf("failed to perform downgrade: %s", err)
		}
	})
	t.Run("up again", func(t *testing.T) {
		options := runner.Options{
			Operations: []runner.MigrationOperation{
				{
					SchemaName: name,
					Type:       runner.MigrationOperationTypeUpgrade,
				},
			},
		}
		if err := migrationRunner.Run(ctx, options); err != nil {
			t.Fatalf("failed to perform initial upgrade: %s", err)
		}
	})
}

func testMigrationIdempotency(t *testing.T, name string, schema *schemas.Schema) {
	t.Helper()

	logger := logtest.Scoped(t)
	db := dbtest.NewRawDB(logger, t)
	all := schema.Definitions.All()

	t.Run("idempotent up", func(t *testing.T) {
		for _, definition := range all {
			if err := applyMigration(db, definition, true); err != nil {
				t.Fatalf("failed to perform upgrade of migration %d: %s", definition.ID, err)
			}

			if definition.NonIdempotent {
				// Some migrations are explicitly non-idempotent (squashed migrations)
				// Skip these here
				continue
			}

			if err := applyMigration(db, definition, true); err != nil {
				t.Fatalf("migration %d is not idempotent%s: %s", definition.ID, formatHint(err), err)
			}
		}
	})

	t.Run("idempotent down", func(t *testing.T) {
		for i := len(all) - 1; i >= 0; i-- {
			definition := all[i]

			if err := applyMigration(db, definition, false); err != nil {
				t.Fatalf("failed to perform downgrade of migration %d: %s", definition.ID, err)
			}

			if definition.NonIdempotent {
				// Some migrations are explicitly non-idempotent (squashed migrations)
				// Skip these here
				continue
			}

			if err := applyMigration(db, definition, false); err != nil {
				t.Fatalf("migration %d is not idempotent%s: %s", definition.ID, formatHint(err), err)
			}
		}
	})
}

func testDownMigrationsDoNotCreateDrift(t *testing.T, name string, schema *schemas.Schema) {
	t.Helper()

	logger := logtest.Scoped(t)
	db := dbtest.NewRawDB(logger, t)
	all := schema.Definitions.All()
	store := store.NewWithDB(observation.TestContextTB(t), db, "")

	for _, definition := range all {
		// Capture initial database schema
		expectedDescriptions, err := store.Describe(context.Background())
		if err != nil {
			t.Fatalf("unexpected error describing schema: %s", err)
		}
		expectedDescription := expectedDescriptions["public"]

		// Run query up
		if err := applyMigration(db, definition, true); err != nil {
			t.Fatalf("failed to perform upgrade of migration %d: %s", definition.ID, err)
		}

		if definition.NonIdempotent {
			// Some migrations are explicitly non-idempotent (squashed migrations)
			// Skip these here
			continue
		}

		// Run query down (should restore previous state)
		if err := applyMigration(db, definition, false); err != nil {
			t.Fatalf("failed to perform downgrade of migration %d: %s", definition.ID, err)
		}

		// Describe database schema and check it against initial schema
		descriptions, err := store.Describe(context.Background())
		if err != nil {
			t.Fatalf("unexpected error describing schema: %s", err)
		}
		description := descriptions["public"]

		// Detect drift between previous state (before to up/down) and new state (after)
		if summaries := drift.CompareSchemaDescriptions(name, "", description, expectedDescription); len(summaries) > 0 {
			for _, summary := range summaries {
				statements := "None"
				if s, ok := summary.Statements(); ok {
					statements = strings.Join(s, "\n >")
				}

				urlHint := "None"
				if u, ok := summary.URLHint(); ok {
					urlHint = fmt.Sprintf("Reproduce query as defined at the following URL: \n > %s", u)
				}

				t.Fatalf(
					"\n Drift detected at migration: %d \n Explanation: \n > %s \n Suggested action: \n > %s. \n Suggested query: \n > %s \n Hint: \n > %s\n",
					definition.ID,
					summary.Problem(),
					summary.Solution(),
					statements,
					urlHint,
				)
			}

			t.Fatalf("Detected drift!")
		}

		// Re-run query up to prepare for next round
		if err := applyMigration(db, definition, true); err != nil {
			t.Fatalf("failed to re-perform upgrade of migration %d: %s", definition.ID, err)
		}
	}
}

func formatHint(err error) string {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return ""
	}

	switch pgErr.Code {
	case
		// `undefined_*` error codes
		"42703", "42883", "42P01",
		"42P02", "42704":

		return ` (hint: use "IF EXISTS" in deletion statements)`

	case
		// `duplicate_*` error codes
		"42701", "42P03", "42P04",
		"42723", "42P05", "42P06",
		"42P07", "42712", "42710":

		return ` (hint: use "IF NOT EXISTS"/"CREATE OR REPLACE" in creation statements (e.g., tables, indexes, views, functions), or drop existing objects prior to creating them (e.g., user-defined types, constraints, triggers))`

	}

	return ""
}

// applyMigration applies migrations for testing. In real-world, they run inside of a
// transaction, sp we mimic that in this helper as well.
func applyMigration(db *sql.DB, definition definition.Definition, up bool) (err error) {
	type execer interface {
		Exec(query string, args ...any) (sql.Result, error)
	}
	var queryRunner execer = db

	if !definition.IsCreateIndexConcurrently {
		tx, err := db.BeginTx(context.Background(), &sql.TxOptions{})
		if err != nil {
			return err
		}
		queryRunner = tx
		defer func() {
			if err != nil {
				err = errors.Append(err, tx.Rollback())
			}
			err = tx.Commit()
		}()
	}

	if up {
		if _, err := queryRunner.Exec(definition.UpQuery.Query(sqlf.PostgresBindVar)); err != nil {
			return errors.Wrapf(err, "failed to apply migration %d:\n```\n%s\n```\n", definition.ID, definition.UpQuery.Query(sqlf.PostgresBindVar))
		}
	} else {
		if _, err := queryRunner.Exec(definition.DownQuery.Query(sqlf.PostgresBindVar)); err != nil {
			return errors.Wrapf(err, "failed to apply migration %d:\n```\n%s\n```\n", definition.ID, definition.DownQuery.Query(sqlf.PostgresBindVar))
		}
	}

	return nil
}

func getAllTables(db *sql.DB) ([]string, error) {
	query := `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema='public'
		AND table_type='BASE TABLE'
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tables, nil
}

func tableHasTenantIDColumn(db *sql.DB, tableName string) (bool, error) {
	q := sqlf.Sprintf(`
		SELECT column_name
		FROM information_schema.columns
		WHERE table_name=%s AND column_name='tenant_id'
	`, tableName)

	var columnName string
	err := db.QueryRow(q.Query(sqlf.PostgresBindVar), q.Args()...).Scan(&columnName)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return columnName == "tenant_id", nil
}
