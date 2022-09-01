package codeintel

import (
	"context"
	"testing"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestMigratorRemovesBoundsWithoutData(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := basestore.NewWithHandle(db.Handle())
	driver := &testMigrationDriver{}
	migrator := newMigrator(store, driver, migratorOptions{
		tableName:     "t_test",
		targetVersion: 2,
		batchSize:     200,
		fields: []fieldSpec{
			{name: "a", postgresType: "integer not null", primaryKey: true},
			{name: "b", postgresType: "integer not null", readOnly: true},
			{name: "c", postgresType: "integer not null"},
		},
	})

	assertProgress := func(expectedProgress float64, applyReverse bool) {
		if progress, err := migrator.Progress(context.Background(), applyReverse); err != nil {
			t.Fatalf("unexpected error querying progress: %s", err)
		} else if progress != expectedProgress {
			t.Errorf("unexpected progress. want=%.2f have=%.2f", expectedProgress, progress)
		}
	}

	if err := store.Exec(context.Background(), sqlf.Sprintf(`
		CREATE TABLE t_test (
			dump_id        integer not null,
			a              integer not null,
			b              integer not null,
			c              integer not null,
			schema_version integer not null,
			primary key (dump_id, a)
		)
	`)); err != nil {
		t.Fatalf("unexpected error creating data table: %s", err)
	}

	if err := store.Exec(context.Background(), sqlf.Sprintf(`
		CREATE TABLE t_test_schema_versions (
				dump_id            integer primary key not null,
				min_schema_version integer not null,
				max_schema_version integer not null
		)
	`)); err != nil {
		t.Fatalf("unexpected error creating schema version table: %s", err)
	}

	n := 600

	for i := 0; i < n; i++ {
		// 33% id=42, 33% id=43, 33% id=44
		dumpID := 42 + i/(n/3)

		if err := store.Exec(context.Background(), sqlf.Sprintf(
			"INSERT INTO t_test (dump_id, a, b, c, schema_version) VALUES (%s, %s, %s, %s, 1)",
			dumpID,
			i,
			i*10,
			i*100,
		)); err != nil {
			t.Fatalf("unexpected error inserting data row: %s", err)
		}
	}

	// 42 is missing; 45 is extra
	for _, dumpID := range []int{43, 44, 45} {
		if err := store.Exec(context.Background(), sqlf.Sprintf(
			"INSERT INTO t_test_schema_versions (dump_id, min_schema_version, max_schema_version) VALUES (%s, 1, 1)",
			dumpID,
		)); err != nil {
			t.Fatalf("unexpected error inserting schema version row: %s", err)
		}
	}

	assertProgress(0, false)

	// process dump 43 (updates bounds)
	if err := migrator.Up(context.Background()); err != nil {
		t.Fatalf("unexpected error performing up migration: %s", err)
	}
	assertProgress(1.0/3.0, false)

	// process dump 44 (updates bounds)
	if err := migrator.Up(context.Background()); err != nil {
		t.Fatalf("unexpected error performing up migration: %s", err)
	}
	assertProgress(2.0/3.0, false)

	// process dump 45 (deletes schema version record with no data)
	if err := migrator.Up(context.Background()); err != nil {
		t.Fatalf("unexpected error performing up migration: %s", err)
	}
	assertProgress(1.0, false)

	// reverse migration of first of remaining two dumps
	if err := migrator.Down(context.Background()); err != nil {
		t.Fatalf("unexpected error performing down migration: %s", err)
	}
	assertProgress(0.5, true)

	// reverse migration of second of remaining two dumps
	if err := migrator.Down(context.Background()); err != nil {
		t.Fatalf("unexpected error performing down migration: %s", err)
	}
	assertProgress(0.0, true)
}

type testMigrationDriver struct{}

func (m *testMigrationDriver) ID() int                 { return 10 }
func (m *testMigrationDriver) Interval() time.Duration { return time.Second }

func (m *testMigrationDriver) MigrateRowUp(scanner scanner) ([]any, error) {
	var a, b, c int
	if err := scanner.Scan(&a, &b, &c); err != nil {
		return nil, err
	}

	return []any{a, b + c}, nil
}

func (m *testMigrationDriver) MigrateRowDown(scanner scanner) ([]any, error) {
	var a, b, c int
	if err := scanner.Scan(&a, &b, &c); err != nil {
		return nil, err
	}

	return []any{a, b - c}, nil
}
