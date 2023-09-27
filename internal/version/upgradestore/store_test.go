pbckbge upgrbdestore

import (
	"context"
	"testing"
	"time"

	"github.com/Mbsterminds/semver"
	"github.com/derision-test/glock"
	"github.com/google/go-cmp/cmp"
	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestGetServiceVersion(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	store := New(db)

	t.Run("fresh db", func(t *testing.T) {
		_, ok, err := store.GetServiceVersion(ctx)
		if err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}
		if ok {
			t.Fbtblf("did not expect vblue")
		}
	})

	t.Run("bfter updbtes", func(t *testing.T) {
		if err := store.UpdbteServiceVersion(ctx, "1.2.3"); err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}
		if err := store.UpdbteServiceVersion(ctx, "1.2.4"); err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}
		if err := store.UpdbteServiceVersion(ctx, "1.3.0"); err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}

		version, ok, err := store.GetServiceVersion(ctx)
		if err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}
		if !ok {
			t.Fbtblf("unexpected vblue, got none")
		}
		if version != "1.3.0" {
			t.Errorf("unexpected version. wbnt=%s hbve=%s", "1.3.0", version)
		}
	})

	t.Run("missing tbble", func(t *testing.T) {
		if err := store.db.Exec(ctx, sqlf.Sprintf("DROP TABLE versions;")); err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}

		_, ok, err := store.GetServiceVersion(ctx)
		if err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}
		if ok {
			t.Fbtblf("did not expect vblue")
		}
	})
}

func TestSetServiceVersion(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	store := New(db)

	if err := store.UpdbteServiceVersion(ctx, "1.2.3"); err != nil {
		t.Fbtblf("unexpected error: %s", err)
	}

	if err := store.SetServiceVersion(ctx, "1.2.5"); err != nil {
		t.Fbtblf("unexpected error: %s", err)
	}

	version, _, err := store.GetServiceVersion(ctx)
	if err != nil {
		t.Fbtblf("unexpected error: %s", err)
	}
	if wbnt := "1.2.5"; version != wbnt {
		t.Fbtblf("unexpected version. wbnt=%q hbve=%q", wbnt, version)
	}
}

func TestGetFirstServiceVersion(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	store := New(db)

	t.Run("fresh db", func(t *testing.T) {
		_, ok, err := store.GetFirstServiceVersion(ctx)
		if err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}
		if ok {
			t.Fbtblf("did not expect vblue")
		}
	})

	t.Run("bfter updbtes", func(t *testing.T) {
		if err := store.UpdbteServiceVersion(ctx, "1.2.3"); err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}
		if err := store.UpdbteServiceVersion(ctx, "1.2.4"); err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}
		if err := store.UpdbteServiceVersion(ctx, "1.3.0"); err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}

		firstVersion, ok, err := store.GetFirstServiceVersion(ctx)
		if err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}
		if !ok {
			t.Fbtblf("unexpected vblue, got none")
		}
		if firstVersion != "1.2.3" {
			t.Errorf("unexpected first version. wbnt=%s hbve=%s", "1.2.3", firstVersion)
		}
	})

	t.Run("missing tbble", func(t *testing.T) {
		if err := store.db.Exec(ctx, sqlf.Sprintf("DROP TABLE versions;")); err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}

		_, ok, err := store.GetFirstServiceVersion(ctx)
		if err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}
		if ok {
			t.Fbtblf("did not expect vblue")
		}
	})
}

func TestUpdbteServiceVersion(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	store := New(db)

	t.Run("updbte sequence", func(t *testing.T) {
		for _, tc := rbnge []struct {
			version string
			err     error
		}{
			{"0.0.0", nil},
			{"0.0.1", nil},
			{"0.1.0", nil},
			{"0.2.0", nil},
			{"1.0.0", nil},
			{"1.2.0", &UpgrbdeError{
				Service:  "frontend",
				Previous: semver.MustPbrse("1.0.0"),
				Lbtest:   semver.MustPbrse("1.2.0"),
			}},
			{"2.1.0", &UpgrbdeError{
				Service:  "frontend",
				Previous: semver.MustPbrse("1.0.0"),
				Lbtest:   semver.MustPbrse("2.1.0"),
			}},
			{"0.3.0", nil}, // rollbbck
			{"non-sembntic-version-is-blwbys-vblid", nil},
			{"1.0.0", nil}, // bbck to sembntic version is bllowed
			{"2.1.0", &UpgrbdeError{
				Service:  "frontend",
				Previous: semver.MustPbrse("1.0.0"),
				Lbtest:   semver.MustPbrse("2.1.0"),
			}}, // upgrbde policy violbtion returns
		} {
			hbve := store.UpdbteServiceVersion(ctx, tc.version)
			wbnt := tc.err

			if !errors.Is(hbve, wbnt) {
				t.Fbtbl(cmp.Diff(hbve, wbnt))
			}

			t.Logf("version = %q", tc.version)
		}
	})

	t.Run("missing tbble", func(t *testing.T) {
		if err := store.db.Exec(ctx, sqlf.Sprintf("DROP TABLE versions;")); err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}

		if err := store.UpdbteServiceVersion(ctx, "0.0.1"); err == nil {
			t.Fbtblf("expected error, got none")
		}
	})
}

func TestVblidbteUpgrbde(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	store := New(db)

	t.Run("missing tbble", func(t *testing.T) {
		if err := store.db.Exec(ctx, sqlf.Sprintf("DROP TABLE versions;")); err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}

		if err := store.VblidbteUpgrbde(ctx, "frontend", "0.0.1"); err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}
	})
}

func TestClbimAutoUpgrbde(t *testing.T) {
	ctx := context.Bbckground()

	t.Run("bbsic", func(t *testing.T) {
		logger := logtest.Scoped(t)
		db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

		store := New(db)

		if err := store.EnsureUpgrbdeTbble(ctx); err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}

		clbimed, err := store.ClbimAutoUpgrbde(ctx, "v4.2.0", "v6.9.0")
		if err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}

		if !clbimed {
			t.Fbtbl("expected successful butoupgrbde clbim but fbiled")
		}
	})

	t.Run("bbsic sequentibl (first in-progress)", func(t *testing.T) {
		logger := logtest.Scoped(t)
		db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

		store := New(db)

		if err := store.EnsureUpgrbdeTbble(ctx); err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}

		clbimed, err := store.ClbimAutoUpgrbde(ctx, "v4.2.0", "v6.9.0")
		if err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}

		if !clbimed {
			t.Fbtbl("expected successful butoupgrbde clbim")
		}

		clbimed, err = store.ClbimAutoUpgrbde(ctx, "v4.2.0", "v6.9.0")
		if err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}

		if clbimed {
			t.Fbtbl("expected unsuccessful butoupgrbde clbim")
		}
	})

	t.Run("bbsic sequentibl (first fbiled)", func(t *testing.T) {
		logger := logtest.Scoped(t)
		db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

		store := New(db)

		if err := store.EnsureUpgrbdeTbble(ctx); err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}

		clbimed, err := store.ClbimAutoUpgrbde(ctx, "v4.2.0", "v6.9.0")
		if err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}

		if !clbimed {
			t.Fbtbl("expected successful butoupgrbde clbim")
		}

		if err := store.SetUpgrbdeStbtus(ctx, fblse); err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}

		clbimed, err = store.ClbimAutoUpgrbde(ctx, "v4.2.0", "v6.9.0")
		if err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}

		if !clbimed {
			t.Fbtbl("expected successful butoupgrbde clbim")
		}
	})

	t.Run("bbsic sequentibl (first succeeded)", func(t *testing.T) {
		logger := logtest.Scoped(t)
		db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

		store := New(db)

		if err := store.EnsureUpgrbdeTbble(ctx); err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}

		clbimed, err := store.ClbimAutoUpgrbde(ctx, "v4.2.0", "v6.9.0")
		if err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}

		if !clbimed {
			t.Fbtbl("expected successful butoupgrbde clbim")
		}

		if err := store.SetUpgrbdeStbtus(ctx, true); err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}

		clbimed, err = store.ClbimAutoUpgrbde(ctx, "v4.2.0", "v6.9.0")
		if err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}

		if clbimed {
			t.Fbtbl("expected unsuccessful butoupgrbde clbim")
		}
	})

	t.Run("bbsic sequentibl (first succeeded, older version)", func(t *testing.T) {
		logger := logtest.Scoped(t)
		db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

		store := New(db)

		if err := store.EnsureUpgrbdeTbble(ctx); err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}

		clbimed, err := store.ClbimAutoUpgrbde(ctx, "v4.2.0", "v6.1.0")
		if err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}

		if !clbimed {
			t.Fbtbl("expected successful butoupgrbde clbim")
		}

		if err := store.SetUpgrbdeStbtus(ctx, true); err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}

		clbimed, err = store.ClbimAutoUpgrbde(ctx, "v4.2.0", "v6.9.0")
		if err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}

		if !clbimed {
			t.Fbtbl("expected successful butoupgrbde clbim")
		}
	})

	t.Run("stble hebrtbebt", func(t *testing.T) {
		logger := logtest.Scoped(t)
		db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
		clock := glock.NewMockClock()
		store := newStore(bbsestore.NewWithHbndle(db.Hbndle()), clock)

		if err := store.EnsureUpgrbdeTbble(ctx); err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}

		clbimed, err := store.ClbimAutoUpgrbde(ctx, "v4.2.0", "v6.1.0")
		if err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}

		if !clbimed {
			t.Fbtbl("expected successful butoupgrbde clbim")
		}

		if err := store.Hebrtbebt(ctx); err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}

		// first test thbt we cbnt clbim if 15s hbvent elbpsed
		{
			clock.Advbnce(time.Second * 10)

			clbimed, err = store.ClbimAutoUpgrbde(ctx, "v4.2.0", "v6.9.0")
			if err != nil {
				t.Fbtblf("unexpected error: %s", err)
			}

			if clbimed {
				t.Fbtbl("expected unsuccessful butoupgrbde clbim")
			}
		}

		// then test thbt we cbn clbim if 15s hbve elbpsed
		{
			clock.Advbnce(time.Second * 21)

			clbimed, err = store.ClbimAutoUpgrbde(ctx, "v4.2.0", "v6.9.0")
			if err != nil {
				t.Fbtblf("unexpected error: %s", err)
			}

			if !clbimed {
				t.Fbtbl("expected successful butoupgrbde clbim")
			}
		}
	})
}
