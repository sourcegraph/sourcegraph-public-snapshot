package repos_test

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
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/github"
)

func testDBStoreUpsertRepos(db *sql.DB) func(*testing.T) {
	return func(t *testing.T) {
		t.Helper()

		kinds := []string{
			"GITHUB",
		}

		ctx := context.Background()
		store := repos.NewDBStore(ctx, db, sql.TxOptions{Isolation: sql.LevelSerializable})

		t.Run("no repos", func(t *testing.T) {
			if err := store.UpsertRepos(ctx); err != nil {
				t.Fatalf("UpsertRepos error: %s", err)
			}
		})

		t.Run("many repos", transact(ctx, store, func(t testing.TB, tx repos.Store) {
			want := make(repos.Repos, 0, 512) // Test more than one page load
			for i := 0; i < cap(want); i++ {
				id := strconv.Itoa(i)
				want = append(want, &repos.Repo{
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
					Sources: map[string]*repos.SourceInfo{
						"extsvc:123": {
							ID:       "extsvc:123",
							CloneURL: "git@github.com:foo/bar.git",
						},
					},
					Metadata: []byte("{}"),
				})
			}

			if err := tx.UpsertRepos(ctx, want...); err != nil {
				t.Errorf("UpsertRepos error: %s", err)
				return
			}

			sort.Slice(want, func(i, j int) bool {
				return want[i].ID < want[j].ID
			})

			have, err := tx.ListRepos(ctx, kinds...)
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

			if err = tx.UpsertRepos(ctx, want...); err != nil {
				t.Errorf("UpsertRepos error: %s", err)
			} else if have, err = tx.ListRepos(ctx); err != nil {
				t.Errorf("ListRepos error: %s", err)
			} else if diff := pretty.Compare(have, want); diff != "" {
				t.Errorf("ListRepos:\n%s", diff)
			}

			want.Apply(repos.Opt.DeletedAt(time.Now().UTC()))

			if err = tx.UpsertRepos(ctx, want...); err != nil {
				t.Errorf("UpsertRepos error: %s", err)
			} else if have, err = tx.ListRepos(ctx); err != nil {
				t.Errorf("ListRepos error: %s", err)
			} else if diff := pretty.Compare(have, want); diff != "" {
				t.Errorf("ListRepos:\n%s", diff)
			}

		}))
	}
}

func testDBStoreListRepos(db *sql.DB) func(*testing.T) {
	foo := repos.Repo{
		Name: "foo",
		Sources: map[string]*repos.SourceInfo{
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
	unmanaged := repos.Repo{
		Name:     "unmanaged",
		Sources:  map[string]*repos.SourceInfo{},
		Metadata: new(github.Repository),
		ExternalRepo: api.ExternalRepoSpec{
			ServiceType: "non_existent_kind",
			ServiceID:   "https://example.com/",
			ID:          "unmanaged",
		},
	}

	// TODO handle this case. Probably load all repos, or also load by external cols
	//
	// foo is managed by a
	// a is deleted => foo is marked deleted
	// foo is renamed to bar upstream
	// a is recreated
	// ListRepos(ctx, "bar") -> won't return bar since it doesn't exist and foo is deleted
	// does the upsert fail, since it thinks bar is a new repo and store doesn't return it

	return func(t *testing.T) {
		t.Helper()

		for _, tc := range []struct {
			name   string
			kinds  []string
			stored repos.Repos
			repos  repos.Repos
			err    error
		}{
			{
				name:  "case-insensitive kinds",
				kinds: []string{"GiThUb"},
				stored: repos.Repos{foo.With(func(r *repos.Repo) {
					r.ExternalRepo.ServiceType = "gItHuB"
				})},
				repos: repos.Repos{foo.With(func(r *repos.Repo) {
					r.ExternalRepo.ServiceType = "gItHuB"
				})},
			},
			{
				name:   "ignores unmanaged",
				kinds:  []string{"github"},
				stored: repos.Repos{&foo, &unmanaged}.Clone(),
				repos:  repos.Repos{&foo}.Clone(),
			},
		} {
			tc := tc
			ctx := context.Background()
			store := repos.NewDBStore(ctx, db, sql.TxOptions{Isolation: sql.LevelDefault})

			t.Run(tc.name, transact(ctx, store, func(t testing.TB, tx repos.Store) {
				if err := tx.UpsertRepos(ctx, tc.stored...); err != nil {
					t.Errorf("failed to setup store: %v", err)
					return
				}

				rs, err := tx.ListRepos(ctx, tc.kinds...)
				if have, want := fmt.Sprint(err), fmt.Sprint(tc.err); have != want {
					t.Errorf("error:\nhave: %v\nwant: %v", have, want)
				}

				have := repos.Repos(rs)
				have.Apply(repos.Opt.ID(0)) // Exclude auto-generated IDs from equality tests
				if want := tc.repos; !reflect.DeepEqual(have, want) {
					t.Errorf("repos: %s", cmp.Diff(have, want))
				}
			}))
		}
	}
}

func testDBStoreGetRepoByName(db *sql.DB) func(*testing.T) {
	foo := repos.Repo{
		Name: "github.com/foo/bar",
		Sources: map[string]*repos.SourceInfo{
			"extsvc:123": {
				ID:       "extsvc:123",
				CloneURL: "git@github.com:foo/bar.git",
			},
		},
		Metadata: new(github.Repository),
		ExternalRepo: api.ExternalRepoSpec{
			ServiceType: "github",
			ServiceID:   "https://github.com/",
			ID:          "bar",
		},
	}

	return func(t *testing.T) {
		t.Helper()

		for _, tc := range []struct {
			test   string
			name   string
			stored repos.Repos
			repo   *repos.Repo
			err    error
		}{
			{
				test: "no results error",
				name: "intergalatical repo lost in spaaaaaace",
				err:  repos.ErrNoResults,
			},
			{
				test:   "success",
				stored: repos.Repos{foo.Clone()},
				name:   foo.Name,
				repo:   foo.Clone(),
			},
		} {
			// NOTE: We use t.Errorf instead of t.Fatalf in order to run defers.

			tc := tc
			ctx := context.Background()
			store := repos.NewDBStore(ctx, db, sql.TxOptions{Isolation: sql.LevelDefault})

			t.Run(tc.test, transact(ctx, store, func(t testing.TB, tx repos.Store) {
				if err := tx.UpsertRepos(ctx, tc.stored...); err != nil {
					t.Errorf("failed to setup store: %v", err)
					return
				}

				repo, err := tx.GetRepoByName(ctx, tc.name)
				if have, want := fmt.Sprint(err), fmt.Sprint(tc.err); have != want {
					t.Errorf("error:\nhave: %v\nwant: %v", have, want)
				}

				if repo != nil {
					repo.ID = 0 // Exclude auto-generated IDs from equality tests
				}

				if have, want := repo, tc.repo; !reflect.DeepEqual(have, want) {
					t.Errorf("repos: %s", cmp.Diff(have, want))
				}
			}))
		}
	}
}

func testDBStoreTransact(db *sql.DB) func(*testing.T) {
	return func(t *testing.T) {
		ctx := context.Background()
		store := repos.NewDBStore(ctx, db, sql.TxOptions{Isolation: sql.LevelDefault})

		txstore, err := store.Transact(ctx)
		if err != nil {
			t.Fatal("expected DBStore to support transactions", err)
		}
		defer txstore.Done()

		_, err = txstore.(repos.Transactor).Transact(ctx)
		have := fmt.Sprintf("%s", err)
		want := "dbstore: already in a transaction"
		if have != want {
			t.Errorf("error:\nhave: %v\nwant: %v", have, want)
		}
	}
}

func transact(ctx context.Context, s repos.Store, test func(testing.TB, repos.Store)) func(*testing.T) {
	return func(t *testing.T) {
		t.Helper()

		tr, ok := s.(repos.Transactor)

		if ok {
			txstore, err := tr.Transact(ctx)
			if err != nil {
				// NOTE: We use t.Errorf instead of t.Fatalf in order to run defers.
				t.Errorf("failed to start transaction: %v", err)
				return
			}
			defer txstore.Done(&errRollback)
			s = &noopTxStore{Store: txstore}
		}

		test(t, s)
	}
}

type noopTxStore struct {
	repos.Store
	count int
}

func (tx *noopTxStore) Transact(context.Context) (repos.TxStore, error) {
	if tx.count != 0 {
		return nil, fmt.Errorf("noopTxStore: %d current transactions", tx.count)
	}
	tx.count++
	// noop
	return tx, nil
}

func (tx *noopTxStore) Done(errs ...*error) {
	if tx.count != 1 {
		panic("no current transactions")
	}
	if len(errs) > 0 && *errs[0] != nil {
		panic("unexpected error in noopTxStore")
	}
	tx.count--
}
