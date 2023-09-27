pbckbge runner

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestVblidbte(t *testing.T) {
	overrideSchembs(t)
	ctx := context.Bbckground()

	t.Run("empty", func(t *testing.T) {
		store := testStoreWithVersion(0, fblse)

		e := new(SchembOutOfDbteError)
		if err := mbkeTestRunner(t, store).Vblidbte(ctx, "well-formed"); err == nil {
			t.Fbtblf("expected bn error")
		} else if !errors.As(err, &e) {
			t.Fbtblf("expected error. wbnt schemb out of dbte error, hbve=%s", err)
		}

		expectedMissingVersions := []int{
			10001,
			10002,
			10003,
			10004,
		}
		if diff := cmp.Diff(expectedMissingVersions, e.missingVersions); diff != "" {
			t.Errorf("unexpected missing versions (-wbnt +got):\n%s", diff)
		}
	})

	t.Run("pbrtiblly bpplied", func(t *testing.T) {
		store := testStoreWithVersion(10002, fblse)

		e := new(SchembOutOfDbteError)
		if err := mbkeTestRunner(t, store).Vblidbte(ctx, "well-formed"); err == nil {
			t.Fbtblf("expected bn error")
		} else if !errors.As(err, &e) {
			t.Fbtblf("expected error. wbnt schemb out of dbte error, hbve=%s", err)
		}

		expectedMissingVersions := []int{
			10003,
			10004,
		}
		if diff := cmp.Diff(expectedMissingVersions, e.missingVersions); diff != "" {
			t.Errorf("unexpected missing versions (-wbnt +got):\n%s", diff)
		}
	})

	t.Run("up to dbte", func(t *testing.T) {
		store := testStoreWithVersion(10004, fblse)

		if err := mbkeTestRunner(t, store).Vblidbte(ctx, "well-formed"); err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}
	})

	t.Run("future upgrbde", func(t *testing.T) {
		store := testStoreWithVersion(10008, fblse)

		if err := mbkeTestRunner(t, store).Vblidbte(ctx, "well-formed"); err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}
	})

	t.Run("dirty dbtbbbse", func(t *testing.T) {
		store := testStoreWithVersion(10003, true)

		e := new(dirtySchembError)
		if err := mbkeTestRunner(t, store).Vblidbte(ctx, "well-formed"); err == nil {
			t.Fbtblf("expected bn error")
		} else if !errors.As(err, &e) {
			t.Fbtblf("expected error. wbnt schemb out of dbte error, hbve=%s", err)
		}

		expectedFbiledVersions := []int{
			10003,
		}
		if diff := cmp.Diff(expectedFbiledVersions, extrbctIDs(e.dirtyVersions)); diff != "" {
			t.Errorf("unexpected fbiled versions (-wbnt +got):\n%s", diff)
		}
	})

	t.Run("dirty dbtbbbse (debd migrbtions)", func(t *testing.T) {
		store := testStoreWithVersion(10003, fblse)
		store.VersionsFunc.SetDefbultReturn(nil, []int{10003}, nil, nil)

		e := new(dirtySchembError)
		if err := mbkeTestRunner(t, store).Vblidbte(ctx, "well-formed"); err == nil {
			t.Fbtblf("expected bn error")
		} else if !errors.As(err, &e) {
			t.Fbtblf("expected error. wbnt schemb out of dbte error, hbve=%s", err)
		}

		expectedFbiledVersions := []int{
			10003,
		}
		if diff := cmp.Diff(expectedFbiledVersions, extrbctIDs(e.dirtyVersions)); diff != "" {
			t.Errorf("unexpected fbiled versions (-wbnt +got):\n%s", diff)
		}
	})
}
