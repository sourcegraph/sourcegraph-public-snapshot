package repos_test

import (
	"context"
	"fmt"
	"reflect"
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
		Name: "bar",
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "bar12345",
			ServiceID:   "https://github.com/",
			ServiceType: "github",
		},
	}

	clock := fakeClock{epoch: time.Now(), step: time.Second}

	for _, tc := range []struct {
		name    string
		sourcer repos.Sourcer
		store   repos.Store
		ctx     context.Context
		now     func() time.Time
		diff    repos.Diff
		stored  repos.Store
		err     string
	}{
		{
			name:    "sourcer error aborts sync",
			sourcer: sourcer(errors.New("boom")),
			err:     "boom",
		},
		{
			name: "sources partial errors aborts sync",
			sourcer: sourcer(nil,
				source(nil, foo.Clone()),
				source(errors.New("boom")),
			),
			err: "1 error occurred:\n\t* boom\n\n",
		},
		{
			name: "sources partial errors aborts sync",
			sourcer: sourcer(nil,
				source(nil, foo.Clone()),
				source(errors.New("boom")),
			),
			err: "1 error occurred:\n\t* boom\n\n",
		},
		{
			name: "sources partial errors aborts sync",
			sourcer: sourcer(nil,
				source(nil, foo.Clone()),
				source(errors.New("boom")),
			),
			err: "1 error occurred:\n\t* boom\n\n",
		},
		{
			name:    "store list error aborts sync",
			sourcer: sourcer(nil, source(nil, foo.Clone())),
			store:   &fakeStore{list: errors.New("boom")},
			err:     "boom",
		},
		{
			name:    "store upsert error aborts sync",
			sourcer: sourcer(nil, source(nil, foo.Clone())),
			store:   &fakeStore{upsert: errors.New("booya")},
			err:     "booya",
		},
		{
			// here we test that if a repo previously had the external id set and one
			// of its sources gets rate limited, we preserve the external id
			name: "had name and external id, got name only from rate-limited source",
			sourcer: sourcer(nil, source(nil, foo.With(func(r *repos.Repo) {
				r.ExternalRepo = api.ExternalRepoSpec{}
			}))),
			store:  store(foo.Clone()),
			diff:   repos.Diff{Unmodified: []repos.Diffable{foo.Clone()}},
			stored: store(foo.Clone()),
			err:    "<nil>",
		},
		{
			name:    "had name and got external_id",
			sourcer: sourcer(nil, source(nil, foo.Clone())),
			store: store(foo.With(func(r *repos.Repo) {
				r.ExternalRepo = api.ExternalRepoSpec{}
			})),
			now: clock.Now,
			diff: repos.Diff{Modified: []repos.Diffable{
				foo.With(modifiedAt(clock.Time(1))),
			}},
			stored: store(foo.With(modifiedAt(clock.Time(1)))),
			err:    "<nil>",
		},
	} {
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

			if have, want := diff, tc.diff; !reflect.DeepEqual(have, want) {
				t.Fatalf("unexpected diff:\n%s", pretty.Compare(have, want))
			}

			if tc.stored != nil {
				have, _ := tc.store.ListRepos(tc.ctx)
				want, _ := tc.stored.ListRepos(tc.ctx)
				if !reflect.DeepEqual(have, want) {
					t.Logf("unexpected stored repos:\n%s", pretty.Compare(have, want))
				}
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
	repos []*repos.Repo
	err   error
}

func source(err error, rs ...*repos.Repo) fakeSource {
	return fakeSource{err: err, repos: rs}
}

func (s fakeSource) ListRepos(context.Context) ([]*repos.Repo, error) {
	return s.repos, s.err
}

type fakeStore struct {
	id     uint32
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

func (s fakeStore) ListRepos(_ context.Context) ([]*repos.Repo, error) {
	if s.list != nil {
		return nil, s.list
	}

	set := make(map[*repos.Repo]struct{}, len(s.repos))
	for _, r := range s.repos {
		if _, ok := set[r]; !ok {
			set[r] = struct{}{}
		}
	}

	repos := make([]*repos.Repo, 0, len(set))
	for r := range set {
		repos = append(repos, r)
	}

	sort.Slice(repos, func(i, j int) bool {
		return repos[i].ID < repos[j].ID
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
		for _, id := range upsert.IDs() {
			repo, ok := s.repos[id]
			if ok {
				repo.Description = upsert.Description
				repo.Language = upsert.Language
				repo.UpdatedAt = upsert.UpdatedAt
				repo.DeletedAt = upsert.DeletedAt
				repo.Archived = upsert.Archived
				repo.Fork = upsert.Fork
				repo.Sources = upsert.Sources
				repo.Metadata = upsert.Metadata
				repo.ExternalRepo = upsert.ExternalRepo
				break
			}

			if upsert.ID == 0 {
				s.id++
				upsert.ID = s.id
			}

			s.repos[id] = upsert
		}
	}

	return nil
}

func modifiedAt(ts time.Time) func(*repos.Repo) {
	return func(r *repos.Repo) {
		r.UpdatedAt = ts
		r.DeletedAt = time.Time{}
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
