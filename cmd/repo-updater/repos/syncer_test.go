package repos_test

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/gitchander/permutation"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/github"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/gitlab"
)

func TestSyncer_Sync(t *testing.T) {
	t.Parallel()

	testSyncerSync(new(repos.FakeStore))(t)

	github := repos.ExternalService{ID: 1, Kind: "github"}
	gitlab := repos.ExternalService{ID: 2, Kind: "gitlab"}

	for _, tc := range []struct {
		name    string
		sourcer repos.Sourcer
		store   repos.Store
		err     string
	}{
		{
			name:    "sourcer error aborts sync",
			sourcer: repos.NewFakeSourcer(errors.New("boom")),
			store:   new(repos.FakeStore),
			err:     "syncer.sync.sourced: boom",
		},
		{
			name: "sources partial errors aborts sync",
			sourcer: repos.NewFakeSourcer(nil,
				repos.NewFakeSource(&github, nil),
				repos.NewFakeSource(&gitlab, errors.New("boom")),
			),
			store: new(repos.FakeStore),
			err:   "syncer.sync.sourced: 1 error occurred:\n\t* boom\n\n",
		},
		{
			name:    "store list error aborts sync",
			sourcer: repos.NewFakeSourcer(nil, repos.NewFakeSource(&github, nil)),
			store:   &repos.FakeStore{ListReposError: errors.New("boom")},
			err:     "syncer.sync.store.list-repos: boom",
		},
		{
			name:    "store upsert error aborts sync",
			sourcer: repos.NewFakeSourcer(nil, repos.NewFakeSource(&github, nil)),
			store:   &repos.FakeStore{UpsertReposError: errors.New("booya")},
			err:     "syncer.sync.store.upsert-repos: booya",
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			clock := repos.NewFakeClock(time.Now(), time.Second)
			now := clock.Now
			ctx := context.Background()

			syncer := repos.NewSyncer(tc.store, tc.sourcer, nil, now)
			_, err := syncer.Sync(ctx, "github")

			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Errorf("have error %q, want %q", have, want)
			}
		})
	}
}

func testSyncerSync(s repos.Store) func(*testing.T) {
	githubService := &repos.ExternalService{
		ID:   1,
		Kind: "GITHUB",
	}

	githubRepo := (&repos.Repo{
		Name:     "github.com/org/foo",
		Metadata: &github.Repository{},
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "foo-external-12345",
			ServiceID:   "https://github.com/",
			ServiceType: "github",
		},
	}).With(
		repos.Opt.RepoSources(githubService.URN()),
	)

	gitlabService := &repos.ExternalService{
		ID:   10,
		Kind: "GITLAB",
	}

	gitlabRepo := (&repos.Repo{
		Name:     "gitlab.com/org/foo",
		Metadata: &gitlab.Project{},
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "12345",
			ServiceID:   "https://gitlab.com/",
			ServiceType: "gitlab",
		},
	}).With(
		repos.Opt.RepoSources(gitlabService.URN()),
	)

	bitbucketServerService := &repos.ExternalService{
		ID:   20,
		Kind: "BITBUCKETSERVER",
	}

	bitbucketServerRepo := (&repos.Repo{
		Name:     "bitbucketserver.mycorp.com/org/foo",
		Metadata: &bitbucketserver.Repo{},
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "23456",
			ServiceID:   "https://bitbucketserver.mycorp.com/",
			ServiceType: "bitbucketServer",
		},
	}).With(
		repos.Opt.RepoSources(bitbucketServerService.URN()),
	)

	otherService := &repos.ExternalService{
		ID:   30,
		Kind: "OTHER",
	}

	otherRepo := (&repos.Repo{
		Name: "git-host.com/org/foo",
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "git-host.com/org/foo",
			ServiceID:   "https://git-host.com/",
			ServiceType: "other",
		},
	}).With(
		repos.Opt.RepoSources(otherService.URN()),
	)

	clock := repos.NewFakeClock(time.Now(), 0)

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

	var testCases []testCase
	for _, tc := range []struct {
		repo *repos.Repo
		svc  *repos.ExternalService
	}{
		{repo: githubRepo, svc: githubService},
		{repo: gitlabRepo, svc: gitlabService},
		{repo: bitbucketServerRepo, svc: bitbucketServerService},
		{repo: otherRepo, svc: otherService},
	} {
		svcdup := tc.svc.With(repos.Opt.ExternalServiceID(tc.svc.ID + 1))
		testCases = append(testCases,
			testCase{
				name: "new repo",
				sourcer: repos.NewFakeSourcer(nil,
					repos.NewFakeSource(tc.svc.Clone(), nil, tc.repo.Clone()),
				),
				store:  s,
				stored: repos.Repos{},
				now:    clock.Now,
				diff: repos.Diff{Added: repos.Repos{tc.repo.With(
					repos.Opt.RepoCreatedAt(clock.Time(1)),
					repos.Opt.RepoSources(tc.svc.Clone().URN()),
				)}},
				err: "<nil>",
			},
			testCase{
				name:    "had name and got external_id",
				sourcer: repos.NewFakeSourcer(nil, repos.NewFakeSource(tc.svc.Clone(), nil, tc.repo.Clone())),
				store:   s,
				stored: repos.Repos{tc.repo.With(func(r *repos.Repo) {
					r.ExternalRepo = api.ExternalRepoSpec{}
				})},
				now: clock.Now,
				diff: repos.Diff{Modified: repos.Repos{
					tc.repo.With(repos.Opt.RepoModifiedAt(clock.Time(1))),
				}},
				err: "<nil>",
			},
			testCase{
				name:    "do not update external_id",
				sourcer: repos.NewFakeSourcer(nil, repos.NewFakeSource(tc.svc.Clone(), nil, tc.repo.Clone())),
				store:   s,
				stored: repos.Repos{tc.repo.With(func(r *repos.Repo) {
					r.ExternalRepo.ID = "old-and-out-of-date"
				})},
				now: clock.Now,
				diff: repos.Diff{
					Added: repos.Repos{tc.repo.With(
						repos.Opt.RepoCreatedAt(clock.Time(1)),
						repos.Opt.RepoModifiedAt(clock.Time(1)),
					)},
					Deleted: repos.Repos{tc.repo.With(func(r *repos.Repo) {
						r.ExternalRepo.ID = "old-and-out-of-date"
						r.Name = fmt.Sprintf("!DELETED!%s!%s", clock.Time(0).UTC().Format("20060102150405"), r.Name)
						r.Sources = map[string]*repos.SourceInfo{}
						r.Enabled = true
						r.DeletedAt = clock.Time(0)
						r.UpdatedAt = clock.Time(0)
					}),
					}},
				err: "<nil>",
			},
			testCase{
				name: "new repo sources",
				sourcer: repos.NewFakeSourcer(nil,
					repos.NewFakeSource(tc.svc.Clone(), nil, tc.repo.Clone()),
					repos.NewFakeSource(svcdup.Clone(), nil, tc.repo.Clone()),
				),
				store:  s,
				stored: repos.Repos{tc.repo.Clone()},
				now:    clock.Now,
				diff: repos.Diff{Modified: repos.Repos{tc.repo.With(
					repos.Opt.RepoModifiedAt(clock.Time(1)),
					repos.Opt.RepoSources(tc.svc.URN(), svcdup.URN()),
				)}},
				err: "<nil>",
			},
			testCase{
				name: "enabled field is not updateable",
				sourcer: repos.NewFakeSourcer(nil, repos.NewFakeSource(tc.svc.Clone(), nil, tc.repo.With(func(r *repos.Repo) {
					r.Enabled = !r.Enabled
				}))),
				store:  s,
				stored: repos.Repos{tc.repo.Clone()},
				now:    clock.Now,
				diff:   repos.Diff{Unmodified: repos.Repos{tc.repo.Clone()}},
				err:    "<nil>",
			},
			testCase{
				name: "enabled field of a undeleted repo is not updateable",
				sourcer: repos.NewFakeSourcer(nil, repos.NewFakeSource(tc.svc.Clone(), nil, tc.repo.With(func(r *repos.Repo) {
					r.Enabled = !r.Enabled
				}))),
				store:  s,
				stored: repos.Repos{tc.repo.With(repos.Opt.RepoDeletedAt(clock.Time(0)))},
				now:    clock.Now,
				diff: repos.Diff{Added: repos.Repos{tc.repo.With(
					repos.Opt.RepoCreatedAt(clock.Time(1)),
				)}},
				err: "<nil>",
			},
			testCase{
				name: "deleted repo source",
				sourcer: repos.NewFakeSourcer(nil,
					repos.NewFakeSource(tc.svc.Clone(), nil, tc.repo.Clone()),
				),
				store: s,
				stored: repos.Repos{tc.repo.With(
					repos.Opt.RepoSources(tc.svc.URN(), svcdup.URN()),
				)},
				now: clock.Now,
				diff: repos.Diff{Modified: repos.Repos{tc.repo.With(
					repos.Opt.RepoModifiedAt(clock.Time(1)),
				)}},
				err: "<nil>",
			},
			testCase{
				name:    "deleted ALL repo sources",
				sourcer: repos.NewFakeSourcer(nil),
				store:   s,
				stored: repos.Repos{tc.repo.With(
					repos.Opt.RepoSources(tc.svc.URN(), svcdup.URN()),
				)},
				now: clock.Now,
				diff: repos.Diff{Deleted: repos.Repos{tc.repo.With(
					repos.Opt.RepoDeletedAt(clock.Time(1)),
					repos.Opt.RepoEnabled(true),
				)}},
				err: "<nil>",
			},
			testCase{
				name:    "renamed repo is detected via external_id",
				sourcer: repos.NewFakeSourcer(nil, repos.NewFakeSource(tc.svc.Clone(), nil, tc.repo.Clone())),
				store:   s,
				stored: repos.Repos{tc.repo.With(func(r *repos.Repo) {
					r.Name = "old-name"
				})},
				now: clock.Now,
				diff: repos.Diff{Modified: repos.Repos{tc.repo.With(
					repos.Opt.RepoModifiedAt(clock.Time(1))),
				}},
				err: "<nil>",
			},
			testCase{
				name:    "renamed repo which was deleted is detected and added",
				sourcer: repos.NewFakeSourcer(nil, repos.NewFakeSource(tc.svc.Clone(), nil, tc.repo.Clone())),
				store:   s,
				stored: repos.Repos{tc.repo.With(func(r *repos.Repo) {
					r.Sources = map[string]*repos.SourceInfo{}
					r.Name = "old-name"
					r.DeletedAt = clock.Time(0)
				})},
				now: clock.Now,
				diff: repos.Diff{Added: repos.Repos{tc.repo.With(
					repos.Opt.RepoCreatedAt(clock.Time(1))),
				}},
				err: "<nil>",
			},
			testCase{
				// We have two stored repos (name1, id1) and (name2, id2)
				// Then we source repos (name1, id2) which got renamed and
				// not (name1, id1) which got deleted.
				name: "renamed to repo that is deleted",
				sourcer: repos.NewFakeSourcer(nil,
					repos.NewFakeSource(tc.svc.Clone(), nil,
						tc.repo.With(func(r *repos.Repo) { r.ExternalRepo.ID = "another-id" }),
					),
				),
				store: s,
				stored: repos.Repos{
					tc.repo.With(func(r *repos.Repo) {
						r.Name = "another-repo"
						r.ExternalRepo.ID = "another-id"
					}),
					tc.repo.Clone(),
				},
				now: clock.Now,
				diff: repos.Diff{
					Deleted: repos.Repos{
						tc.repo.With(func(r *repos.Repo) {
							r.Name = fmt.Sprintf("!DELETED!%s!%s", clock.Time(0).UTC().Format("20060102150405"), r.Name)
							r.Sources = map[string]*repos.SourceInfo{}
							r.Enabled = true
							r.DeletedAt = clock.Time(0)
							r.UpdatedAt = clock.Time(0)
						}),
					},
					Modified: repos.Repos{
						tc.repo.With(
							repos.Opt.RepoModifiedAt(clock.Time(1)),
							func(r *repos.Repo) { r.ExternalRepo.ID = "another-id" },
						),
					},
				},
				err: "<nil>",
			},
			func() testCase {
				var update interface{}
				switch strings.ToLower(tc.repo.ExternalRepo.ServiceType) {
				case "github":
					update = &github.Repository{IsArchived: true}
				case "gitlab":
					update = &gitlab.Project{Archived: true}
				case "bitbucketserver":
					update = &bitbucketserver.Repo{Public: true}
				case "other":
					return testCase{}
				default:
					panic("test must be extended with new external service kind")
				}

				return testCase{
					name: "metadata update",
					sourcer: repos.NewFakeSourcer(nil, repos.NewFakeSource(tc.svc.Clone(), nil,
						tc.repo.With(repos.Opt.RepoModifiedAt(clock.Time(1)),
							repos.Opt.RepoMetadata(update)),
					)),
					store:  s,
					stored: repos.Repos{tc.repo.Clone()},
					now:    clock.Now,
					diff: repos.Diff{Modified: repos.Repos{tc.repo.With(
						repos.Opt.RepoModifiedAt(clock.Time(1)),
						repos.Opt.RepoMetadata(update),
					)}},
					err: "<nil>",
				}
			}(),
		)

	}

	return func(t *testing.T) {
		t.Helper()

		for _, tc := range testCases {
			if tc.name == "" {
				continue
			}

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
				diff, err := syncer.Sync(ctx)

				var want repos.Repos
				want.Concat(diff.Added, diff.Modified, diff.Unmodified, diff.Deleted)
				want.Apply(repos.Opt.RepoID(0)) // Exclude auto-generated ID from comparisons
				sort.Sort(want)

				if have, want := fmt.Sprint(err), tc.err; have != want {
					t.Errorf("have error %q, want %q", have, want)
				}
				if err != nil {
					return
				}

				if diff := cmp.Diff(diff, tc.diff); diff != "" {
					t.Fatalf("unexpected diff:\n%s", diff)
				}

				if st != nil {
					var have repos.Repos
					have, _ = st.ListRepos(ctx, repos.StoreListReposArgs{Deleted: true})
					have.Apply(repos.Opt.RepoID(0))
					sort.Sort(have)

					if diff := cmp.Diff(have, want); diff != "" {
						viewAll(t, "have", have...)
						viewAll(t, "want", want...)
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
	now, err := time.Parse(time.RFC3339, "2019-05-01T09:57:05Z")
	if err != nil {
		t.Fatal(err)
	}

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
			name: "repo renamed to an already deleted repo",
			store: repos.Repos{
				{Name: "a", ExternalRepo: eid("a"), DeletedAt: now},
			},
			source: repos.Repos{
				{Name: "a", ExternalRepo: eid("b")},
			},
			diff: repos.Diff{
				Added: repos.Repos{
					{Name: "a", ExternalRepo: eid("b")},
				},
				Modified: repos.Repos{
					{Name: "!DELETED!20190501095705!a", ExternalRepo: eid("a"), DeletedAt: now},
				},
			},
		},
		{
			name: "repo renamed to a repo that gets deleted",
			store: repos.Repos{
				{Name: "a", ExternalRepo: eid("a")},
				{Name: "b", ExternalRepo: eid("b")},
			},
			source: repos.Repos{
				{Name: "a", ExternalRepo: eid("b")},
			},
			diff: repos.Diff{
				Deleted:  repos.Repos{{Name: "!DELETED!20190501095705!a", ExternalRepo: eid("a")}},
				Modified: repos.Repos{{Name: "a", ExternalRepo: eid("b")}},
			},
		},
		{
			name: "swapped repo",
			store: repos.Repos{
				{Name: "foo", ExternalRepo: eid("1"), Description: "foo"},
				{Name: "bar", ExternalRepo: eid("2"), Description: "bar"},
			},
			source: repos.Repos{
				{Name: "bar", ExternalRepo: eid("1"), Description: "foo"},
				{Name: "foo", ExternalRepo: eid("2"), Description: "bar"},
			},
			diff: repos.Diff{
				Modified: repos.Repos{
					{Name: "bar", ExternalRepo: eid("1"), Description: "foo"},
					{Name: "foo", ExternalRepo: eid("2"), Description: "bar"},
				},
			},
		},
		{
			name: "unset external repo is not associated with set external repo",
			store: repos.Repos{
				{Name: "foo", ExternalRepo: eid("1"), Description: "foo"},
				{Name: "bar", Description: "bar"},
			},
			source: repos.Repos{
				{Name: "bar", ExternalRepo: eid("1"), Description: "foo"},
			},
			diff: repos.Diff{
				Modified: repos.Repos{
					{Name: "bar", ExternalRepo: eid("1"), Description: "foo"},
				},
				Deleted: repos.Repos{
					{Name: "!DELETED!20190501095705!bar", Description: "bar"},
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

			diff := repos.NewDiff(tc.source, tc.store, now)
			diff.Sort()
			tc.diff.Sort()
			if cDiff := cmp.Diff(diff, tc.diff); cDiff != "" {
				//t.Logf("have: %s\nwant: %s\n", pp.Sprint(diff), pp.Sprint(tc.diff))
				t.Fatalf("unexpected diff:\n%s", cDiff)
			}
		})
	}
}

func view(r *repos.Repo) string {
	if r == nil {
		return "<nil>"
	}
	return fmt.Sprintf("(%s, %s, %v)", r.Name, r.ExternalRepo.ID, !r.DeletedAt.IsZero())
}

func viewAll(t testing.TB, prefix string, rs ...*repos.Repo) {
	t.Helper()
	for _, r := range rs {
		t.Logf("%s: %s\n", prefix, view(r))
	}
}
