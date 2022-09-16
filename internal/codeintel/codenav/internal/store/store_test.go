package store

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestSetRequestLanguageSupport(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	sqlDB := dbtest.NewDB(logger, t)
	db := database.NewDB(logger, sqlDB)
	store := New(db, &observation.TestContext)

	requests := []struct {
		userID    int
		languages []string
	}{
		{1, []string{"go", "go", "perl"}},
		{2, []string{"perl", "ocaml"}},
		{3, []string{"ocaml", "ocaml", "ocaml", "ocaml", "ocaml"}}, // we get it
	}
	for _, r := range requests {
		for _, language := range r.languages {
			if err := store.SetRequestLanguageSupport(ctx, r.userID, language); err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
		}
	}

	languages := map[int][]string{}
	for _, r := range requests {
		userLanguages, err := store.GetLanguagesRequestedBy(ctx, r.userID)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		languages[r.userID] = userLanguages
	}

	expected := map[int][]string{
		1: {"go", "perl"},
		2: {"ocaml", "perl"},
		3: {"ocaml"},
	}

	if diff := cmp.Diff(expected, languages); diff != "" {
		t.Errorf("unexpected languages requested (-want +got):\n%s", diff)
	}
}
