package repos_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gitchander/permutation"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/github"
)

func TestSyncer_Sync(t *testing.T) {
	t.Parallel()
	testSyncerSync(new(repos.FakeStore))
}

func testSyncerSync(s repos.Store) func(*testing.T) {
	foo := repos.Repo{
		Name:     "foo",
		Metadata: &github.Repository{},
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "foo-external-12345",
			ServiceID:   "https://github.com/",
			ServiceType: "github",
		},
	}

	type testCase struct {
		name    string
		sourcer repos.Sourcer
		store   repos.Store
		stored  repos.Repos
		ctx     context.Context
		now     func() time.Time
		diff    repos.Diff
		err     string
	}

	testCases := []testCase{
		{
			name:    "sourcer error aborts sync",
			sourcer: repos.NewFakeSourcer(errors.New("boom")),
			err:     "syncer.sync.sourced: boom",
		},
		{
			name: "sources partial errors aborts sync",
			sourcer: repos.NewFakeSourcer(nil,
				repos.NewFakeSource("a", "github", nil, foo.Clone()),
				repos.NewFakeSource("b", "github", errors.New("boom")),
			),
			err: "syncer.sync.sourced: 1 error occurred:\n\t* boom\n\n",
		},
		{
			name:    "store list error aborts sync",
			sourcer: repos.NewFakeSourcer(nil, repos.NewFakeSource("a", "github", nil, foo.Clone())),
			store:   &repos.FakeStore{ListReposError: errors.New("boom")},
			err:     "syncer.sync.store.list-repos: boom",
		},
		{
			name:    "store upsert error aborts sync",
			sourcer: repos.NewFakeSourcer(nil, repos.NewFakeSource("a", "github", nil, foo.Clone())),
			store:   &repos.FakeStore{UpsertReposError: errors.New("booya")},
			err:     "syncer.sync.store.upsert-repos: booya",
		},
	}

	return func(t *testing.T) {
		t.Helper()

		{
			clock := repos.NewFakeClock(time.Now(), time.Second)
			testCases = append(testCases, testCase{
				name:    "new repo",
				sourcer: repos.NewFakeSourcer(nil, repos.NewFakeSource("a", "github", nil, foo.Clone())),
				store:   s,
				stored:  repos.Repos{},
				now:     clock.Now,
				diff: repos.Diff{Added: repos.Repos{
					foo.With(repos.Opt.CreatedAt(clock.Time(1)), repos.Opt.Sources("a")),
				}},
				err: "<nil>",
			})
		}

		{
			clock := repos.NewFakeClock(time.Now(), time.Second)
			testCases = append(testCases, testCase{
				name:    "had name and got external_id",
				sourcer: repos.NewFakeSourcer(nil, repos.NewFakeSource("a", "github", nil, foo.Clone())),
				store:   s,
				stored: repos.Repos{foo.With(repos.Opt.Sources("a"), func(r *repos.Repo) {
					r.ExternalRepo.ID = ""
				})},
				now: clock.Now,
				diff: repos.Diff{Modified: repos.Repos{
					foo.With(repos.Opt.ModifiedAt(clock.Time(1)), repos.Opt.Sources("a")),
				}},
				err: "<nil>",
			})
		}

		{
			clock := repos.NewFakeClock(time.Now(), time.Second)
			testCases = append(testCases, testCase{
				name: "new repo sources",
				sourcer: repos.NewFakeSourcer(nil,
					repos.NewFakeSource("a", "github", nil, foo.Clone()),
					repos.NewFakeSource("b", "github", nil, foo.Clone()),
				),
				store:  s,
				stored: repos.Repos{foo.With(repos.Opt.Sources("a"))},
				now:    clock.Now,
				diff: repos.Diff{Modified: repos.Repos{
					foo.With(repos.Opt.ModifiedAt(clock.Time(1)), repos.Opt.Sources("a", "b")),
				}},
				err: "<nil>",
			})
		}

		{
			clock := repos.NewFakeClock(time.Now(), time.Second)
			testCases = append(testCases, testCase{
				name: "enabled field is not updateable",
				sourcer: repos.NewFakeSourcer(nil, repos.NewFakeSource("a", "github", nil, foo.With(func(r *repos.Repo) {
					r.Enabled = !r.Enabled
				}))),
				store:  s,
				stored: repos.Repos{foo.With(repos.Opt.Sources("a"))},
				now:    clock.Now,
				diff:   repos.Diff{Unmodified: repos.Repos{foo.With(repos.Opt.Sources("a"))}},
				err:    "<nil>",
			})
		}

		{
			clock := repos.NewFakeClock(time.Now(), time.Second)
			testCases = append(testCases, testCase{
				name: "enabled field of a undeleted repo is not updateable",
				sourcer: repos.NewFakeSourcer(nil, repos.NewFakeSource("a", "github", nil, foo.With(func(r *repos.Repo) {
					r.Enabled = !r.Enabled
				}))),
				store:  s,
				stored: repos.Repos{foo.With(repos.Opt.Sources("a"), repos.Opt.DeletedAt(clock.Time(0)))},
				now:    clock.Now,
				diff: repos.Diff{Added: repos.Repos{
					foo.With(repos.Opt.Sources("a"), repos.Opt.CreatedAt(clock.Time(1))),
				}},
				err: "<nil>",
			})
		}

		{
			clock := repos.NewFakeClock(time.Now(), time.Second)
			testCases = append(testCases, testCase{
				name: "deleted repo source",
				sourcer: repos.NewFakeSourcer(nil,
					repos.NewFakeSource("a", "github", nil, foo.Clone()),
				),
				store:  s,
				stored: repos.Repos{foo.With(repos.Opt.Sources("a", "b"))},
				now:    clock.Now,
				diff: repos.Diff{Modified: repos.Repos{
					foo.With(repos.Opt.ModifiedAt(clock.Time(1)), repos.Opt.Sources("a")),
				}},
				err: "<nil>",
			})
		}

		{
			clock := repos.NewFakeClock(time.Now(), time.Second)
			testCases = append(testCases, testCase{
				name:    "deleted ALL repo sources",
				sourcer: repos.NewFakeSourcer(nil),
				store:   s,
				stored:  repos.Repos{foo.With(repos.Opt.Sources("a", "b"))},
				now:     clock.Now,
				diff: repos.Diff{Deleted: repos.Repos{
					foo.With(repos.Opt.DeletedAt(clock.Time(1))),
				}},
				err: "<nil>",
			})
		}

		{
			clock := repos.NewFakeClock(time.Now(), time.Second)
			testCases = append(testCases, testCase{
				name:    "renamed repo is detected via external_id",
				sourcer: repos.NewFakeSourcer(nil, repos.NewFakeSource("a", "github", nil, foo.Clone())),
				store:   s,
				stored: repos.Repos{foo.With(repos.Opt.Sources("a"), func(r *repos.Repo) {
					r.Name = "old-name"
				})},
				now: clock.Now,
				diff: repos.Diff{Modified: repos.Repos{
					foo.With(repos.Opt.Sources("a"), repos.Opt.ModifiedAt(clock.Time(1))),
				}},
				err: "<nil>",
			})
		}

		{
			clock := repos.NewFakeClock(time.Now(), time.Second)
			testCases = append(testCases, testCase{
				name:    "renamed repo which was deleted is detected and added",
				sourcer: repos.NewFakeSourcer(nil, repos.NewFakeSource("a", "github", nil, foo.Clone())),
				store:   s,
				stored: repos.Repos{foo.With(func(r *repos.Repo) {
					r.Sources = map[string]*repos.SourceInfo{}
					r.Name = "old-name"
					r.DeletedAt = clock.Time(0)
				})},
				now: clock.Now,
				diff: repos.Diff{Added: repos.Repos{
					foo.With(repos.Opt.Sources("a"), repos.Opt.CreatedAt(clock.Time(1))),
				}},
				err: "<nil>",
			})
		}

		{
			clock := repos.NewFakeClock(time.Now(), time.Second)
			testCases = append(testCases, testCase{
				name: "metadata update",
				sourcer: repos.NewFakeSourcer(nil, repos.NewFakeSource("a", "github", nil,
					foo.With(repos.Opt.ModifiedAt(clock.Time(1)),
						repos.Opt.Sources("a"),
						repos.Opt.Metadata(&github.Repository{IsArchived: true})),
				)),
				store:  s,
				stored: repos.Repos{foo.With(repos.Opt.Sources("a"))},
				now:    clock.Now,
				diff: repos.Diff{Modified: repos.Repos{
					foo.With(repos.Opt.ModifiedAt(clock.Time(1)),
						repos.Opt.Sources("a"),
						repos.Opt.Metadata(&github.Repository{IsArchived: true})),
				}},
				err: "<nil>",
			})
		}

		kinds := []string{"github"}

		for _, tc := range testCases {
			tc := tc
			ctx := context.Background()

			t.Run(tc.name, transact(ctx, tc.store, func(t testing.TB, st repos.Store) {
				defer func() {
					if err := recover(); err != nil {
						t.Fatalf("%q panicked: %v", tc.name, err)
					}
				}()

				now := tc.now
				if now == nil {
					clock := repos.NewFakeClock(time.Now(), time.Second)
					now = clock.Now
				}

				ctx := tc.ctx
				if ctx == nil {
					ctx = context.Background()
				}

				if st != nil && len(tc.stored) > 0 {
					if err := st.UpsertRepos(ctx, tc.stored...); err != nil {
						t.Fatalf("failed to prepare store: %v", err)
					}
				}

				syncer := repos.NewSyncer(st, tc.sourcer, nil, now)
				diff, err := syncer.Sync(ctx, kinds...)

				var want repos.Repos
				want.Concat(diff.Added, diff.Modified, diff.Unmodified, diff.Deleted)
				want.Apply(repos.Opt.ID(0)) // Exclude auto-generated ID from comparisons

				if have, want := fmt.Sprint(err), tc.err; have != want {
					t.Errorf("have error %q, want %q", have, want)
				}
				if err != nil {
					return
				}

				if diff := cmp.Diff(diff, tc.diff); diff != "" {
					// t.Logf("have: %s\nwant: %s\n", pp.Sprint(have), pp.Sprint(want))
					t.Fatalf("unexpected diff:\n%s", diff)
				}

				if st != nil {
					have, _ := st.ListRepos(ctx)
					for _, d := range have {
						d.ID = 0 // Exclude auto-generated ID from comparisons
					}
					if diff := cmp.Diff(repos.Repos(have), want); diff != "" {
						// t.Logf("have: %s\nwant: %s\n", pp.Sprint(have), pp.Sprint(want))
						t.Fatalf("unexpected stored repos:\n%s", diff)
					}
				}
			}))
		}
	}
}

func TestDiff(t *testing.T) {
	t.Parallel()

	eid := func(id string) api.ExternalRepoSpec {
		return api.ExternalRepoSpec{
			ID: id,
		}
	}
	now := time.Now()

	type testCase struct {
		name   string
		store  repos.Repos
		source repos.Repos
		diff   repos.Diff
	}

	cases := []testCase{
		{
			name: "empty",
			diff: repos.Diff{},
		},
		{
			name:   "added",
			source: repos.Repos{{ExternalRepo: eid("1")}},
			diff:   repos.Diff{Added: repos.Repos{{ExternalRepo: eid("1")}}},
		},
		{
			name:  "deleted",
			store: repos.Repos{{ExternalRepo: eid("1")}},
			diff:  repos.Diff{Deleted: repos.Repos{{ExternalRepo: eid("1")}}},
		},
		{
			name:   "modified",
			store:  repos.Repos{{ExternalRepo: eid("1"), Description: "foo"}},
			source: repos.Repos{{ExternalRepo: eid("1"), Description: "bar"}},
			diff: repos.Diff{Modified: repos.Repos{
				{ExternalRepo: eid("1"), Description: "bar"},
			}},
		},
		{
			name:   "unmodified",
			store:  repos.Repos{{ExternalRepo: eid("1"), Description: "foo"}},
			source: repos.Repos{{ExternalRepo: eid("1"), Description: "foo"}},
			diff: repos.Diff{Unmodified: repos.Repos{
				{ExternalRepo: eid("1"), Description: "foo"},
			}},
		},
		{
			name: "duplicates in source are merged",
			source: repos.Repos{
				{ExternalRepo: eid("1"), Description: "foo", Sources: map[string]*repos.SourceInfo{
					"a": {ID: "a"},
				}},
				{ExternalRepo: eid("1"), Description: "bar", Sources: map[string]*repos.SourceInfo{
					"b": {ID: "b"},
				}},
			},
			diff: repos.Diff{Added: repos.Repos{
				{ExternalRepo: eid("1"), Description: "bar", Sources: map[string]*repos.SourceInfo{
					"a": {ID: "a"},
					"b": {ID: "b"},
				}},
			}},
		},
		{
			name: "duplicate with a changed name is merged correctly",
			store: repos.Repos{
				{Name: "1", ExternalRepo: eid("1"), Description: "foo"},
			},
			source: repos.Repos{
				{Name: "2", ExternalRepo: eid("1"), Description: "foo"},
			},
			diff: repos.Diff{Modified: repos.Repos{
				{Name: "2", ExternalRepo: eid("1"), Description: "foo"},
			}},
		},
		{
			name: "duplicate with added external id is merged correctly",
			store: repos.Repos{
				{Name: "1", Description: "foo"},
			},
			source: repos.Repos{
				{Name: "1", ExternalRepo: eid("1"), Description: "foo"},
			},
			diff: repos.Diff{Modified: repos.Repos{
				{Name: "1", ExternalRepo: eid("1"), Description: "foo"},
			}},
		},
		{
			name: "no duplicate with added external id and changed name",
			store: repos.Repos{
				{Name: "1", Description: "foo"},
			},
			source: repos.Repos{
				{Name: "2", ExternalRepo: eid("1"), Description: "foo"},
			},
			diff: repos.Diff{
				Deleted: repos.Repos{
					{Name: "1", Description: "foo"},
				},
				Added: repos.Repos{
					{Name: "2", ExternalRepo: eid("1"), Description: "foo"},
				},
			},
		},
		{
			name: "unmodified preserves stored repo",
			store: repos.Repos{
				{ExternalRepo: eid("1"), Description: "foo", UpdatedAt: now},
			},
			source: repos.Repos{
				{ExternalRepo: eid("1"), Description: "foo"},
			},
			diff: repos.Diff{Unmodified: repos.Repos{
				{ExternalRepo: eid("1"), Description: "foo", UpdatedAt: now},
			}},
		},
	}

	permutedCases := []testCase{
		{
			// Repo renamed and repo created with old name
			name: "renamed repo",
			store: repos.Repos{
				{Name: "1", ExternalRepo: eid("old"), Description: "foo"},
			},
			source: repos.Repos{
				{Name: "2", ExternalRepo: eid("old"), Description: "foo"},
				{Name: "1", ExternalRepo: eid("new"), Description: "bar"},
			},
			diff: repos.Diff{
				Modified: repos.Repos{
					{Name: "2", ExternalRepo: eid("old"), Description: "foo"},
				},
				Added: repos.Repos{
					{Name: "1", ExternalRepo: eid("new"), Description: "bar"},
				},
			},
		},
		{
			name: "swapped repo",
			store: repos.Repos{
				{Name: "foo", ExternalRepo: eid("1"), Description: "foo"},
				{Name: "bar", ExternalRepo: eid("2"), Description: "bar"},
			},
			source: repos.Repos{
				{Name: "bar", ExternalRepo: eid("1"), Description: "bar"},
				{Name: "foo", ExternalRepo: eid("2"), Description: "foo"},
			},
			diff: repos.Diff{
				Modified: repos.Repos{
					{Name: "bar", ExternalRepo: eid("1"), Description: "bar"},
					{Name: "foo", ExternalRepo: eid("2"), Description: "foo"},
				},
			},
		},
	}

	for _, tc := range permutedCases {
		store := permutation.New(tc.store)
		for i := 0; store.Next(); i++ {
			source := permutation.New(tc.source)
			for j := 0; source.Next(); j++ {
				cases = append(cases, testCase{
					name:   fmt.Sprintf("%s/permutation_%d_%d", tc.name, i, j),
					store:  tc.store.Clone(),
					source: tc.source.Clone(),
					diff:   tc.diff,
				})
			}
		}
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			diff := repos.NewDiff(tc.source, tc.store)
			diff.Sort()
			tc.diff.Sort()
			if cDiff := cmp.Diff(diff, tc.diff); cDiff != "" {
				//t.Logf("have: %s\nwant: %s\n", pp.Sprint(diff), pp.Sprint(tc.diff))
				t.Fatalf("unexpected diff:\n%s", cDiff)
			}
		})
	}
}
