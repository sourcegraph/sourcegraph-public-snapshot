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
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/awscodecommit"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketcloud"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitolite"
)

func TestSyncer_Sync(t *testing.T) {
	t.Parallel()

	testSyncerSync(t, new(repos.FakeStore))(t)

	github := repos.ExternalService{ID: 1, Kind: extsvc.KindGitHub}
	gitlab := repos.ExternalService{ID: 2, Kind: extsvc.KindGitLab}

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

			syncer := &repos.Syncer{
				Store:   tc.store,
				Sourcer: tc.sourcer,
				Now:     now,
			}
			err := syncer.Sync(ctx)

			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Errorf("have error %q, want %q", have, want)
			}

			if have, want := fmt.Sprint(syncer.LastSyncError()), tc.err; have != want {
				t.Errorf("have LastSyncError %q, want %q", have, want)
			}
		})
	}
}

func testSyncerSync(t *testing.T, s repos.Store) func(*testing.T) {
	servicesPerKind := createExternalServices(t, s)

	githubService := servicesPerKind[extsvc.KindGitHub]

	githubRepo := (&repos.Repo{
		Name:     "github.com/org/foo",
		Metadata: &github.Repository{},
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "foo-external-12345",
			ServiceID:   "https://github.com/",
			ServiceType: extsvc.TypeGitHub,
		},
	}).With(
		repos.Opt.RepoSources(githubService.URN()),
	)

	gitlabService := servicesPerKind[extsvc.KindGitLab]

	gitlabRepo := (&repos.Repo{
		Name:     "gitlab.com/org/foo",
		Metadata: &gitlab.Project{},
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "12345",
			ServiceID:   "https://gitlab.com/",
			ServiceType: extsvc.TypeGitLab,
		},
	}).With(
		repos.Opt.RepoSources(gitlabService.URN()),
	)

	bitbucketServerService := servicesPerKind[extsvc.KindBitbucketServer]

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

	awsCodeCommitService := servicesPerKind[extsvc.KindAWSCodeCommit]

	awsCodeCommitRepo := (&repos.Repo{
		Name:     "git-codecommit.us-west-1.amazonaws.com/stripe-go",
		Metadata: &awscodecommit.Repository{},
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "f001337a-3450-46fd-b7d2-650c0EXAMPLE",
			ServiceID:   "arn:aws:codecommit:us-west-1:999999999999:",
			ServiceType: extsvc.TypeAWSCodeCommit,
		},
	}).With(
		repos.Opt.RepoSources(awsCodeCommitService.URN()),
	)

	otherService := servicesPerKind[extsvc.KindOther]

	otherRepo := (&repos.Repo{
		Name: "git-host.com/org/foo",
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "git-host.com/org/foo",
			ServiceID:   "https://git-host.com/",
			ServiceType: extsvc.TypeOther,
		},
	}).With(
		repos.Opt.RepoSources(otherService.URN()),
	)

	gitoliteService := servicesPerKind[extsvc.KindGitolite]

	gitoliteRepo := (&repos.Repo{
		Name:     "gitolite.mycorp.com/foo",
		Metadata: &gitolite.Repo{},
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "foo",
			ServiceID:   "git@gitolite.mycorp.com",
			ServiceType: extsvc.TypeGitolite,
		},
	}).With(
		repos.Opt.RepoSources(gitoliteService.URN()),
	)

	bitbucketCloudService := servicesPerKind[extsvc.KindBitbucketCloud]

	bitbucketCloudRepo := (&repos.Repo{
		Name:     "bitbucket.org/team/foo",
		Metadata: &bitbucketcloud.Repo{},
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "{e164a64c-bd73-4a40-b447-d71b43f328a8}",
			ServiceID:   "https://bitbucket.org/",
			ServiceType: extsvc.TypeBitbucketCloud,
		},
	}).With(
		repos.Opt.RepoSources(bitbucketCloudService.URN()),
	)

	clock := repos.NewFakeClock(time.Now(), 0)

	svcdup := repos.ExternalService{
		Kind:        extsvc.KindGitHub,
		DisplayName: "Github2 - Test",
		Config:      `{"url": "https://github.com"}`,
		CreatedAt:   clock.Now(),
		UpdatedAt:   clock.Now(),
	}

	// create a few external services
	if err := s.UpsertExternalServices(context.Background(), &svcdup); err != nil {
		t.Fatalf("failed to insert external services: %v", err)
	}

	var services []repos.ExternalService
	for _, svc := range servicesPerKind {
		services = append(services, *svc)
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
		testCases = append(testCases,
			testCase{
				name: tc.repo.Name + "/new repo",
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
				name: tc.repo.Name + "/new repo sources",
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
				name: tc.repo.Name + "/deleted repo source",
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
				name:    tc.repo.Name + "/deleted ALL repo sources",
				sourcer: repos.NewFakeSourcer(nil),
				store:   s,
				stored: repos.Repos{tc.repo.With(
					repos.Opt.RepoSources(tc.svc.URN(), svcdup.URN()),
				)},
				now: clock.Now,
				diff: repos.Diff{Deleted: repos.Repos{tc.repo.With(
					repos.Opt.RepoDeletedAt(clock.Time(1)),
				)}},
				err: "<nil>",
			},
			testCase{
				name:    tc.repo.Name + "/renamed repo is detected via external_id",
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
				name: tc.repo.Name + "/repo got renamed to another repo that gets deleted",
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
				name: tc.repo.Name + "/repo inserted with same name as another repo that gets deleted",
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
				name: tc.repo.Name + "/repo inserted with same name as repo without id",
				sourcer: repos.NewFakeSourcer(nil,
					repos.NewFakeSource(tc.svc.Clone(), nil,
						tc.repo,
					),
				),
				store: s,
				stored: repos.Repos{
					tc.repo.With(repos.Opt.RepoName("old-name")),  // same external id as sourced
					tc.repo.With(repos.Opt.RepoExternalID("bar")), // same name as sourced
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
				name:    tc.repo.Name + "/renamed repo which was deleted is detected and added",
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
				name: tc.repo.Name + "/repos have their names swapped",
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
				name: tc.repo.Name + "/case insensitive name",
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
				typ, ok := extsvc.ParseServiceType(tc.repo.ExternalRepo.ServiceType)
				if !ok {
					panic(fmt.Sprintf("test must be extended with new external service kind: %q", strings.ToLower(tc.repo.ExternalRepo.ServiceType)))
				}
				switch typ {
				case extsvc.TypeGitHub:
					update = &github.Repository{IsArchived: true}
				case extsvc.TypeGitLab:
					update = &gitlab.Project{Archived: true}
				case extsvc.TypeBitbucketServer:
					update = &bitbucketserver.Repo{Public: true}
				case extsvc.TypeBitbucketCloud:
					update = &bitbucketcloud.Repo{IsPrivate: true}
				case extsvc.TypeAWSCodeCommit:
					update = &awscodecommit.Repository{Description: "new description"}
				case extsvc.TypeOther, extsvc.TypeGitolite:
					return testCase{}
				default:
					panic(fmt.Sprintf("test must be extended with new external service kind: %q", strings.ToLower(tc.repo.ExternalRepo.ServiceType)))
				}

				return testCase{
					name: tc.repo.Name + "/metadata update",
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
					cloned := tc.stored.Clone()
					if err := st.InsertRepos(ctx, cloned...); err != nil {
						t.Fatalf("failed to prepare store: %v", err)
					}
				}

				syncer := &repos.Syncer{
					Store:   st,
					Sourcer: tc.sourcer,
					Now:     now,
				}
				err := syncer.Sync(ctx)

				if have, want := fmt.Sprint(err), tc.err; have != want {
					t.Errorf("have error %q, want %q", have, want)
				}

				if err != nil {
					return
				}

				if st != nil {
					var want, have repos.Repos
					want.Concat(tc.diff.Added, tc.diff.Modified, tc.diff.Unmodified)
					have, _ = st.ListRepos(ctx, repos.StoreListReposArgs{})

					want = want.With(repos.Opt.RepoID(0))
					have = have.With(repos.Opt.RepoID(0))
					sort.Sort(want)
					sort.Sort(have)

					repos.Assert.ReposEqual(want...)(t, have)
				}
			}))
		}
	}
}

func TestSync_SyncSubset(t *testing.T) {
	t.Parallel()

	testSyncSubset(t, new(repos.FakeStore))(t)
}

func testSyncSubset(t *testing.T, s repos.Store) func(*testing.T) {
	clock := repos.NewFakeClock(time.Now(), time.Second)

	servicesPerKind := createExternalServices(t, s)

	repo := &repos.Repo{
		ID:          0, // explicitly make default value for sourced repo
		Name:        "github.com/foo/bar",
		Description: "The description",
		Language:    "barlang",
		Archived:    false,
		Fork:        false,
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAA==",
			ServiceType: extsvc.TypeGitHub,
			ServiceID:   "https://github.com/",
		},
		Sources: map[string]*repos.SourceInfo{
			servicesPerKind[extsvc.KindGitHub].URN(): {
				ID:       servicesPerKind[extsvc.KindGitHub].URN(),
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
					if err := st.InsertRepos(ctx, tc.stored.Clone()...); err != nil {
						t.Fatalf("failed to prepare store: %v", err)
					}
				}

				clock := clock
				syncer := &repos.Syncer{
					Store: st,
					Now:   clock.Now,
				}
				err := syncer.SyncSubset(ctx, tc.sourced.Clone()...)
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
		// 🚨 SECURITY: Tests to ensure we detect repository visibility changes.
		{
			name: "repository visiblity changed",
			store: repos.Repos{
				{Name: "foo", ExternalRepo: eid("1"), Description: "foo", Private: false},
			},
			source: repos.Repos{
				{Name: "foo", ExternalRepo: eid("1"), Description: "foo", Private: true},
			},
			diff: repos.Diff{
				Modified: repos.Repos{
					{Name: "foo", ExternalRepo: eid("1"), Description: "foo", Private: true},
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

func TestSync_Run(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())

	// some ceremony to setup metadata on our test repos
	svc := &repos.ExternalService{ID: 1, Kind: extsvc.KindGitHub}
	mk := func(name string) *repos.Repo {
		return &repos.Repo{
			Name:     name,
			Metadata: &github.Repository{},
			ExternalRepo: api.ExternalRepoSpec{
				ID:          name,
				ServiceID:   "https://github.com",
				ServiceType: svc.Kind,
			},
		}
	}

	// Our test will have 1 initial repo, and discover a new repo on sourcing.
	stored := repos.Repos{mk("initial")}.With(repos.Opt.RepoSources(svc.URN()))
	sourced := repos.Repos{mk("initial"), mk("new")}

	syncer := &repos.Syncer{
		Store:        &repos.FakeStore{},
		Sourcer:      repos.NewFakeSourcer(nil, repos.NewFakeSource(svc, nil, sourced...)),
		Synced:       make(chan repos.Diff),
		SubsetSynced: make(chan repos.Diff),
		Now:          time.Now,
	}

	// Initial repos in store
	syncer.Store.InsertRepos(ctx, stored...)

	done := make(chan struct{})
	go func() {
		defer close(done)
		syncer.Run(ctx, func() time.Duration { return 0 })
	}()

	// Ignore fields store adds
	ignore := cmpopts.IgnoreFields(repos.Repo{}, "ID", "CreatedAt", "UpdatedAt", "Sources")

	// The first thing sent down Synced is the list of repos in store.
	diff := <-syncer.Synced
	if d := cmp.Diff(repos.Diff{Unmodified: stored}, diff, ignore); d != "" {
		t.Fatalf("initial Synced mismatch (-want +got):\n%s", d)
	}

	// Next up it should find the new repo and send it down SubsetSynced
	diff = <-syncer.SubsetSynced
	if d := cmp.Diff(repos.Diff{Added: repos.Repos{mk("new")}}, diff, ignore); d != "" {
		t.Fatalf("SubsetSynced mismatch (-want +got):\n%s", d)
	}

	// Finally we get the final diff, which will have everything listed as
	// Unmodified since we added when we did SubsetSynced.
	diff = <-syncer.Synced
	if d := cmp.Diff(repos.Diff{Unmodified: sourced}, diff, ignore); d != "" {
		t.Fatalf("final Synced mismatch (-want +got):\n%s", d)
	}

	// We check synced again to test us going around the Run loop 2 times in
	// total.
	diff = <-syncer.Synced
	if d := cmp.Diff(repos.Diff{Unmodified: sourced}, diff, ignore); d != "" {
		t.Fatalf("second final Synced mismatch (-want +got):\n%s", d)
	}

	// Cancel context and the run loop should stop
	cancel()
	<-done
}
