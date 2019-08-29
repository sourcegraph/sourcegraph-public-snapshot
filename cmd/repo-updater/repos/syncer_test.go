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
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/awscodecommit"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/bitbucketcloud"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/github"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/gitolite"
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
			err:     "syncer.sync.sourced: 1 error occurred:\n\t* boom\n\n",
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
			_, err := syncer.Sync(ctx)

			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Errorf("have error %q, want %q", have, want)
			}

			if have, want := fmt.Sprint(syncer.LastSyncError()), tc.err; have != want {
				t.Errorf("have LastSyncError %q, want %q", have, want)
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
		Enabled:  true,
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
		Enabled:  true,
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
		Enabled:  true,
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "23456",
			ServiceID:   "https://bitbucketserver.mycorp.com/",
			ServiceType: "bitbucketServer",
		},
	}).With(
		repos.Opt.RepoSources(bitbucketServerService.URN()),
	)

	awsCodeCommitService := &repos.ExternalService{
		ID:   30,
		Kind: "AWSCODECOMMIT",
	}

	awsCodeCommitRepo := (&repos.Repo{
		Name:     "git-codecommit.us-west-1.amazonaws.com/stripe-go",
		Metadata: &awscodecommit.Repository{},
		Enabled:  true,
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "f001337a-3450-46fd-b7d2-650c0EXAMPLE",
			ServiceID:   "arn:aws:codecommit:us-west-1:999999999999:",
			ServiceType: "awscodecommit",
		},
	}).With(
		repos.Opt.RepoSources(awsCodeCommitService.URN()),
	)

	otherService := &repos.ExternalService{
		ID:   40,
		Kind: "OTHER",
	}

	otherRepo := (&repos.Repo{
		Name:    "git-host.com/org/foo",
		Enabled: true,
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "git-host.com/org/foo",
			ServiceID:   "https://git-host.com/",
			ServiceType: "other",
		},
	}).With(
		repos.Opt.RepoSources(otherService.URN()),
	)

	gitoliteService := &repos.ExternalService{
		ID:   50,
		Kind: "GITOLITE",
	}

	gitoliteRepo := (&repos.Repo{
		Name:     "gitolite.mycorp.com/foo",
		Metadata: &gitolite.Repo{},
		Enabled:  true,
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "foo",
			ServiceID:   "git@gitolite.mycorp.com",
			ServiceType: "gitolite",
		},
	}).With(
		repos.Opt.RepoSources(gitoliteService.URN()),
	)

	bitbucketCloudService := &repos.ExternalService{
		ID:   60,
		Kind: "BITBUCKETCLOUD",
	}

	bitbucketCloudRepo := (&repos.Repo{
		Name:     "bitbucket.org/team/foo",
		Metadata: &bitbucketcloud.Repo{},
		Enabled:  true,
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "{e164a64c-bd73-4a40-b447-d71b43f328a8}",
			ServiceID:   "https://bitbucket.org/",
			ServiceType: "bitbucketCloud",
		},
	}).With(
		repos.Opt.RepoSources(bitbucketCloudService.URN()),
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
		{repo: awsCodeCommitRepo, svc: awsCodeCommitService},
		{repo: otherRepo, svc: otherService},
		{repo: gitoliteRepo, svc: gitoliteService},
		{repo: bitbucketCloudRepo, svc: bitbucketCloudService},
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
					r.ExternalRepo.ID = ""
				})},
				now: clock.Now,
				diff: repos.Diff{Modified: repos.Repos{
					tc.repo.With(repos.Opt.RepoModifiedAt(clock.Time(1))),
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
				diff: repos.Diff{Modified: repos.Repos{
					tc.repo.With(
						repos.Opt.RepoModifiedAt(clock.Time(1))),
				}},
				err: "<nil>",
			},
			testCase{
				name: "repo got renamed to another repo that gets deleted",
				sourcer: repos.NewFakeSourcer(nil,
					repos.NewFakeSource(tc.svc.Clone(), nil,
						tc.repo.With(func(r *repos.Repo) { r.ExternalRepo.ID = "another-id" }),
					),
				),
				store: s,
				stored: repos.Repos{
					tc.repo.Clone(),
					tc.repo.With(func(r *repos.Repo) {
						r.Name = "another-repo"
						r.ExternalRepo.ID = "another-id"
					}),
				},
				now: clock.Now,
				diff: repos.Diff{
					Deleted: repos.Repos{
						tc.repo.With(func(r *repos.Repo) {
							r.Enabled = true
							r.Sources = map[string]*repos.SourceInfo{}
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
			testCase{
				name: "repo inserted with same name as another repo that gets deleted",
				sourcer: repos.NewFakeSourcer(nil,
					repos.NewFakeSource(tc.svc.Clone(), nil,
						tc.repo,
					),
				),
				store: s,
				stored: repos.Repos{
					tc.repo.With(repos.Opt.RepoExternalID("another-id")),
				},
				now: clock.Now,
				diff: repos.Diff{
					Added: repos.Repos{
						tc.repo.With(
							repos.Opt.RepoCreatedAt(clock.Time(1)),
							repos.Opt.RepoModifiedAt(clock.Time(1)),
						),
					},
					Deleted: repos.Repos{
						tc.repo.With(func(r *repos.Repo) {
							r.ExternalRepo.ID = "another-id"
							r.Sources = map[string]*repos.SourceInfo{}
							r.DeletedAt = clock.Time(0)
							r.UpdatedAt = clock.Time(0)
						}),
					},
				},
				err: "<nil>",
			},
			testCase{
				name: "repo inserted with same name as repo without id",
				sourcer: repos.NewFakeSourcer(nil,
					repos.NewFakeSource(tc.svc.Clone(), nil,
						tc.repo,
					),
				),
				store: s,
				stored: repos.Repos{
					tc.repo.With(repos.Opt.RepoName("old-name")), // same external id as sourced
					tc.repo.With(repos.Opt.RepoExternalID("")),   // same name as sourced
				}.With(repos.Opt.RepoCreatedAt(clock.Time(1))),
				now: clock.Now,
				diff: repos.Diff{
					Modified: repos.Repos{
						tc.repo.With(
							repos.Opt.RepoCreatedAt(clock.Time(1)),
							repos.Opt.RepoModifiedAt(clock.Time(1)),
						),
					},
					Deleted: repos.Repos{
						tc.repo.With(func(r *repos.Repo) {
							r.ExternalRepo.ID = ""
							r.Sources = map[string]*repos.SourceInfo{}
							r.DeletedAt = clock.Time(0)
							r.UpdatedAt = clock.Time(0)
							r.CreatedAt = clock.Time(0)
						}),
					},
				},
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
				diff: repos.Diff{Added: repos.Repos{
					tc.repo.With(
						repos.Opt.RepoCreatedAt(clock.Time(1))),
				}},
				err: "<nil>",
			},
			testCase{
				name: "repos have their names swapped",
				sourcer: repos.NewFakeSourcer(nil, repos.NewFakeSource(tc.svc.Clone(), nil,
					tc.repo.With(func(r *repos.Repo) {
						r.Name = "foo"
						r.ExternalRepo.ID = "1"
					}),
					tc.repo.With(func(r *repos.Repo) {
						r.Name = "bar"
						r.ExternalRepo.ID = "2"
					}),
				)),
				now:   clock.Now,
				store: s,
				stored: repos.Repos{
					tc.repo.With(func(r *repos.Repo) {
						r.Name = "bar"
						r.ExternalRepo.ID = "1"
					}),
					tc.repo.With(func(r *repos.Repo) {
						r.Name = "foo"
						r.ExternalRepo.ID = "2"
					}),
				},
				diff: repos.Diff{
					Modified: repos.Repos{
						tc.repo.With(func(r *repos.Repo) {
							r.Name = "foo"
							r.ExternalRepo.ID = "1"
							r.UpdatedAt = clock.Time(0)
						}),
						tc.repo.With(func(r *repos.Repo) {
							r.Name = "bar"
							r.ExternalRepo.ID = "2"
							r.UpdatedAt = clock.Time(0)
						}),
					},
				},
				err: "<nil>",
			},
			testCase{
				name: "case insensitive name",
				sourcer: repos.NewFakeSourcer(nil, repos.NewFakeSource(tc.svc.Clone(), nil,
					tc.repo.Clone(),
					tc.repo.With(repos.Opt.RepoName(strings.ToUpper(tc.repo.Name))),
				)),
				store:  s,
				stored: repos.Repos{tc.repo.With(repos.Opt.RepoName(strings.ToUpper(tc.repo.Name)))},
				now:    clock.Now,
				diff:   repos.Diff{Modified: repos.Repos{tc.repo.With(repos.Opt.RepoModifiedAt(clock.Time(0)))}},
				err:    "<nil>",
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
				case "bitbucketcloud":
					update = &bitbucketcloud.Repo{IsPrivate: true}
				case "awscodecommit":
					update = &awscodecommit.Repository{Description: "new description"}
				case "other", "gitolite":
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
					if err := st.UpsertRepos(ctx, tc.stored.Clone()...); err != nil {
						t.Fatalf("failed to prepare store: %v", err)
					}
				}

				syncer := repos.NewSyncer(st, tc.sourcer, nil, now)
				diff, err := syncer.Sync(ctx)

				if have, want := fmt.Sprint(err), tc.err; have != want {
					t.Errorf("have error %q, want %q", have, want)
				}

				if err != nil {
					return
				}

				for _, d := range []struct {
					name       string
					have, want repos.Repos
				}{
					{"added", diff.Added, tc.diff.Added},
					{"deleted", diff.Deleted, tc.diff.Deleted},
					{"modified", diff.Modified, tc.diff.Modified},
					{"unmodified", diff.Unmodified, tc.diff.Unmodified},
				} {
					t.Logf("diff.%s", d.name)
					repos.Assert.ReposEqual(d.want...)(t, d.have)
				}

				if st != nil {
					var want repos.Repos
					want.Concat(diff.Added, diff.Modified, diff.Unmodified)
					sort.Sort(want)

					have, _ := st.ListRepos(ctx, repos.StoreListReposArgs{})
					repos.Assert.ReposEqual(want...)(t, have)
				}
			}))
		}
	}
}

func TestSync_SyncSubset(t *testing.T) {
	t.Parallel()

	testSyncSubset(new(repos.FakeStore))(t)
}

func testSyncSubset(s repos.Store) func(*testing.T) {
	clock := repos.NewFakeClock(time.Now(), time.Second)

	repo := &repos.Repo{
		ID:          0, // explicitly make default value for sourced repo
		Name:        "github.com/foo/bar",
		Description: "The description",
		Language:    "barlang",
		Enabled:     true,
		Archived:    false,
		Fork:        false,
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAA==",
			ServiceType: "github",
			ServiceID:   "https://github.com/",
		},
		Sources: map[string]*repos.SourceInfo{
			"extsvc:123": {
				ID:       "extsvc:123",
				CloneURL: "git@github.com:foo/bar.git",
			},
		},
		Metadata: &github.Repository{
			ID:            "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAA==",
			URL:           "github.com/foo/bar",
			DatabaseID:    1234,
			Description:   "The description",
			NameWithOwner: "foo/bar",
		},
	}

	testCases := []struct {
		name    string
		sourced repos.Repos
		stored  repos.Repos
		assert  repos.ReposAssertion
	}{{
		name:   "no sourced",
		stored: repos.Repos{repo.With(repos.Opt.RepoCreatedAt(clock.Time(2)))},
		assert: repos.Assert.ReposEqual(repo.With(repos.Opt.RepoCreatedAt(clock.Time(2)))),
	}, {
		name:    "insert",
		sourced: repos.Repos{repo},
		assert:  repos.Assert.ReposEqual(repo.With(repos.Opt.RepoCreatedAt(clock.Time(2)))),
	}, {
		name:    "update",
		sourced: repos.Repos{repo},
		stored:  repos.Repos{repo.With(repos.Opt.RepoCreatedAt(clock.Time(2)))},
		assert: repos.Assert.ReposEqual(repo.With(
			repos.Opt.RepoModifiedAt(clock.Time(2)),
			repos.Opt.RepoCreatedAt(clock.Time(2)))),
	}, {
		name:    "update name",
		sourced: repos.Repos{repo},
		stored: repos.Repos{repo.With(
			repos.Opt.RepoName("old/name"),
			repos.Opt.RepoCreatedAt(clock.Time(2)))},
		assert: repos.Assert.ReposEqual(repo.With(
			repos.Opt.RepoModifiedAt(clock.Time(2)),
			repos.Opt.RepoCreatedAt(clock.Time(2)))),
	}, {
		name:    "delete conflicting name",
		sourced: repos.Repos{repo},
		stored: repos.Repos{repo.With(
			repos.Opt.RepoExternalID("old id"),
			repos.Opt.RepoCreatedAt(clock.Time(2)))},
		assert: repos.Assert.ReposEqual(repo.With(
			repos.Opt.RepoCreatedAt(clock.Time(2)))),
	}, {
		name:    "rename and delete conflicting name",
		sourced: repos.Repos{repo},
		stored: repos.Repos{
			repo.With(
				repos.Opt.RepoExternalID("old id"),
				repos.Opt.RepoCreatedAt(clock.Time(2))),
			repo.With(
				repos.Opt.RepoName("old name"),
				repos.Opt.RepoCreatedAt(clock.Time(2))),
		},
		assert: repos.Assert.ReposEqual(repo.With(
			repos.Opt.RepoCreatedAt(clock.Time(2)))),
	}}

	return func(t *testing.T) {
		t.Helper()

		for _, tc := range testCases {
			if tc.name == "" {
				continue
			}

			tc := tc
			ctx := context.Background()

			t.Run(tc.name, transact(ctx, s, func(t testing.TB, st repos.Store) {
				defer func() {
					if err := recover(); err != nil {
						t.Fatalf("%q panicked: %v", tc.name, err)
					}
				}()

				if len(tc.stored) > 0 {
					if err := st.UpsertRepos(ctx, tc.stored.Clone()...); err != nil {
						t.Fatalf("failed to prepare store: %v", err)
					}
				}

				clock := clock
				syncer := repos.NewSyncer(st, nil, nil, clock.Now)
				_, err := syncer.SyncSubset(ctx, tc.sourced.Clone()...)
				if err != nil {
					t.Fatal(err)
				}

				have, err := st.ListRepos(ctx, repos.StoreListReposArgs{})
				if err != nil {
					t.Fatal(err)
				}

				tc.assert(t, have)
			}))
		}
	}
}

func TestDiff(t *testing.T) {
	t.Parallel()

	eid := func(id string) api.ExternalRepoSpec {
		return api.ExternalRepoSpec{
			ID:          id,
			ServiceType: "fake",
			ServiceID:   "https://fake.com",
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
			name: "modified",
			store: repos.Repos{
				{ExternalRepo: eid("1"), Name: "foo", Description: "foo"},
				{ExternalRepo: eid("2"), Name: "bar"},
			},
			source: repos.Repos{
				{ExternalRepo: eid("1"), Name: "foo", Description: "bar"},
				{ExternalRepo: eid("2"), Name: "bar", URI: "2"},
			},
			diff: repos.Diff{Modified: repos.Repos{
				{ExternalRepo: eid("1"), Name: "foo", Description: "bar"},
				{ExternalRepo: eid("2"), Name: "bar", URI: "2"},
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
			name:  "repo renamed to an already deleted repo",
			store: repos.Repos{},
			source: repos.Repos{
				{Name: "a", ExternalRepo: eid("b")},
			},
			diff: repos.Diff{
				Added: repos.Repos{
					{Name: "a", ExternalRepo: eid("b")},
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
				Deleted:  repos.Repos{{Name: "a", ExternalRepo: eid("a")}},
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
		{
			name: "deterministic merging of source",
			source: repos.Repos{
				{Name: "foo", ExternalRepo: eid("1"), Description: "desc1", Sources: map[string]*repos.SourceInfo{"a": nil}},
				{Name: "foo", ExternalRepo: eid("1"), Description: "desc2", Sources: map[string]*repos.SourceInfo{"b": nil}},
			},
			diff: repos.Diff{
				Added: repos.Repos{
					{Name: "foo", ExternalRepo: eid("1"), Description: "desc2", Sources: map[string]*repos.SourceInfo{"a": nil, "b": nil}},
				},
			},
		},
		{
			name: "conflict on case insensitive name",
			source: repos.Repos{
				{Name: "foo", ExternalRepo: eid("1")},
				{Name: "Foo", ExternalRepo: eid("2")},
			},
			diff: repos.Diff{
				Added: repos.Repos{
					{Name: "Foo", ExternalRepo: eid("2")},
				},
			},
		},
		{
			name: "conflict on case insensitive name no external",
			store: repos.Repos{
				{Name: "fOO"},
			},
			source: repos.Repos{
				{Name: "fOO", ExternalRepo: eid("fOO")},
				{Name: "Foo", ExternalRepo: eid("Foo")},
			},
			diff: repos.Diff{
				Modified: repos.Repos{
					{Name: "Foo", ExternalRepo: eid("Foo")},
				},
			},
		},
		{
			name: "conflict on case insensitive name exists 1",
			store: repos.Repos{
				{Name: "foo", ExternalRepo: eid("1")},
			},
			source: repos.Repos{
				{Name: "foo", ExternalRepo: eid("1")},
				{Name: "Foo", ExternalRepo: eid("2")},
			},
			diff: repos.Diff{
				Added: repos.Repos{
					{Name: "Foo", ExternalRepo: eid("2")},
				},
				Deleted: repos.Repos{
					{Name: "foo", ExternalRepo: eid("1")},
				},
			},
		},
		{
			name: "conflict on case insensitive name exists 2",
			store: repos.Repos{
				{Name: "Foo", ExternalRepo: eid("2")},
			},
			source: repos.Repos{
				{Name: "foo", ExternalRepo: eid("1")},
				{Name: "Foo", ExternalRepo: eid("2")},
			},
			diff: repos.Diff{
				Unmodified: repos.Repos{
					{Name: "Foo", ExternalRepo: eid("2")},
				},
			},
		},
		{
			name: "associate by name",
			store: repos.Repos{
				{Name: "foo"},
				{Name: "baz"},
			},
			source: repos.Repos{
				{Name: "foo", ExternalRepo: eid("1")},
				{Name: "bar", ExternalRepo: eid("2")},
			},
			diff: repos.Diff{
				Added: repos.Repos{
					{Name: "bar", ExternalRepo: eid("2")},
				},
				Modified: repos.Repos{
					{Name: "foo", ExternalRepo: eid("1")},
				},
				Deleted: repos.Repos{
					{Name: "baz"},
				},
			},
		},
		{
			name: "associate by name conflict",
			store: repos.Repos{
				{Name: "foo"},
				{Name: "bar", ExternalRepo: eid("1")},
			},
			source: repos.Repos{
				{Name: "foo", ExternalRepo: eid("1")},
				{Name: "bar", ExternalRepo: eid("2")},
			},
			diff: repos.Diff{
				Added: repos.Repos{
					{Name: "bar", ExternalRepo: eid("2")},
				},
				Modified: repos.Repos{
					{Name: "foo", ExternalRepo: eid("1")},
				},
				Deleted: repos.Repos{
					{Name: "foo"},
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
				// t.Logf("have: %s\nwant: %s\n", pp.Sprint(diff), pp.Sprint(tc.diff))
				t.Fatalf("unexpected diff:\n%s", cDiff)
			}
		})
	}
}
