package runner

import (
	"context"
	"fmt"
	"strings"
	"testing"

	mockassert "github.com/derision-test/go-mockgen/testutil/assert"
)

func TestRunnerRun(t *testing.T) {
	overrideSchemas(t)
	ctx := context.Background()

	t.Run("upgrade", func(t *testing.T) {
		store := testStoreWithVersion(10000, false)

		if err := testRunner(store).Run(ctx, Options{
			Up:              true,
			TargetMigration: 10000 + 2,
			SchemaNames:     []string{"well-formed"},
		}); err != nil {
			t.Fatalf("unexpected error running upgrade: %s", err)
		}

		mockassert.CalledN(t, store.UpFunc, 2)
		mockassert.NotCalled(t, store.DownFunc)
	})

	t.Run("downgrade", func(t *testing.T) {
		store := testStoreWithVersion(10003, false)

		if err := testRunner(store).Run(ctx, Options{
			Up:              false,
			TargetMigration: 10003 - 2,
			SchemaNames:     []string{"well-formed"},
		}); err != nil {
			t.Fatalf("unexpected error running downgrade: %s", err)
		}

		mockassert.NotCalled(t, store.UpFunc)
		mockassert.CalledN(t, store.DownFunc, 2)
	})

	t.Run("upgrade error", func(t *testing.T) {
		store := testStoreWithVersion(10000, false)
		store.UpFunc.PushReturn(fmt.Errorf("uh-oh"))

		if err := testRunner(store).Run(ctx, Options{
			Up:              true,
			TargetMigration: 10000 + 1,
			SchemaNames:     []string{"query-error"},
		}); err == nil || !strings.Contains(err.Error(), "uh-oh") {
			t.Fatalf("unexpected error running upgrade. want=%q have=%q", "uh-oh", err)
		}

		mockassert.CalledN(t, store.UpFunc, 1)
		mockassert.NotCalled(t, store.DownFunc)
	})

	t.Run("downgrade error", func(t *testing.T) {
		store := testStoreWithVersion(10001, false)
		store.DownFunc.PushReturn(fmt.Errorf("uh-oh"))

		if err := testRunner(store).Run(ctx, Options{
			Up:              false,
			TargetMigration: 10001 - 1,
			SchemaNames:     []string{"query-error"},
		}); err == nil || !strings.Contains(err.Error(), "uh-oh") {
			t.Fatalf("unexpected error running downgrade. want=%q have=%q", "uh-oh", err)
		}

		mockassert.NotCalled(t, store.UpFunc)
		mockassert.CalledN(t, store.DownFunc, 1)
	})

	t.Run("unknown schema", func(t *testing.T) {
		if err := testRunner(testStoreWithVersion(10000, false)).Run(ctx, Options{
			Up:              true,
			TargetMigration: 10000 + 1,
			SchemaNames:     []string{"unknown"},
		}); err == nil || !strings.Contains(err.Error(), "unknown schema") {
			t.Fatalf("unexpected error running upgrade. want=%q have=%q", "unknown schema", err)
		}
	})

	t.Run("dirty database (pre-lock)", func(t *testing.T) {
		store := testStoreWithVersion(10000, true)

		if err := testRunner(store).Run(ctx, Options{
			Up:              true,
			TargetMigration: 10000 + 1,
			SchemaNames:     []string{"well-formed"},
		}); err == nil || !strings.Contains(err.Error(), "dirty database") {
			t.Fatalf("unexpected error running upgrade. want=%q have=%q", "dirty database", err)
		}
	})

	t.Run("dirty database (post-lock)", func(t *testing.T) {
		store := testStoreWithVersion(10002, true)
		store.VersionFunc.SetDefaultReturn(10001, true, true, nil)

		if err := testRunner(store).Run(ctx, Options{
			Up:              true,
			TargetMigration: 10002 + 0,
			SchemaNames:     []string{"well-formed"},
		}); err == nil || !strings.Contains(err.Error(), "dirty database") {
			t.Fatalf("unexpected error running upgrade. want=%q have=%q", "dirty database", err)
		}
	})

	t.Run("migration contention (concurrent lock)", func(t *testing.T) {
		store := testStoreWithVersion(10000, false)
		store.VersionFunc.SetDefaultReturn(10002, true, true, nil)
		store.TryLockFunc.SetDefaultReturn(false, func(err error) error { return err }, nil)

		if err := testRunner(store).Run(ctx, Options{
			Up:              true,
			TargetMigration: 10000 + 1,
			SchemaNames:     []string{"well-formed"},
		}); err == nil || !strings.Contains(err.Error(), "contention") {
			t.Fatalf("unexpected error running upgrade. want=%q have=%q", "contention", err)
		}
	})

	t.Run("migration contention (version changed)", func(t *testing.T) {
		store := testStoreWithVersion(10002, false)
		store.VersionFunc.PushReturn(10001, false, true, nil)

		if err := testRunner(store).Run(ctx, Options{
			Up:              true,
			TargetMigration: 10002 + 1,
			SchemaNames:     []string{"well-formed"},
		}); err == nil || !strings.Contains(err.Error(), "contention") {
			t.Fatalf("unexpected error running upgrade. want=%q have=%q", "contention", err)
		}
	})
}
