package repos_test

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/kylelemons/godebug/pretty"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

func TestSyncer_Sync(t *testing.T) {
	foo := repos.Repo{
		Name: "foo",
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "foo-external-12345",
			ServiceID:   "https://github.com/",
			ServiceType: "github",
		},
	}

	bar := repos.Repo{Name: "bar"}

	type testCase struct {
		name    string
		sourcer repos.Sourcer
		store   repos.Store
		ctx     context.Context
		now     func() time.Time
		diff    repos.Diff
		err     string
	}

	testCases := []testCase{
		{
			name:    "sourcer error aborts sync",
			sourcer: sourcer(errors.New("boom")),
			err:     "boom",
		},
		{
			name: "sources partial errors aborts sync",
			sourcer: sourcer(nil,
				source("a", nil, foo.Clone()),
				source("b", errors.New("boom")),
			),
			err: "1 error occurred:\n\t* boom\n\n",
		},
		{
			name:    "store list error aborts sync",
			sourcer: sourcer(nil, source("a", nil, foo.Clone())),
			store:   &fakeStore{list: errors.New("boom")},
			err:     "boom",
		},
		{
			name:    "store upsert error aborts sync",
			sourcer: sourcer(nil, source("a", nil, foo.Clone())),
			store:   &fakeStore{upsert: errors.New("booya")},
			err:     "booya",
		},
		{
			// here we test that if a repo previously had the external id set and one
			// of its sources gets rate limited, we preserve the external id
			name: "had name and external id, got name only from rate-limited source",
			sourcer: sourcer(nil, source("a", nil, foo.With(func(r *repos.Repo) {
				r.ExternalRepo = api.ExternalRepoSpec{}
			}))),
			store: store(foo.With(sources("a"))),
			diff:  repos.Diff{Unmodified: repos.Repos{foo.With(sources("a"))}},
			err:   "<nil>",
		},
	}

	{
		clock := fakeClock{epoch: time.Now(), step: time.Second}
		testCases = append(testCases, testCase{
			name:    "had name and got external_id",
			sourcer: sourcer(nil, source("a", nil, foo.Clone())),
			store: store(foo.With(sources("a"), func(r *repos.Repo) {
				r.ExternalRepo = api.ExternalRepoSpec{}
			})),
			now: clock.Now,
			diff: repos.Diff{Modified: repos.Repos{
				foo.With(modifiedAt(clock.Time(1)), sources("a")),
			}},
			err: "<nil>",
		})
	}

	{
		clock := fakeClock{epoch: time.Now(), step: time.Second}
		testCases = append(testCases, testCase{
			name: "new repo sources",
			sourcer: sourcer(nil,
				source("a", nil, foo.Clone()),
				source("b", nil, foo.Clone()),
			),
			store: store(foo.With(sources("a"))),
			now:   clock.Now,
			diff: repos.Diff{Modified: repos.Repos{
				foo.With(modifiedAt(clock.Time(1)), sources("a", "b")),
			}},
			err: "<nil>",
		})
	}

	{
		clock := fakeClock{epoch: time.Now(), step: time.Second}
		testCases = append(testCases, testCase{
			name: "deleted repo source",
			sourcer: sourcer(nil,
				source("a", nil, foo.Clone()),
			),
			store: store(foo.With(sources("a", "b"))),
			now:   clock.Now,
			diff: repos.Diff{Modified: repos.Repos{
				foo.With(modifiedAt(clock.Time(1)), sources("a")),
			}},
			err: "<nil>",
		})
	}

	{
		clock := fakeClock{epoch: time.Now(), step: time.Second}
		testCases = append(testCases, testCase{
			name:    "deleted ALL repo sources",
			sourcer: sourcer(nil),
			store:   store(foo.With(sources("a", "b"))),
			now:     clock.Now,
			diff: repos.Diff{Deleted: repos.Repos{
				foo.With(deletedAt(clock.Time(1))),
			}},
			err: "<nil>",
		})
	}

	{
		clock := fakeClock{epoch: time.Now(), step: time.Second}
		testCases = append(testCases, testCase{
			name:    "renamed repo is detected via external_id",
			sourcer: sourcer(nil, source("a", nil, foo.Clone())),
			store: store(foo.With(sources("a"), func(r *repos.Repo) {
				r.Name = "old-name"
			})),
			now: clock.Now,
			diff: repos.Diff{Modified: repos.Repos{
				foo.With(sources("a"), modifiedAt(clock.Time(1))),
			}},
			err: "<nil>",
		})
	}

	{
		clock := fakeClock{epoch: time.Now(), step: time.Second}
		testCases = append(testCases, testCase{
			name:    "renamed repo without external_id is re-created",
			sourcer: sourcer(nil, source("a", nil, bar.Clone())),
			store:   store(foo.With(sources("a"))),
			now:     clock.Now,
			diff: repos.Diff{
				Added: repos.Repos{
					bar.With(sources("a"), createdAt(clock.Time(1))),
				},
				Deleted: repos.Repos{
					foo.With(sources("a"), deletedAt(clock.Time(1))),
				},
			},
			err: "<nil>",
		})
	}

	{
		clock := fakeClock{epoch: time.Now(), step: time.Second}
		testCases = append(testCases, testCase{
			name: "metadata update",
			sourcer: sourcer(nil, source("a", nil,
				foo.With(modifiedAt(clock.Time(1)),
					sources("a"),
					metadata([]byte(`{"metadata": true}`))),
			)),
			store: store(foo.With(sources("a"))),
			now:   clock.Now,
			diff: repos.Diff{Modified: repos.Repos{
				foo.With(modifiedAt(clock.Time(1)),
					sources("a"),
					metadata([]byte(`{"metadata": true}`))),
			}},
			err: "<nil>",
		})
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			now := tc.now
			if now == nil {
				now = time.Now
			}

			syncer := repos.NewSyncer(0, tc.store, tc.sourcer, nil, now)
			diff, err := syncer.Sync(tc.ctx)

			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Errorf("have error %q, want %q", have, want)
			}

			if cmp := pretty.Compare(diff, tc.diff); cmp != "" {
				// t.Logf("have: %s\nwant: %s\n", pp.Sprint(have), pp.Sprint(want))
				t.Fatalf("unexpected diff:\n%s", cmp)
			}

			if tc.store != nil {
				have, _ := tc.store.ListRepos(tc.ctx)

				var want []*repos.Repo
				for _, ds := range []repos.Repos{
					diff.Added,
					diff.Modified,
					diff.Unmodified,
					diff.Deleted,
				} {
					for _, d := range ds {
						want = append(want, d)
					}
				}

				if cmp := pretty.Compare(have, want); cmp != "" {
					// t.Logf("have: %s\nwant: %s\n", pp.Sprint(have), pp.Sprint(want))
					t.Fatalf("unexpected stored repos:\n%s", cmp)
				}
			}
		})
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

	for _, tc := range []struct {
		name   string
		store  repos.Repos
		source repos.Repos
		diff   repos.Diff
	}{
		{
			name: "empty",
			diff: repos.Diff{},
		},
		{
			name:   "added",
			source: repos.Repos{{Name: "1"}},
			diff:   repos.Diff{Added: repos.Repos{{Name: "1"}}},
		},
		{
			name:  "deleted",
			store: repos.Repos{{Name: "1"}},
			diff:  repos.Diff{Deleted: repos.Repos{{Name: "1"}}},
		},
		{
			name:   "modified",
			store:  repos.Repos{{Name: "1", Description: "foo"}},
			source: repos.Repos{{Name: "1", Description: "bar"}},
			diff:   repos.Diff{Modified: repos.Repos{{Name: "1", Description: "bar"}}},
		},
		{
			name:   "unmodified",
			store:  repos.Repos{{Name: "1", Description: "foo"}},
			source: repos.Repos{{Name: "1", Description: "foo"}},
			diff:   repos.Diff{Unmodified: repos.Repos{{Name: "1", Description: "foo"}}},
		},
		{
			name:   "duplicates in source", // first duplicate wins
			source: repos.Repos{{Name: "1", Description: "foo"}, {Name: "1", Description: "bar"}},
			diff:   repos.Diff{Added: repos.Repos{{Name: "1", Description: "foo"}}},
		},
		{
			// This test case is covering the scenario when a repo had a name and got
			// an external_id with the latest sync. In this case, we want to merge those
			// two repos into one with the external_id set.
			name: "duplicate with a second ID is merged correctly",
			store: repos.Repos{
				{Name: "1", Description: "foo"},
			},
			source: repos.Repos{
				{Name: "1", ExternalRepo: eid("second id"), Description: "foo"},
			},
			diff: repos.Diff{Modified: repos.Repos{{Name: "1", ExternalRepo: eid("second id"), Description: "foo"}}},
		},
		{
			name:   "duplicate with a single common ID is merged correctly",
			store:  repos.Repos{{Name: "1", ExternalRepo: eid("second id"), Description: "foo", CreatedAt: now}},
			source: repos.Repos{{Name: "2", ExternalRepo: eid("second id"), Description: "foo"}},
			diff:   repos.Diff{Modified: repos.Repos{{Name: "2", ExternalRepo: eid("second id"), Description: "foo"}}},
		},
		{
			name:   "unmodified preserves before ",
			store:  repos.Repos{{Name: "1", Description: "foo", CreatedAt: now}},
			source: repos.Repos{{Name: "1", Description: "foo"}},
			diff:   repos.Diff{Unmodified: repos.Repos{{Name: "1", Description: "foo", CreatedAt: now}}},
		},
		{
			// Repo renamed and repo created with old name
			name:   "renamed repo",
			store:  repos.Repos{{Name: "1", ExternalRepo: eid("old"), Description: "foo"}},
			source: repos.Repos{{Name: "1", ExternalRepo: eid("new"), Description: "bar"}, {Name: "2", ExternalRepo: eid("old"), Description: "foo"}},
			diff: repos.Diff{
				Modified: repos.Repos{{Name: "2", ExternalRepo: eid("old"), Description: "foo"}},
				Added:    repos.Repos{{Name: "1", ExternalRepo: eid("new"), Description: "bar"}},
			},
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			for _, ds := range []repos.Repos{
				tc.store,
				tc.source,
				tc.diff.Added,
				tc.diff.Modified,
				tc.diff.Unmodified,
				// Exclude Deleted since it will have no source set
			} {
				for i := range ds {
					ds[i].Sources = []string{"a"}
				}
			}

			syncer := repos.NewSyncer(0, store(tc.store...), sourcer(nil, source("a", nil, tc.source...)), nil, func() time.Time { return time.Time{} })
			diff, err := syncer.Sync(context.Background())
			if err != nil {
				t.Fatal(err)
			}

			if cmp := pretty.Compare(diff, tc.diff); cmp != "" {
				// t.Logf("have: %s\nwant: %s\n", pp.Sprint(have), pp.Sprint(want))
				t.Fatalf("unexpected diff:\n%s", cmp)
			}
		})
	}
}

//
// Test utilities
//

type fakeSourcer struct {
	err  error
	srcs []repos.Source
}

func sourcer(err error, srcs ...repos.Source) fakeSourcer {
	return fakeSourcer{err: err, srcs: srcs}
}

func (s fakeSourcer) ListSources(context.Context) ([]repos.Source, error) {
	return s.srcs, s.err
}

type fakeSource struct {
	urn   string
	repos []*repos.Repo
	err   error
}

func source(urn string, err error, rs ...*repos.Repo) fakeSource {
	return fakeSource{urn: urn, err: err, repos: rs}
}

func (s fakeSource) ListRepos(context.Context) ([]*repos.Repo, error) {
	repos := make([]*repos.Repo, len(s.repos))
	for i, r := range s.repos {
		repos[i] = r.With(sources(s.urn))
	}
	return repos, s.err
}

type fakeStore struct {
	repos  map[string]*repos.Repo
	list   error
	upsert error
}

func store(rs ...*repos.Repo) *fakeStore {
	s := fakeStore{repos: make(map[string]*repos.Repo, len(rs))}
	for _, r := range rs {
		for _, id := range r.IDs() {
			s.repos[id] = r
		}
	}
	return &s
}

func (s fakeStore) ListRepos(context.Context, ...string) ([]*repos.Repo, error) {
	if s.list != nil {
		return nil, s.list
	}

	set := make(map[*repos.Repo]bool, len(s.repos))
	repos := make([]*repos.Repo, 0, len(s.repos))
	for _, r := range s.repos {
		if !set[r] {
			repos = append(repos, r)
			set[r] = true
		}
	}

	sort.Slice(repos, func(i, j int) bool {
		return repos[i].Name < repos[j].Name
	})

	return repos, nil
}

func (s *fakeStore) UpsertRepos(_ context.Context, upserts ...*repos.Repo) error {
	if s.upsert != nil {
		return s.upsert
	}

	if s.repos == nil {
		s.repos = make(map[string]*repos.Repo, len(upserts))
	}

	for _, upsert := range upserts {
		var repo *repos.Repo
		for _, id := range upsert.IDs() {
			if repo = s.repos[id]; repo != nil {
				break
			}
		}

		if repo != nil {
			repo.Name = upsert.Name
			repo.Description = upsert.Description
			repo.Language = upsert.Language
			repo.UpdatedAt = upsert.UpdatedAt
			repo.DeletedAt = upsert.DeletedAt
			repo.Archived = upsert.Archived
			repo.Fork = upsert.Fork
			repo.Sources = upsert.Sources
			repo.Metadata = upsert.Metadata
			repo.ExternalRepo = upsert.ExternalRepo
			continue
		}

		for _, id := range upsert.IDs() {
			s.repos[id] = upsert
		}
	}

	return nil
}

//
// Repo functional options
//

func createdAt(ts time.Time) func(*repos.Repo) {
	return func(r *repos.Repo) {
		r.CreatedAt = ts
		r.DeletedAt = time.Time{}
	}
}

func modifiedAt(ts time.Time) func(*repos.Repo) {
	return func(r *repos.Repo) {
		r.UpdatedAt = ts
		r.DeletedAt = time.Time{}
	}
}

func deletedAt(ts time.Time) func(*repos.Repo) {
	return func(r *repos.Repo) {
		r.UpdatedAt = ts
		r.DeletedAt = ts
		r.Sources = []string{}
	}
}

func sources(srcs ...string) func(*repos.Repo) {
	return func(r *repos.Repo) {
		r.Sources = srcs
	}
}

func metadata(md []byte) func(*repos.Repo) {
	return func(r *repos.Repo) {
		r.Metadata = md
	}
}

func externalID(id string) func(*repos.Repo) {
	return func(r *repos.Repo) {
		r.ExternalRepo.ID = id
	}
}

type fakeClock struct {
	epoch time.Time
	step  time.Duration
	steps int
}

func (c *fakeClock) Now() time.Time {
	c.steps++
	return c.Time(c.steps)
}

func (c fakeClock) Time(steps int) time.Time {
	return c.epoch.Add(time.Duration(steps) * c.step).UTC()
}
