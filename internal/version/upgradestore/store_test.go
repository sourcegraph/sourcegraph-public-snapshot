package upgradestore

import (
	"context"
	"testing"

	"github.com/Masterminds/semver"
	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestGetServiceVersion(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db)

	t.Run("fresh db", func(t *testing.T) {
		_, ok, err := store.GetServiceVersion(ctx, "service")
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if ok {
			t.Fatalf("did not expect value")
		}
	})

	t.Run("after updates", func(t *testing.T) {
		if err := store.UpdateServiceVersion(ctx, "service", "1.2.3"); err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if err := store.UpdateServiceVersion(ctx, "service", "1.2.4"); err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if err := store.UpdateServiceVersion(ctx, "service", "1.3.0"); err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		version, ok, err := store.GetServiceVersion(ctx, "service")
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if !ok {
			t.Fatalf("unexpected value, got none")
		}
		if version != "1.3.0" {
			t.Errorf("unexpected version. want=%s have=%s", "1.3.0", version)
		}
	})

	t.Run("missing table", func(t *testing.T) {
		if err := store.db.Exec(ctx, sqlf.Sprintf("DROP TABLE versions;")); err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		_, ok, err := store.GetServiceVersion(ctx, "service")
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if ok {
			t.Fatalf("did not expect value")
		}
	})
}

func TestSetServiceVersion(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db)

	if err := store.UpdateServiceVersion(ctx, "service", "1.2.3"); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if err := store.SetServiceVersion(ctx, "service", "1.2.5"); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	version, _, err := store.GetServiceVersion(ctx, "service")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if want := "1.2.5"; version != want {
		t.Fatalf("unexpected version. want=%q have=%q", want, version)
	}
}

func TestGetFirstServiceVersion(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db)

	t.Run("fresh db", func(t *testing.T) {
		_, ok, err := store.GetFirstServiceVersion(ctx, "service")
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if ok {
			t.Fatalf("did not expect value")
		}
	})

	t.Run("after updates", func(t *testing.T) {
		if err := store.UpdateServiceVersion(ctx, "service", "1.2.3"); err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if err := store.UpdateServiceVersion(ctx, "service", "1.2.4"); err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if err := store.UpdateServiceVersion(ctx, "service", "1.3.0"); err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		firstVersion, ok, err := store.GetFirstServiceVersion(ctx, "service")
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if !ok {
			t.Fatalf("unexpected value, got none")
		}
		if firstVersion != "1.2.3" {
			t.Errorf("unexpected first version. want=%s have=%s", "1.2.3", firstVersion)
		}
	})

	t.Run("missing table", func(t *testing.T) {
		if err := store.db.Exec(ctx, sqlf.Sprintf("DROP TABLE versions;")); err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		_, ok, err := store.GetFirstServiceVersion(ctx, "service")
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if ok {
			t.Fatalf("did not expect value")
		}
	})
}

func TestUpdateServiceVersion(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db)

	t.Run("update sequence", func(t *testing.T) {
		for _, tc := range []struct {
			version string
			err     error
		}{
			{"0.0.0", nil},
			{"0.0.1", nil},
			{"0.1.0", nil},
			{"0.2.0", nil},
			{"1.0.0", nil},
			{"1.2.0", &UpgradeError{
				Service:  "service",
				Previous: semver.MustParse("1.0.0"),
				Latest:   semver.MustParse("1.2.0"),
			}},
			{"2.1.0", &UpgradeError{
				Service:  "service",
				Previous: semver.MustParse("1.0.0"),
				Latest:   semver.MustParse("2.1.0"),
			}},
			{"0.3.0", nil}, // rollback
			{"non-semantic-version-is-always-valid", nil},
			{"1.0.0", nil}, // back to semantic version is allowed
			{"2.1.0", &UpgradeError{
				Service:  "service",
				Previous: semver.MustParse("1.0.0"),
				Latest:   semver.MustParse("2.1.0"),
			}}, // upgrade policy violation returns
		} {
			have := store.UpdateServiceVersion(ctx, "service", tc.version)
			want := tc.err

			if !errors.Is(have, want) {
				t.Fatal(cmp.Diff(have, want))
			}

			t.Logf("version = %q", tc.version)
		}
	})

	t.Run("missing table", func(t *testing.T) {
		if err := store.db.Exec(ctx, sqlf.Sprintf("DROP TABLE versions;")); err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		if err := store.UpdateServiceVersion(ctx, "service", "0.0.1"); err == nil {
			t.Fatalf("expected error, got none")
		}
	})
}

func TestValidateUpgrade(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db)

	t.Run("missing table", func(t *testing.T) {
		if err := store.db.Exec(ctx, sqlf.Sprintf("DROP TABLE versions;")); err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		if err := store.ValidateUpgrade(ctx, "service", "0.0.1"); err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
	})
}
