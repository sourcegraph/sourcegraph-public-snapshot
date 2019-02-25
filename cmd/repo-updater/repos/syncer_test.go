package repos

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/k0kubun/pp"
	"github.com/kylelemons/godebug/pretty"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/github"
)

func TestSyncer_Sync(t *testing.T) {
	t.Parallel()
	testSyncerSync(&fakeStore{repos: map[string]*Repo{}})
}

func testSyncerSync(s Store) func(*testing.T) {
	foo := Repo{
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
		sourcer Sourcer
		store   Store
		stored  Repos
		ctx     context.Context
		now     func() time.Time
		diff    Diff
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
	}

	return func(t *testing.T) {
		t.Helper()

		{
			clock := fakeClock{epoch: time.Now(), step: time.Second}
			testCases = append(testCases, testCase{
				name:    "had name and got external_id",
				sourcer: sourcer(nil, source("a", nil, foo.Clone())),
				store:   s,
				stored: Repos{foo.With(sources("a"), func(r *Repo) {
					r.ExternalRepo = api.ExternalRepoSpec{}
				})},
				now: clock.Now,
				diff: Diff{Modified: Repos{
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
				store:  s,
				stored: Repos{foo.With(sources("a"))},
				now:    clock.Now,
				diff: Diff{Modified: Repos{
					foo.With(modifiedAt(clock.Time(1)), sources("a", "b")),
				}},
				err: "<nil>",
			})
		}

		{
			clock := fakeClock{epoch: time.Now(), step: time.Second}
			testCases = append(testCases, testCase{
				name: "enabled field is not updateable",
				sourcer: sourcer(nil, source("a", nil, foo.With(func(r *Repo) {
					r.Enabled = !r.Enabled
				}))),
				store:  s,
				stored: Repos{foo.With(sources("a"))},
				now:    clock.Now,
				diff:   Diff{Unmodified: Repos{foo.With(sources("a"))}},
				err:    "<nil>",
			})
		}

		{
			clock := fakeClock{epoch: time.Now(), step: time.Second}
			testCases = append(testCases, testCase{
				name: "deleted repo source",
				sourcer: sourcer(nil,
					source("a", nil, foo.Clone()),
				),
				store:  s,
				stored: Repos{foo.With(sources("a", "b"))},
				now:    clock.Now,
				diff: Diff{Modified: Repos{
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
				store:   s,
				stored:  Repos{foo.With(sources("a", "b"))},
				now:     clock.Now,
				diff: Diff{Deleted: Repos{
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
				store:   s,
				stored: Repos{foo.With(sources("a"), func(r *Repo) {
					r.Name = "old-name"
				})},
				now: clock.Now,
				diff: Diff{Modified: Repos{
					foo.With(sources("a"), modifiedAt(clock.Time(1))),
				}},
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
						metadata(&github.Repository{IsArchived: true})),
				)),
				store:  s,
				stored: Repos{foo.With(sources("a"))},
				now:    clock.Now,
				diff: Diff{Modified: Repos{
					foo.With(modifiedAt(clock.Time(1)),
						sources("a"),
						metadata(&github.Repository{IsArchived: true})),
				}},
				err: "<nil>",
			})
		}

		for _, tc := range testCases {
			tc := tc
			ctx := context.Background()

			t.Run(tc.name, transact(ctx, tc.store, func(t testing.TB, st Store) {
				now := tc.now
				if now == nil {
					now = time.Now
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

				syncer := NewSyncer(0, st, tc.sourcer, nil, now)
				diff, err := syncer.Sync(ctx)

				var want []*Repo
				for _, ds := range []Repos{
					diff.Added,
					diff.Modified,
					diff.Unmodified,
				} {
					for _, d := range ds {
						d.ID = 0 // Exclude auto-generated ID from comparisons
						want = append(want, d)
					}
				}
				for _, d := range diff.Deleted {
					d.ID = 0 // Exclude auto-generated ID from comparisons
				}

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
					if diff := cmp.Diff(have, want); diff != "" {
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

	for _, tc := range []struct {
		name   string
		store  Repos
		source Repos
		diff   Diff
	}{
		{
			name: "empty",
			diff: Diff{},
		},
		{
			name:   "added",
			source: Repos{{ExternalRepo: eid("1")}},
			diff:   Diff{Added: Repos{{ExternalRepo: eid("1")}}},
		},
		{
			name:  "deleted",
			store: Repos{{ExternalRepo: eid("1")}},
			diff:  Diff{Deleted: Repos{{ExternalRepo: eid("1")}}},
		},
		{
			name:   "modified",
			store:  Repos{{ExternalRepo: eid("1"), Description: "foo"}},
			source: Repos{{ExternalRepo: eid("1"), Description: "bar"}},
			diff: Diff{Modified: Repos{
				{ExternalRepo: eid("1"), Description: "bar"},
			}},
		},
		{
			name:   "unmodified",
			store:  Repos{{ExternalRepo: eid("1"), Description: "foo"}},
			source: Repos{{ExternalRepo: eid("1"), Description: "foo"}},
			diff: Diff{Unmodified: Repos{
				{ExternalRepo: eid("1"), Description: "foo"},
			}},
		},
		{
			name: "duplicates in source are merged",
			source: Repos{
				{ExternalRepo: eid("1"), Description: "foo", Sources: map[string]*SourceInfo{
					"a": {ID: "a"},
				}},
				{ExternalRepo: eid("1"), Description: "bar", Sources: map[string]*SourceInfo{
					"b": {ID: "b"},
				}},
			},
			diff: Diff{Added: Repos{
				{ExternalRepo: eid("1"), Description: "bar", Sources: map[string]*SourceInfo{
					"a": {ID: "a"},
					"b": {ID: "b"},
				}},
			}},
		},
		{
			name: "duplicate with a changed name is merged correctly",
			store: Repos{
				{Name: "1", ExternalRepo: eid("1"), Description: "foo"},
			},
			source: Repos{
				{Name: "2", ExternalRepo: eid("1"), Description: "foo"},
			},
			diff: Diff{Modified: Repos{
				{Name: "2", ExternalRepo: eid("1"), Description: "foo"},
			}},
		},
		{
			name: "duplicate with added external id is merged correctly",
			store: Repos{
				{Name: "1", Description: "foo"},
			},
			source: Repos{
				{Name: "1", ExternalRepo: eid("1"), Description: "foo"},
			},
			diff: Diff{Modified: Repos{
				{Name: "1", ExternalRepo: eid("1"), Description: "foo"},
			}},
		},
		{
			name: "no duplicate with added external id and changed name",
			store: Repos{
				{Name: "1", Description: "foo"},
			},
			source: Repos{
				{Name: "2", ExternalRepo: eid("1"), Description: "foo"},
			},
			diff: Diff{
				Deleted: Repos{
					{Name: "1", Description: "foo"},
				},
				Added: Repos{
					{Name: "2", ExternalRepo: eid("1"), Description: "foo"},
				},
			},
		},
		{
			name: "unmodified preserves stored repo",
			store: Repos{
				{ExternalRepo: eid("1"), Description: "foo", UpdatedAt: now},
			},
			source: Repos{
				{ExternalRepo: eid("1"), Description: "foo"},
			},
			diff: Diff{Unmodified: Repos{
				{ExternalRepo: eid("1"), Description: "foo", UpdatedAt: now},
			}},
		},
		{
			// Repo renamed and repo created with old name
			name: "renamed repo",
			store: Repos{
				{Name: "1", ExternalRepo: eid("old"), Description: "foo"},
			},
			source: Repos{
				{Name: "2", ExternalRepo: eid("old"), Description: "foo"},
				{Name: "1", ExternalRepo: eid("new"), Description: "bar"},
			},
			diff: Diff{
				Modified: Repos{
					{Name: "2", ExternalRepo: eid("old"), Description: "foo"},
				},
				Added: Repos{
					{Name: "1", ExternalRepo: eid("new"), Description: "bar"},
				},
			},
		},
		{
			name: "swapped repo",
			store: Repos{
				{Name: "foo", ExternalRepo: eid("1"), Description: "foo"},
				{Name: "bar", ExternalRepo: eid("2"), Description: "bar"},
			},
			source: Repos{
				{Name: "bar", ExternalRepo: eid("1"), Description: "bar"},
				{Name: "foo", ExternalRepo: eid("2"), Description: "foo"},
			},
			diff: Diff{
				Modified: Repos{
					{Name: "bar", ExternalRepo: eid("1"), Description: "bar"},
					{Name: "foo", ExternalRepo: eid("2"), Description: "foo"},
				},
			},
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			diff := NewDiff(tc.source, tc.store)
			diff.Sort()
			tc.diff.Sort()
			if cmp := pretty.Compare(diff, tc.diff); cmp != "" {
				t.Logf("have: %s\nwant: %s\n", pp.Sprint(diff), pp.Sprint(tc.diff))
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
	srcs []Source
}

func sourcer(err error, srcs ...Source) fakeSourcer {
	return fakeSourcer{err: err, srcs: srcs}
}

func (s fakeSourcer) ListSources(context.Context) ([]Source, error) {
	return s.srcs, s.err
}

type fakeSource struct {
	urn   string
	repos []*Repo
	err   error
}

func source(urn string, err error, rs ...*Repo) fakeSource {
	return fakeSource{urn: urn, err: err, repos: rs}
}

func (s fakeSource) ListRepos(context.Context) ([]*Repo, error) {
	repos := make([]*Repo, len(s.repos))
	for i, r := range s.repos {
		repos[i] = r.With(sources(s.urn))
	}
	return repos, s.err
}

type fakeStore struct {
	repos  map[string]*Repo
	get    error
	list   error
	upsert error
}

func (s fakeStore) GetRepoByName(ctx context.Context, name string) (*Repo, error) {
	if s.get != nil {
		return nil, s.get
	}

	r := s.repos[name]
	if r == nil {
		return nil, ErrNoResults
	}

	return r, nil
}

func (s fakeStore) ListRepos(context.Context, ...string) ([]*Repo, error) {
	if s.list != nil {
		return nil, s.list
	}

	set := make(map[*Repo]bool, len(s.repos))
	repos := make([]*Repo, 0, len(s.repos))
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

func (s *fakeStore) UpsertRepos(_ context.Context, upserts ...*Repo) error {
	if s.upsert != nil {
		return s.upsert
	}

	if s.repos == nil {
		s.repos = make(map[string]*Repo, len(upserts))
	}

	for _, upsert := range upserts {
		repo := s.repos[upsert.Name]
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

		s.repos[upsert.Name] = upsert
	}

	return nil
}

//
// Repo functional options
//

func createdAt(ts time.Time) func(*Repo) {
	return func(r *Repo) {
		r.CreatedAt = ts
		r.DeletedAt = time.Time{}
	}
}

func modifiedAt(ts time.Time) func(*Repo) {
	return func(r *Repo) {
		r.UpdatedAt = ts
		r.DeletedAt = time.Time{}
	}
}

func deletedAt(ts time.Time) func(*Repo) {
	return func(r *Repo) {
		r.UpdatedAt = ts
		r.DeletedAt = ts
		r.Sources = map[string]*SourceInfo{}
	}
}

func sources(srcs ...string) func(*Repo) {
	return func(r *Repo) {
		r.Sources = map[string]*SourceInfo{}
		for _, src := range srcs {
			r.Sources[src] = &SourceInfo{ID: src}
		}
	}
}

func metadata(md interface{}) func(*Repo) {
	return func(r *Repo) {
		r.Metadata = md
	}
}

func externalID(id string) func(*Repo) {
	return func(r *Repo) {
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
