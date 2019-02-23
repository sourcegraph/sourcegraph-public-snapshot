package repos

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/kylelemons/godebug/pretty"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/github"
)

// This error is passed to txstore.Done in order to always
// roll-back the transaction a test case executes in.
// This is meant to ensure each test case has a clean slate.
var errRollback = errors.New("tx: rollback")

func TestIntegration_DBStore(t *testing.T) {
	t.Parallel()

	db, cleanup := testDatabase(t)
	defer cleanup()

	for _, tc := range []struct {
		name string
		test func(*testing.T)
	}{
		{"UpsertRepos", testDBStoreUpsertRepos(db)},
		{"ListRepos", testDBStoreListRepos(db)},
	} {
		t.Run(tc.name, tc.test)
	}
}

func testDBStoreUpsertRepos(db *sql.DB) func(*testing.T) {
	return func(t *testing.T) {
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
					Description: "The description",
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
					Sources: map[string]*SourceInfo{
						"extsvc:123": {
							ID:       "extsvc:123",
							CloneURL: "git@github.com:foo/bar.git",
						},
					},
					Metadata: []byte("{}"),
				})
			}

			txstore, err := store.Transact(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer txstore.Done(&errRollback)

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

			suffix := "-updated"
			now := time.Now()
			for _, r := range want {
				r.Name += suffix
				r.Description += suffix
				r.Language += suffix
				r.UpdatedAt = now
				r.Archived = !r.Archived
				r.Fork = !r.Fork
			}

			if err = txstore.UpsertRepos(ctx, want...); err != nil {
				t.Errorf("UpsertRepos error: %s", err)
			} else if have, err = txstore.ListRepos(ctx); err != nil {
				t.Errorf("ListRepos error: %s", err)
			} else if diff := pretty.Compare(have, want); diff != "" {
				t.Errorf("ListRepos:\n%s", diff)
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
}

func testDBStoreListRepos(db *sql.DB) func(*testing.T) {
	foo := Repo{
		Name: "foo",
		Sources: map[string]*SourceInfo{
			"extsvc:123": {
				ID:       "extsvc:123",
				CloneURL: "git@github.com:bar/foo.git",
			},
		},
		Metadata: new(github.Repository),
		ExternalRepo: api.ExternalRepoSpec{
			ServiceType: "github",
			ServiceID:   "https://github.com/",
			ID:          "foo",
		},
	}

	return func(t *testing.T) {
		for _, tc := range []struct {
			name    string
			kinds   []string
			ctx     context.Context
			names   []string
			in, out []*Repo
			err     error
		}{
			{
				name:  "case-insensitive kinds",
				kinds: []string{"GiThUb"},
				in: Repos{foo.With(func(r *Repo) {
					r.ExternalRepo.ServiceType = "gItHuB"
				})},
				out: Repos{foo.With(func(r *Repo) {
					r.ExternalRepo.ServiceType = "gItHuB"
				})},
			},
		} {
			// NOTE: We use t.Errorf instead of t.Fatalf in order to run defers.

			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				ctx := tc.ctx
				if ctx == nil {
					ctx = context.Background()
				}

				txstore, err := NewDBStore(
					ctx,
					db,
					tc.kinds,
					sql.TxOptions{Isolation: sql.LevelDefault},
				).Transact(ctx)

				if err != nil {
					t.Errorf("failed to start transaction: %v", err)
					return
				}

				defer txstore.Done(&errRollback)

				if err := txstore.UpsertRepos(ctx, tc.in...); err != nil {
					t.Errorf("failed to setup store: %v", err)
					return
				}

				out, err := txstore.ListRepos(ctx, tc.names...)
				if have, want := fmt.Sprint(err), fmt.Sprint(tc.err); have != want {
					t.Errorf("error:\nhave: %v\nwant: %v", have, want)
				}

				for _, r := range out {
					r.ID = 0 // Exclude auto-generated IDs from equality tests
				}

				if have, want := out, tc.out; !reflect.DeepEqual(have, want) {
					t.Errorf("repos: %s", cmp.Diff(have, want))
				}
			})
		}
	}
}
