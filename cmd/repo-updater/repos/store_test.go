package repos

import (
	"context"
	"database/sql"
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/kylelemons/godebug/pretty"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

func TestIntegration_DBStore(t *testing.T) {
	t.Parallel()

	db, cleanup := testDatabase(t)
	defer cleanup()

	kinds := []string{
		"GITHUB",
	}

	ctx := context.Background()
	store := NewDBStore(ctx, db, kinds, sql.TxOptions{Isolation: sql.LevelSerializable})

	t.Run("no repos", func(t *testing.T) {
		if err := store.UpsertRepos(ctx); err != nil {
			t.Fatalf("UpsertRepos error: %s", err)
		}
	})

	t.Run("many repos", func(t *testing.T) {
		want := make([]*Repo, 0, 512) // Test more than one page load
		for i := 0; i < cap(want); i++ {
			id := strconv.Itoa(i)
			want = append(want, &Repo{
				Name:        "github.com/foo/bar" + id,
				Description: "It's a foo's bar",
				Language:    "barlang",
				Enabled:     true,
				Archived:    false,
				Fork:        false,
				CreatedAt:   time.Now().UTC(),
				ExternalRepo: api.ExternalRepoSpec{
					ID:          id,
					ServiceType: "github",
					ServiceID:   "http://github.com",
				},
			})
		}

		txstore, err := store.Transact(ctx)
		if err != nil {
			t.Fatal(err)
		}
		defer txstore.Done(&err)

		// NOTE(tsenart): We use t.Errorf followed by a return statement instead
		// of t.Fatalf so that the defered txstore.Done is executed.

		if err = txstore.UpsertRepos(ctx, want...); err != nil {
			t.Errorf("UpsertRepos error: %s", err)
			return
		}

		sort.Slice(want, func(i, j int) bool {
			return want[i].ID < want[j].ID
		})

		have, err := txstore.ListRepos(ctx)
		if err != nil {
			t.Errorf("ListRepos error: %s", err)
			return
		}

		if diff := pretty.Compare(have, want); diff != "" {
			t.Errorf("ListRepos:\n%s", diff)
			return
		}

		for i := 1; i <= 5; i++ {
			suffix := " " + strconv.Itoa(i)
			now := time.Now()
			for _, r := range want {
				r.Name += suffix
				r.Description += suffix
				r.Language += suffix
				r.UpdatedAt = now
				r.Archived = !r.Archived
				r.Fork = !r.Fork

				// Not updateable fields. Check that that UpsertRepos
				// restores their original value.
				r.ID += 10000
				r.Enabled = !r.Enabled
				r.CreatedAt = r.CreatedAt.Add(time.Minute)
			}

			if err = txstore.UpsertRepos(ctx, want...); err != nil {
				t.Errorf("UpsertRepos error: %s", err)
			} else if have, err = txstore.ListRepos(ctx); err != nil {
				t.Errorf("ListRepos error: %s", err)
			} else if diff := pretty.Compare(have, want); diff != "" {
				t.Errorf("ListRepos:\n%s", diff)
			}
		}

		for _, repo := range want {
			repo.DeletedAt = time.Now().UTC()
		}

		if err = txstore.UpsertRepos(ctx, want...); err != nil {
			t.Errorf("UpsertRepos error: %s", err)
		} else if have, err = txstore.ListRepos(ctx); err != nil {
			t.Errorf("ListRepos error: %s", err)
		} else if diff := pretty.Compare(have, []*Repo{}); diff != "" {
			t.Errorf("ListRepos:\n%s", diff)
		}

	})
}
