package runner

import (
	"context"
	"fmt"
	"io/fs"
	"testing"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/definition"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/runner/testdata"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
)

func makeTestSchemas(t *testing.T) []*schemas.Schema {
	return []*schemas.Schema{
		makeTestSchema(t, "well-formed"),
		makeTestSchema(t, "query-error"),
		makeTestSchema(t, "concurrent-index"),
	}
}

func makeTestSchema(t *testing.T, name string) *schemas.Schema {
	fsys, err := fs.Sub(testdata.Content, name)
	if err != nil {
		t.Fatalf("malformed migration definitions %q: %s", name, err)
	}

	definitions, err := definition.ReadDefinitions(fsys, name)
	if err != nil {
		t.Fatalf("malformed migration definitions %q: %s", name, err)
	}

	return &schemas.Schema{
		Name:                name,
		MigrationsTableName: fmt.Sprintf("%s_migrations_table", name),
		Definitions:         definitions,
	}
}

func overrideSchemas(t *testing.T) {
	liveSchemas := schemas.Schemas
	schemas.Schemas = makeTestSchemas(t)
	t.Cleanup(func() { schemas.Schemas = liveSchemas })
}

func testStoreWithVersion(version int, dirty bool) *MockStore {
	migrationHook := func(ctx context.Context, definition definition.Definition) error {
		version = definition.ID
		return nil
	}

	store := NewMockStore()
	store.TransactFunc.SetDefaultReturn(store, nil)
	store.DoneFunc.SetDefaultHook(func(err error) error { return err })
	store.TryLockFunc.SetDefaultReturn(true, func(err error) error { return err }, nil)
	store.UpFunc.SetDefaultHook(migrationHook)
	store.DownFunc.SetDefaultHook(migrationHook)
	store.WithMigrationLogFunc.SetDefaultHook(func(_ context.Context, _ definition.Definition, _ bool, f func() error) error { return f() })
	store.VersionsFunc.SetDefaultHook(func(ctx context.Context) ([]int, []int, []int, error) {
		if dirty {
			return nil, nil, []int{version}, nil
		}
		if version == 0 {
			return nil, nil, nil, nil
		}

		base := 10001
		ids := make([]int, 0, 4)
		for v := base; v <= version; v++ {
			ids = append(ids, v)
		}

		return ids, nil, nil, nil
	})

	return store
}

func makeTestRunner(t *testing.T, store Store) *Runner {
	logger := logtest.Scoped(t)
	testSchemas := makeTestSchemas(t)
	storeFactories := make(map[string]StoreFactory, len(testSchemas))

	for _, testSchema := range testSchemas {
		storeFactories[testSchema.Name] = func(ctx context.Context) (Store, error) {
			return store, nil
		}
	}

	return NewRunner(logger, storeFactories)
}
