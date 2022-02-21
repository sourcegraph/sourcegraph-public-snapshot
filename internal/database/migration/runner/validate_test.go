package runner

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestValidate(t *testing.T) {
	overrideSchemas(t)
	ctx := context.Background()

	t.Run("empty", func(t *testing.T) {
		store := testStoreWithVersion(0, false)

		e := new(SchemaOutOfDateError)
		if err := makeTestRunner(t, store).Validate(ctx, "well-formed"); err == nil {
			t.Fatalf("expected an error")
		} else if !errors.As(err, &e) {
			t.Fatalf("expected error. want schema out of date error, have=%s", err)
		}

		expectedMissingVersions := []int{
			10001,
			10002,
			10003,
			10004,
		}
		if diff := cmp.Diff(expectedMissingVersions, e.missingVersions); diff != "" {
			t.Errorf("unexpected missing versions (-want +got):\n%s", diff)
		}
	})

	t.Run("partially applied", func(t *testing.T) {
		store := testStoreWithVersion(10002, false)

		e := new(SchemaOutOfDateError)
		if err := makeTestRunner(t, store).Validate(ctx, "well-formed"); err == nil {
			t.Fatalf("expected an error")
		} else if !errors.As(err, &e) {
			t.Fatalf("expected error. want schema out of date error, have=%s", err)
		}

		expectedMissingVersions := []int{
			10003,
			10004,
		}
		if diff := cmp.Diff(expectedMissingVersions, e.missingVersions); diff != "" {
			t.Errorf("unexpected missing versions (-want +got):\n%s", diff)
		}
	})

	t.Run("up to date", func(t *testing.T) {
		store := testStoreWithVersion(10004, false)

		if err := makeTestRunner(t, store).Validate(ctx, "well-formed"); err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
	})

	t.Run("future upgrade", func(t *testing.T) {
		store := testStoreWithVersion(10008, false)

		if err := makeTestRunner(t, store).Validate(ctx, "well-formed"); err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
	})

	t.Run("dirty database", func(t *testing.T) {
		store := testStoreWithVersion(10003, true)

		e := new(dirtySchemaError)
		if err := makeTestRunner(t, store).Validate(ctx, "well-formed"); err == nil {
			t.Fatalf("expected an error")
		} else if !errors.As(err, &e) {
			t.Fatalf("expected error. want schema out of date error, have=%s", err)
		}

		expectedFailedVersions := []int{
			10003,
		}
		if diff := cmp.Diff(expectedFailedVersions, extractIDs(e.dirtyVersions)); diff != "" {
			t.Errorf("unexpected failed versions (-want +got):\n%s", diff)
		}
	})

	t.Run("dirty database (dead migrations)", func(t *testing.T) {
		store := testStoreWithVersion(10003, false)
		store.VersionsFunc.SetDefaultReturn(nil, []int{10003}, nil, nil)

		e := new(dirtySchemaError)
		if err := makeTestRunner(t, store).Validate(ctx, "well-formed"); err == nil {
			t.Fatalf("expected an error")
		} else if !errors.As(err, &e) {
			t.Fatalf("expected error. want schema out of date error, have=%s", err)
		}

		expectedFailedVersions := []int{
			10003,
		}
		if diff := cmp.Diff(expectedFailedVersions, extractIDs(e.dirtyVersions)); diff != "" {
			t.Errorf("unexpected failed versions (-want +got):\n%s", diff)
		}
	})
}
