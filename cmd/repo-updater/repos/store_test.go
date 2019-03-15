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

func testDBStoreListExternalServices(db *sql.DB) func(*testing.T) {
	clock := repos.NewFakeClock(time.Now(), 0)
	now := clock.Now()

	github := repos.ExternalService{
		Kind:        "github",
		DisplayName: "Github - Test",
		Config:      `{"url": "https://github.com"}`,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	equal := func(es ...*repos.ExternalService) func(testing.TB, repos.ExternalServices) {
		want := repos.ExternalServices(es)
		return func(t testing.TB, have repos.ExternalServices) {
			have.Apply(repos.Opt.ExternalServiceID(0)) // Exclude auto-generated IDs from equality tests
			if !reflect.DeepEqual(have, want) {
				t.Errorf("external services: %s", cmp.Diff(have, want))
			}
		}
	}

	orderedBy := func(ord func(a, b *repos.ExternalService) bool) func(testing.TB, repos.ExternalServices) {
		return func(t testing.TB, have repos.ExternalServices) {
			want := have.Clone()
			sort.Slice(want, func(i, j int) bool {
				return ord(want[i], want[j])
			})
			if !reflect.DeepEqual(have, want) {
				t.Errorf("external services: %s", cmp.Diff(have, want))
			}
		}
	}

	return func(t *testing.T) {
		t.Helper()

		for _, tc := range []struct {
			name   string
			kinds  []string
			stored repos.ExternalServices
			assert func(testing.TB, repos.ExternalServices)
			err    error
		}{
			{
				name:   "case-insensitive kinds",
				kinds:  []string{"GiThUb"},
				stored: repos.ExternalServices{&github},
				assert: equal(&github),
			},
			{
				name:  "returns soft deleted external services",
				kinds: []string{"github"},
				stored: repos.ExternalServices{
					github.With(repos.Opt.ExternalServiceDeletedAt(now)),
				},
				assert: equal(github.With(repos.Opt.ExternalServiceDeletedAt(now))),
			},
			{
				name:   "results are in ascending order by id",
				kinds:  []string{"github"},
				stored: mkExternalServices(512, &github),
				assert: orderedBy(func(a, b *repos.ExternalService) bool {
					return a.ID < b.ID
				}),
			},
		} {
			tc := tc
			ctx := context.Background()
			store := repos.NewDBStore(ctx, db, sql.TxOptions{Isolation: sql.LevelDefault})

			t.Run(tc.name, transact(ctx, store, func(t testing.TB, tx repos.Store) {
				if err := tx.UpsertExternalServices(ctx, tc.stored.Clone()...); err != nil {
					t.Errorf("failed to setup store: %v", err)
					return
				}

				es, err := tx.ListExternalServices(ctx, tc.kinds...)
				if have, want := fmt.Sprint(err), fmt.Sprint(tc.err); have != want {
					t.Errorf("error:\nhave: %v\nwant: %v", have, want)
				}

				if tc.assert != nil {
					tc.assert(t, es)
				}
			}))
		}
	}
}

func testDBStoreUpsertExternalServices(db *sql.DB) func(*testing.T) {
	clock := repos.NewFakeClock(time.Now(), 0)
	now := clock.Now()

	return func(t *testing.T) {
		t.Helper()

		kinds := []string{
			"github",
		}

		ctx := context.Background()
		store := repos.NewDBStore(ctx, db, sql.TxOptions{Isolation: sql.LevelSerializable})

		t.Run("no external services", func(t *testing.T) {
			if err := store.UpsertExternalServices(ctx); err != nil {
				t.Fatalf("UpsertExternalServices error: %s", err)
			}
		})

		t.Run("many external services", transact(ctx, store, func(t testing.TB, tx repos.Store) {
			// Test more than one page load
			want := mkExternalServices(512, &repos.ExternalService{
				Kind:        "github",
				DisplayName: "Github - Test",
				Config:      `{"url": "https://github.com"}`,
				CreatedAt:   now,
				UpdatedAt:   now,
			})

			if err := tx.UpsertExternalServices(ctx, want...); err != nil {
				t.Errorf("UpsertExternalServices error: %s", err)
				return
			}

			sort.Sort(want)

			have, err := tx.ListExternalServices(ctx, kinds...)
			if err != nil {
				t.Errorf("ListExternalServices error: %s", err)
				return
			}

			if diff := pretty.Compare(have, want); diff != "" {
				t.Errorf("ListExternalServices:\n%s", diff)
				return
			}

			now := clock.Now()
			suffix := "-updated"
			for _, r := range want {
				r.DisplayName += suffix
				r.Kind += suffix
				r.Config += suffix
				r.UpdatedAt = now
				r.CreatedAt = now
			}

			if err = tx.UpsertExternalServices(ctx, want.Clone()...); err != nil {
				t.Errorf("UpsertExternalServices error: %s", err)
			} else if have, err = tx.ListExternalServices(ctx); err != nil {
				t.Errorf("ListExternalServices error: %s", err)
			} else if diff := pretty.Compare(have, want); diff != "" {
				t.Errorf("ListExternalServices:\n%s", diff)
			}

			want.Apply(repos.Opt.ExternalServiceDeletedAt(now))

			if err = tx.UpsertExternalServices(ctx, want.Clone()...); err != nil {
				t.Errorf("UpsertExternalServices error: %s", err)
			} else if have, err = tx.ListExternalServices(ctx); err != nil {
				t.Errorf("ListExternalServices error: %s", err)
			} else if diff := pretty.Compare(have, want); diff != "" {
				t.Errorf("ListExternalServices:\n%s", diff)
			}
		}))
	}
}

func testDBStoreUpsertRepos(db *sql.DB) func(*testing.T) {
	clock := repos.NewFakeClock(time.Now(), 0)
	now := clock.Now()

	return func(t *testing.T) {
		t.Helper()

		kinds := []string{
			"github",
		}

		ctx := context.Background()
		store := repos.NewDBStore(ctx, db, sql.TxOptions{Isolation: sql.LevelSerializable})

		t.Run("no repos", func(t *testing.T) {
			if err := store.UpsertRepos(ctx); err != nil {
				t.Fatalf("UpsertRepos error: %s", err)
			}
		})

		t.Run("many repos", transact(ctx, store, func(t testing.TB, tx repos.Store) {
			// Test more than one page load
			want := mkRepos(512, &repos.Repo{
				Name:        "github.com/foo/bar",
				Description: "The description",
				Language:    "barlang",
				Enabled:     true,
				Archived:    false,
				Fork:        false,
				CreatedAt:   now,
				ExternalRepo: api.ExternalRepoSpec{
					ID:          "AAAAA==",
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
			now := clock.Now()
			for _, r := range want {
				r.Name += suffix
				r.Description += suffix
				r.Language += suffix
				r.UpdatedAt = now
				r.CreatedAt = now
				r.Archived = !r.Archived
				r.Fork = !r.Fork
			}

			if err = tx.UpsertRepos(ctx, want.Clone()...); err != nil {
				t.Errorf("UpsertRepos error: %s", err)
			} else if have, err = tx.ListRepos(ctx); err != nil {
				t.Errorf("ListRepos error: %s", err)
			} else if diff := pretty.Compare(have, want); diff != "" {
				t.Errorf("ListRepos:\n%s", diff)
			}

			want.Apply(repos.Opt.RepoDeletedAt(now))

			if err = tx.UpsertRepos(ctx, want.Clone()...); err != nil {
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
	clock := repos.NewFakeClock(time.Now(), 0)
	now := clock.Now()

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

	equal := func(rs ...*repos.Repo) func(testing.TB, repos.Repos) {
		want := repos.Repos(rs)
		return func(t testing.TB, have repos.Repos) {
			have.Apply(repos.Opt.RepoID(0)) // Exclude auto-generated IDs from equality tests
			if !reflect.DeepEqual(have, want) {
				t.Errorf("repos: %s", cmp.Diff(have, want))
			}
		}
	}

	orderedBy := func(ord func(a, b *repos.Repo) bool) func(testing.TB, repos.Repos) {
		return func(t testing.TB, have repos.Repos) {
			want := have.Clone()
			sort.Slice(want, func(i, j int) bool {
				return ord(want[i], want[j])
			})
			if !reflect.DeepEqual(have, want) {
				t.Errorf("repos: %s", cmp.Diff(have, want))
			}
		}
	}

	return func(t *testing.T) {
		t.Helper()

		for _, tc := range []struct {
			name   string
			kinds  []string
			stored repos.Repos
			repos  func(testing.TB, repos.Repos)
			err    error
		}{
			{
				name:  "case-insensitive kinds",
				kinds: []string{"GiThUb"},
				stored: repos.Repos{foo.With(func(r *repos.Repo) {
					r.ExternalRepo.ServiceType = "gItHuB"
				})},
				repos: equal(foo.With(func(r *repos.Repo) {
					r.ExternalRepo.ServiceType = "gItHuB"
				})),
			},
			{
				name:   "ignores unmanaged",
				kinds:  []string{"github"},
				stored: repos.Repos{&foo, &unmanaged}.Clone(),
				repos:  equal(&foo),
			},
			{
				name:   "returns soft deleted repos",
				kinds:  []string{"github"},
				stored: repos.Repos{foo.With(repos.Opt.RepoDeletedAt(now))},
				repos:  equal(foo.With(repos.Opt.RepoDeletedAt(now))),
			},
			{
				name:   "returns repos in ascending order by id",
				kinds:  []string{"github"},
				stored: mkRepos(512, &foo),
				repos: orderedBy(func(a, b *repos.Repo) bool {
					return a.ID < b.ID
				}),
			},
		} {
			tc := tc
			ctx := context.Background()
			store := repos.NewDBStore(ctx, db, sql.TxOptions{Isolation: sql.LevelDefault})

			t.Run(tc.name, transact(ctx, store, func(t testing.TB, tx repos.Store) {
				if err := tx.UpsertRepos(ctx, tc.stored.Clone()...); err != nil {
					t.Errorf("failed to setup store: %v", err)
					return
				}

				rs, err := tx.ListRepos(ctx, tc.kinds...)
				if have, want := fmt.Sprint(err), fmt.Sprint(tc.err); have != want {
					t.Errorf("error:\nhave: %v\nwant: %v", have, want)
				}

				if tc.repos != nil {
					tc.repos(t, rs)
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

func mkRepos(n int, base *repos.Repo) repos.Repos {
	rs := make(repos.Repos, 0, n)
	for i := 0; i < n; i++ {
		id := strconv.Itoa(i)
		r := base.Clone()
		r.Name += id
		r.ExternalRepo.ID += id
		rs = append(rs, r)
	}
	return rs
}

func mkExternalServices(n int, base *repos.ExternalService) repos.ExternalServices {
	es := make(repos.ExternalServices, 0, n)
	for i := 0; i < n; i++ {
		id := strconv.Itoa(i)
		r := base.Clone()
		r.DisplayName += id
		es = append(es, r)
	}
	return es
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
