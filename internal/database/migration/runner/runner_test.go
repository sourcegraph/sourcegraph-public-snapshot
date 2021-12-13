package runner

import (
	"context"
	"fmt"
	"io/fs"
	"strings"
	"testing"

	"github.com/cockroachdb/errors"
	mockassert "github.com/derision-test/go-mockgen/testutil/assert"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/definition"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/runner/testdata"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
)

var testSchemas = []*schemas.Schema{
	makeTestSchema("well-formed"),
	makeTestSchema("query-error"),
}

func TestRunnerRun(t *testing.T) {
	overrideSchemas(t)
	ctx := context.Background()

	t.Run("upgrade", func(t *testing.T) {
		store := testStoreWithVersion(10000)

		if err := testRunner(store).Run(ctx, Options{
			Up:            true,
			NumMigrations: 2,
			SchemaNames:   []string{"well-formed"},
		}); err != nil {
			t.Fatalf("unexpected error running upgrade: %s", err)
		}

		mockassert.CalledN(t, store.UpFunc, 2)
		mockassert.NotCalled(t, store.DownFunc)
	})

	t.Run("downgrade", func(t *testing.T) {
		store := testStoreWithVersion(10003)

		if err := testRunner(store).Run(ctx, Options{
			Up:            false,
			NumMigrations: 2,
			SchemaNames:   []string{"well-formed"},
		}); err != nil {
			t.Fatalf("unexpected error running downgrade: %s", err)
		}

		mockassert.NotCalled(t, store.UpFunc)
		mockassert.CalledN(t, store.DownFunc, 2)
	})

	t.Run("upgrade error", func(t *testing.T) {
		store := testStoreWithVersion(10000)
		store.UpFunc.PushReturn(fmt.Errorf("uh-oh"))

		if err := testRunner(store).Run(ctx, Options{
			Up:            true,
			NumMigrations: 1,
			SchemaNames:   []string{"query-error"},
		}); err == nil || !strings.Contains(err.Error(), "uh-oh") {
			t.Fatalf("unexpected error running upgrade. want=%q have=%q", "uh-oh", err)
		}

		mockassert.CalledN(t, store.UpFunc, 1)
		mockassert.NotCalled(t, store.DownFunc)
	})

	t.Run("downgrade error", func(t *testing.T) {
		store := testStoreWithVersion(10001)
		store.DownFunc.PushReturn(fmt.Errorf("uh-oh"))

		if err := testRunner(store).Run(ctx, Options{
			Up:            false,
			NumMigrations: 1,
			SchemaNames:   []string{"query-error"},
		}); err == nil || !strings.Contains(err.Error(), "uh-oh") {
			t.Fatalf("unexpected error running downgrade. want=%q have=%q", "uh-oh", err)
		}

		mockassert.NotCalled(t, store.UpFunc)
		mockassert.CalledN(t, store.DownFunc, 1)
	})

	t.Run("unknown schema", func(t *testing.T) {
		if err := testRunner(testStoreWithVersion(10000)).Run(ctx, Options{
			Up:            true,
			NumMigrations: 1,
			SchemaNames:   []string{"unknown"},
		}); err == nil || !strings.Contains(err.Error(), "unknown schema") {
			t.Fatalf("unexpected error running upgrade. want=%q have=%q", "unknown schema", err)
		}
	})

	t.Run("checks dirty database on startup", func(t *testing.T) {
		store := testStoreWithVersion(10000)
		store.VersionFunc.SetDefaultReturn(10002, true, true, nil)

		if err := testRunner(store).Run(ctx, Options{
			Up:            true,
			NumMigrations: 1,
			SchemaNames:   []string{"well-formed"},
		}); err == nil || !strings.Contains(err.Error(), "dirty database") {
			t.Fatalf("unexpected error running upgrade. want=%q have=%q", "dirty database", err)
		}
	})
}

func TestRunnerValidate(t *testing.T) {
	overrideSchemas(t)
	ctx := context.Background()

	t.Run("very old schema", func(t *testing.T) {
		store := testStoreWithVersion(250)

		outOfDateError := new(SchemaOutOfDateError)
		if err := testRunner(store).Validate(ctx, "well-formed"); !errors.As(err, &outOfDateError) || outOfDateError.expectedVersion != 10003 || outOfDateError.currentVersion != 250 {
			t.Fatalf("unexpected error running validation. want=(unexpected version; expected 10003, currently 250) have=%s", err)
		}
	})

	t.Run("old schema", func(t *testing.T) {
		store := testStoreWithVersion(10001)

		outOfDateError := new(SchemaOutOfDateError)
		if err := testRunner(store).Validate(ctx, "well-formed"); !errors.As(err, &outOfDateError) || outOfDateError.expectedVersion != 10003 || outOfDateError.currentVersion != 10001 {
			t.Fatalf("unexpected error running validation. want=(unexpected version; expected 10003, currently 10001) have=%s", err)
		}
	})

	t.Run("correct version", func(t *testing.T) {
		store := testStoreWithVersion(10003)

		if err := testRunner(store).Validate(ctx, "well-formed"); err != nil {
			t.Fatalf("unexpected error running validation: %s", err)
		}
	})

	t.Run("future schema", func(t *testing.T) {
		store := testStoreWithVersion(10004)

		if err := testRunner(store).Validate(ctx, "well-formed"); err != nil {
			t.Fatalf("unexpected error running validation: %s", err)
		}
	})

	t.Run("distant future schema", func(t *testing.T) {
		store := testStoreWithVersion(50010)

		if err := testRunner(store).Validate(ctx, "well-formed"); err != nil {
			t.Fatalf("unexpected error running validation: %s", err)
		}
	})
}

func makeTestSchema(name string) *schemas.Schema {
	fs, err := fs.Sub(testdata.Content, name)
	if err != nil {
		panic(fmt.Sprintf("malformed migration definitions %q: %s", name, err))
	}

	definitions, err := definition.ReadDefinitions(fs)
	if err != nil {
		panic(fmt.Sprintf("malformed migration definitions %q: %s", name, err))
	}

	return &schemas.Schema{
		Name:                name,
		MigrationsTableName: fmt.Sprintf("%s_migrations_table", name),
		FS:                  fs,
		Definitions:         definitions,
	}
}

func overrideSchemas(t *testing.T) {
	liveSchemas := schemas.Schemas
	schemas.Schemas = testSchemas
	t.Cleanup(func() { schemas.Schemas = liveSchemas })
}

func testStoreWithVersion(version int) *MockStore {
	migrationHook := func(ctx context.Context, definition definition.Definition) error {
		version = definition.ID
		return nil
	}

	store := NewMockStore()
	store.LockFunc.SetDefaultReturn(true, func(err error) error { return err }, nil)
	store.UpFunc.SetDefaultHook(migrationHook)
	store.DownFunc.SetDefaultHook(migrationHook)
	store.VersionFunc.SetDefaultHook(func(ctx context.Context) (int, bool, bool, error) {
		return version, false, true, nil
	})

	return store
}

func testRunner(store Store) *Runner {
	storeFactories := make(map[string]StoreFactory, len(testSchemas))

	for _, testSchema := range testSchemas {
		storeFactories[testSchema.Name] = func(ctx context.Context) (Store, error) {
			return store, nil
		}
	}

	return NewRunner(storeFactories)
}
