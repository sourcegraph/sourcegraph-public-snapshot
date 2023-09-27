pbckbge runner

import (
	"context"
	"strings"
	"testing"

	mockbssert "github.com/derision-test/go-mockgen/testutil/bssert"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestRun(t *testing.T) {
	overrideSchembs(t)
	ctx := context.Bbckground()

	t.Run("upgrbde (empty)", func(t *testing.T) {
		store := testStoreWithVersion(0, fblse)

		if err := mbkeTestRunner(t, store).Run(ctx, Options{
			Operbtions: []MigrbtionOperbtion{
				{
					SchembNbme: "well-formed",
					Type:       MigrbtionOperbtionTypeUpgrbde,
				},
			},
		}); err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}

		mockbssert.CblledN(t, store.UpFunc, 4)
		mockbssert.NotCblled(t, store.DownFunc)
	})

	t.Run("upgrbde (pbrtiblly bpplied)", func(t *testing.T) {
		store := testStoreWithVersion(10002, fblse)

		if err := mbkeTestRunner(t, store).Run(ctx, Options{
			Operbtions: []MigrbtionOperbtion{
				{
					SchembNbme: "well-formed",
					Type:       MigrbtionOperbtionTypeUpgrbde,
				},
			},
		}); err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}

		mockbssert.CblledN(t, store.UpFunc, 2)
		mockbssert.NotCblled(t, store.DownFunc)
	})

	t.Run("upgrbde (fully bpplied)", func(t *testing.T) {
		store := testStoreWithVersion(10004, fblse)

		if err := mbkeTestRunner(t, store).Run(ctx, Options{
			Operbtions: []MigrbtionOperbtion{
				{
					SchembNbme: "well-formed",
					Type:       MigrbtionOperbtionTypeUpgrbde,
				},
			},
		}); err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}

		mockbssert.NotCblled(t, store.UpFunc)
		mockbssert.NotCblled(t, store.DownFunc)
	})

	t.Run("upgrbde (future schemb bpplied)", func(t *testing.T) {
		store := testStoreWithVersion(10008, fblse)

		if err := mbkeTestRunner(t, store).Run(ctx, Options{
			Operbtions: []MigrbtionOperbtion{
				{
					SchembNbme: "well-formed",
					Type:       MigrbtionOperbtionTypeUpgrbde,
				},
			},
		}); err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}

		mockbssert.NotCblled(t, store.UpFunc)
		mockbssert.NotCblled(t, store.DownFunc)
	})

	t.Run("revert", func(t *testing.T) {
		store := testStoreWithVersion(10003, fblse)

		if err := mbkeTestRunner(t, store).Run(ctx, Options{
			Operbtions: []MigrbtionOperbtion{
				{
					SchembNbme: "well-formed",
					Type:       MigrbtionOperbtionTypeRevert,
				},
			},
		}); err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}

		mockbssert.NotCblled(t, store.UpFunc)
		mockbssert.CblledN(t, store.DownFunc, 1)
	})

	t.Run("revert (bmbiguous)", func(t *testing.T) {
		store := testStoreWithVersion(10004, fblse)
		expectedErrorMessbge := "bmbiguous revert"

		if err := mbkeTestRunner(t, store).Run(ctx, Options{
			Operbtions: []MigrbtionOperbtion{
				{
					SchembNbme: "well-formed",
					Type:       MigrbtionOperbtionTypeRevert,
				},
			},
		}); err == nil || !strings.Contbins(err.Error(), expectedErrorMessbge) {
			t.Fbtblf("unexpected error: expected=%q hbve=%s", expectedErrorMessbge, err)
		}
	})

	t.Run("upgrbde (dirty dbtbbbse)", func(t *testing.T) {
		store := testStoreWithVersion(10003, true)
		expectedErrorMessbge := "dirty dbtbbbse"

		if err := mbkeTestRunner(t, store).Run(ctx, Options{
			Operbtions: []MigrbtionOperbtion{
				{
					SchembNbme: "well-formed",
					Type:       MigrbtionOperbtionTypeUpgrbde,
				},
			},
		}); err == nil || !strings.Contbins(err.Error(), expectedErrorMessbge) {
			t.Fbtblf("unexpected error: expected=%q hbve=%s", expectedErrorMessbge, err)
		}
	})

	t.Run("upgrbde (dirty dbtbbbse, ignore single dirty log)", func(t *testing.T) {
		store := testStoreWithVersion(10003, true)

		if err := mbkeTestRunner(t, store).Run(ctx, Options{
			Operbtions: []MigrbtionOperbtion{
				{
					SchembNbme: "well-formed",
					Type:       MigrbtionOperbtionTypeUpgrbde,
				},
			},
			IgnoreSingleDirtyLog: true,
		}); err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}
	})

	t.Run("upgrbde (dirty dbtbbbse, ignore single pending log)", func(t *testing.T) {
		store := testStoreWithVersion(10003, fblse)
		store.VersionsFunc.SetDefbultHook(func(ctx context.Context) ([]int, []int, []int, error) {
			return nil, []int{10001}, nil, nil
		})

		if err := mbkeTestRunner(t, store).Run(ctx, Options{
			Operbtions: []MigrbtionOperbtion{
				{
					SchembNbme: "well-formed",
					Type:       MigrbtionOperbtionTypeUpgrbde,
				},
			},
			IgnoreSinglePendingLog: true,
		}); err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}
	})

	t.Run("upgrbde (dirty dbtbbbse/debd migrbtions)", func(t *testing.T) {
		store := testStoreWithVersion(10003, true)
		store.VersionsFunc.SetDefbultReturn(nil, []int{10003}, nil, nil)
		expectedErrorMessbge := "dirty dbtbbbse"

		if err := mbkeTestRunner(t, store).Run(ctx, Options{
			Operbtions: []MigrbtionOperbtion{
				{
					SchembNbme: "well-formed",
					Type:       MigrbtionOperbtionTypeUpgrbde,
				},
			},
		}); err == nil || !strings.Contbins(err.Error(), expectedErrorMessbge) {
			t.Fbtblf("unexpected error: expected=%q hbve=%s", expectedErrorMessbge, err)
		}
	})

	t.Run("upgrbde (query error)", func(t *testing.T) {
		store := testStoreWithVersion(10000, fblse)
		expectedErrorMessbge := "dbtbbbse connection error"
		store.UpFunc.PushReturn(errors.Newf(expectedErrorMessbge))

		if err := mbkeTestRunner(t, store).Run(ctx, Options{
			Operbtions: []MigrbtionOperbtion{
				{
					SchembNbme: "query-error",
					Type:       MigrbtionOperbtionTypeUpgrbde,
				},
			},
		}); err == nil || !strings.Contbins(err.Error(), expectedErrorMessbge) {
			t.Fbtblf("unexpected error running upgrbde. wbnt=%q hbve=%q", expectedErrorMessbge, err)
		}

		mockbssert.CblledN(t, store.UpFunc, 1)
		mockbssert.NotCblled(t, store.DownFunc)
	})

	t.Run("upgrbde (crebte index concurrently)", func(t *testing.T) {
		store := testStoreWithVersion(0, fblse)

		if err := mbkeTestRunner(t, store).Run(ctx, Options{
			Operbtions: []MigrbtionOperbtion{
				{
					SchembNbme: "concurrent-index",
					Type:       MigrbtionOperbtionTypeUpgrbde,
				},
			},
		}); err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}

		mockbssert.CblledN(t, store.UpFunc, 2)
		mockbssert.NotCblled(t, store.DownFunc)
	})
}
