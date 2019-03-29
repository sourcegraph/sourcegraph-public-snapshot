package repos_test

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/kylelemons/godebug/pretty"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/github"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/gitlab"
)

func TestFakeStore(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		test func(*testing.T)
	}{
		{"ListExternalServices", testStoreListExternalServices(new(repos.FakeStore))},
		{"UpsertExternalServices", testStoreUpsertExternalServices(new(repos.FakeStore))},
		{"GetRepoByName", testStoreGetRepoByName(new(repos.FakeStore))},
		{"UpsertRepos", testStoreUpsertRepos(new(repos.FakeStore))},
		{"ListRepos", testStoreListRepos(new(repos.FakeStore))},
	} {
		t.Run(tc.name, tc.test)
	}
}

func testStoreListExternalServices(store repos.Store) func(*testing.T) {
	clock := repos.NewFakeClock(time.Now(), 0)
	now := clock.Now()

	github := repos.ExternalService{
		Kind:        "GITHUB",
		DisplayName: "Github - Test",
		Config:      `{"url": "https://github.com"}`,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	gitlab := repos.ExternalService{
		Kind:        "GITLAB",
		DisplayName: "GitLab - Test",
		Config:      `{"url": "https://gitlab.com"}`,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	svcs := repos.ExternalServices{
		&github,
		&gitlab,
	}

	return func(t *testing.T) {
		t.Helper()

		for _, tc := range []struct {
			name   string
			kinds  []string
			stored repos.ExternalServices
			assert repos.ExternalServicesAssertion
			err    error
		}{
			{
				name:   "returned kind is uppercase",
				kinds:  svcs.Kinds(),
				stored: repos.ExternalServices{&github, &gitlab},
				assert: repos.Assert.ExternalServicesEqual(&github, &gitlab),
			},
			{
				name:   "case-insensitive kinds",
				kinds:  []string{"GiThUb", "GitLab"},
				stored: repos.ExternalServices{&github, &gitlab},
				assert: repos.Assert.ExternalServicesEqual(&github, &gitlab),
			},
			{
				name:  "returns soft deleted external services",
				kinds: svcs.Kinds(),
				stored: repos.ExternalServices{
					github.With(repos.Opt.ExternalServiceDeletedAt(now)),
				},
				assert: repos.Assert.ExternalServicesEqual(
					github.With(repos.Opt.ExternalServiceDeletedAt(now)),
				),
			},
			{
				name:  "results are in ascending order by id",
				kinds: []string{}, // All kinds
				stored: append(
					mkExternalServices(256, &github),
					mkExternalServices(256, &gitlab)...,
				),
				assert: repos.Assert.ExternalServicesOrderedBy(
					func(a, b *repos.ExternalService) bool {
						return a.ID < b.ID
					},
				),
			},
		} {
			tc := tc
			ctx := context.Background()

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

func testStoreUpsertExternalServices(store repos.Store) func(*testing.T) {
	clock := repos.NewFakeClock(time.Now(), 0)
	now := clock.Now()

	return func(t *testing.T) {
		t.Helper()

		github := repos.ExternalService{
			Kind:        "GITHUB",
			DisplayName: "Github - Test",
			Config:      `{"url": "https://github.com"}`,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		gitlab := repos.ExternalService{
			Kind:        "GITLAB",
			DisplayName: "GitLab - Test",
			Config:      `{"url": "https://gitlab.com"}`,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		svcs := repos.ExternalServices{
			&github,
			&gitlab,
		}

		ctx := context.Background()

		t.Run("no external services", func(t *testing.T) {
			if err := store.UpsertExternalServices(ctx); err != nil {
				t.Fatalf("UpsertExternalServices error: %s", err)
			}
		})

		t.Run("many external services", transact(ctx, store, func(t testing.TB, tx repos.Store) {
			// Test more than one page load
			want := append(
				mkExternalServices(256, &github),
				mkExternalServices(256, &gitlab)...,
			)

			if err := tx.UpsertExternalServices(ctx, want...); err != nil {
				t.Errorf("UpsertExternalServices error: %s", err)
				return
			}

			for _, e := range want {
				if e.Kind != strings.ToUpper(e.Kind) {
					t.Errorf("external service kind didn't get upper-cased: %q", e.Kind)
					break
				}
			}

			sort.Sort(want)

			have, err := tx.ListExternalServices(ctx, svcs.Kinds()...)
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

			if err = tx.UpsertExternalServices(ctx, want...); err != nil {
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

func testStoreUpsertRepos(store repos.Store) func(*testing.T) {
	clock := repos.NewFakeClock(time.Now(), 0)
	now := clock.Now()

	return func(t *testing.T) {
		t.Helper()

		kinds := []string{
			"github",
			"gitlab",
		}

		github := repos.Repo{
			Name:        "github.com/foo/bar",
			Description: "The description",
			Language:    "barlang",
			Enabled:     true,
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
			Metadata: new(github.Repository),
		}

		gitlab := repos.Repo{
			Name:        "gitlab.com/foo/bar",
			Description: "The description",
			Language:    "barlang",
			Enabled:     true,
			CreatedAt:   now,
			ExternalRepo: api.ExternalRepoSpec{
				ID:          "1234",
				ServiceType: "gitlab",
				ServiceID:   "http://gitlab.com",
			},
			Sources: map[string]*repos.SourceInfo{
				"extsvc:1234": {
					ID:       "extsvc:1234",
					CloneURL: "git@gitlab.com:foo/bar.git",
				},
			},
			Metadata: new(gitlab.Project),
		}

		ctx := context.Background()

		t.Run("no repos", func(t *testing.T) {
			if err := store.UpsertRepos(ctx); err != nil {
				t.Fatalf("UpsertRepos error: %s", err)
			}
		})

		t.Run("many repos", transact(ctx, store, func(t testing.TB, tx repos.Store) {
			// Test more than one page load
			want := append(
				mkRepos(256, &github),
				mkRepos(256, &gitlab)...,
			)

			if err := tx.UpsertRepos(ctx, want...); err != nil {
				t.Errorf("UpsertRepos error: %s", err)
				return
			}

			sort.Sort(want)

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

func testStoreListRepos(store repos.Store) func(*testing.T) {
	clock := repos.NewFakeClock(time.Now(), 0)
	now := clock.Now()

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

	github := repos.Repo{
		Name: "github.com/bar/foo",
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

	gitlab := repos.Repo{
		Name: "gitlab.com/bar/foo",
		Sources: map[string]*repos.SourceInfo{
			"extsvc:123": {
				ID:       "extsvc:123",
				CloneURL: "git@gitlab.com:bar/foo.git",
			},
		},
		Metadata: new(gitlab.Project),
		ExternalRepo: api.ExternalRepoSpec{
			ServiceType: "gitlab",
			ServiceID:   "https://gitlab.com/",
			ID:          "123",
		},
	}

	kinds := []string{
		"github",
		"gitlab",
	}

	type testCase struct {
		name   string
		kinds  []string
		stored repos.Repos
		repos  repos.ReposAssertion
		err    error
	}

	var testCases []testCase
	{
		stored := repos.Repos{
			github.With(func(r *repos.Repo) {
				r.ExternalRepo.ServiceType = "gItHuB"
			}),
			gitlab.With(func(r *repos.Repo) {
				r.ExternalRepo.ServiceType = "GitLab"
			}),
		}

		testCases = append(testCases, testCase{
			name:   "case-insensitive kinds",
			kinds:  []string{"GiThUb", "GitLab"},
			stored: stored,
			repos:  repos.Assert.ReposEqual(stored...),
		})
	}

	testCases = append(testCases, testCase{
		name:   "ignores unmanaged",
		kinds:  kinds,
		stored: repos.Repos{&github, &gitlab, &unmanaged}.Clone(),
		repos:  repos.Assert.ReposEqual(&github, &gitlab),
	})

	{
		stored := (&repos.Repos{&github, &gitlab}).With(repos.Opt.RepoDeletedAt(now))
		testCases = append(testCases, testCase{
			name:   "returns soft deleted repos",
			kinds:  kinds,
			stored: stored,
			repos:  repos.Assert.ReposEqual(stored...),
		})
	}

	testCases = append(testCases, testCase{
		name:  "returns repos in ascending order by id",
		kinds: []string{}, // All kinds
		stored: append(
			mkRepos(256, &github),
			mkRepos(256, &gitlab)...,
		),
		repos: repos.Assert.ReposOrderedBy(func(a, b *repos.Repo) bool {
			return a.ID < b.ID
		}),
	})

	return func(t *testing.T) {
		t.Helper()

		for _, tc := range testCases {
			tc := tc
			ctx := context.Background()

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

func testStoreGetRepoByName(store repos.Store) func(*testing.T) {
	github := repos.Repo{
		Name: "github.com/bar/foo",
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

	gitlab := repos.Repo{
		Name: "gitlab.com/bar/foo",
		Sources: map[string]*repos.SourceInfo{
			"extsvc:123": {
				ID:       "extsvc:123",
				CloneURL: "git@gitlab.com:bar/foo.git",
			},
		},
		Metadata: new(gitlab.Project),
		ExternalRepo: api.ExternalRepoSpec{
			ServiceType: "gitlab",
			ServiceID:   "https://gitlab.com/",
			ID:          "123",
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
				test: "success",
				stored: repos.Repos{
					&github,
					&gitlab,
				},
				name: github.Name,
				repo: github.Clone(),
			},
		} {
			// NOTE: We use t.Errorf instead of t.Fatalf in order to run defers.

			tc := tc
			ctx := context.Background()

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

func testDBStoreTransact(store *repos.DBStore) func(*testing.T) {
	return func(t *testing.T) {
		ctx := context.Background()

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
			s = &noopTxStore{TB: t, Store: txstore}
		}

		test(t, s)
	}
}

type noopTxStore struct {
	testing.TB
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
	tx.Helper()

	if tx.count != 1 {
		tx.Fatal("no current transactions")
	}
	if len(errs) > 0 && *errs[0] != nil {
		tx.Fatal(fmt.Sprintf("unexpected error in noopTxStore: %v", *errs[0]))
	}
	tx.count--
}
