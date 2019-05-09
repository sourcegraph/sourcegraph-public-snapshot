package repos_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/awscodecommit"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/github"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/gitolite"
	"github.com/sourcegraph/sourcegraph/pkg/jsonc"
)

func TestEnabledStateDeprecationMigration(t *testing.T) {
	testEnabledStateDeprecationMigration(new(repos.FakeStore))(t)
}

func testEnabledStateDeprecationMigration(store repos.Store) func(*testing.T) {
	clock := repos.NewFakeClock(time.Now(), 0)
	now := clock.Now()

	githubService := repos.ExternalService{
		ID:          1,
		Kind:        "GITHUB",
		DisplayName: "github.com - test",
		Config: formatJSON(`
		{
			// Some comment
			"url": "https://github.com",
			"token": "secret",
			"repositoryQuery": ["none"]
		}`),
	}

	githubRepo := repos.Repo{
		Name:    "github.com/foo/bar",
		Enabled: false,
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "bar",
			ServiceType: "github",
			ServiceID:   "http://github.com",
		},
		Sources: map[string]*repos.SourceInfo{},
		Metadata: &github.Repository{
			ID:            "bar",
			NameWithOwner: "foo/bar",
		},
	}

	gitlabService := repos.ExternalService{
		ID:          2,
		Kind:        "GITLAB",
		DisplayName: "gitlab.com - test",
		Config: formatJSON(`
		{
			// Some comment
			"url": "https://gitlab.com",
			"token": "secret",
			"projectQuery": ["none"]
		}`),
	}

	gitlabRepo := repos.Repo{
		Name:    "gitlab.com/foo/bar",
		Enabled: false,
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "1",
			ServiceType: "gitlab",
			ServiceID:   "http://gitlab.com",
		},
		Sources: map[string]*repos.SourceInfo{},
		Metadata: &gitlab.Project{
			ProjectCommon: gitlab.ProjectCommon{
				ID:                1,
				PathWithNamespace: "foo/bar",
			},
		},
	}

	bitbucketServerService := repos.ExternalService{
		ID:          3,
		Kind:        "BITBUCKETSERVER",
		DisplayName: "Bitbucket Server - Test",
		Config: formatJSON(`
		{
			// Some comment
			"url": "https://bitbucketserver.mycorp.com",
			"username": "admin",
			"token": "secret",
			"repositoryQuery": ["none"]
		}`),
	}

	bitbucketServerRepo := repos.Repo{
		Name:    "bitbucketserver.mycorp.com/foo/bar",
		Enabled: false,
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "1",
			ServiceType: "bitbucketServer",
			ServiceID:   "http://bitbucketserver.mycorp.com",
		},
		Sources: map[string]*repos.SourceInfo{},
		Metadata: &bitbucketserver.Repo{
			ID:   1,
			Slug: "bar",
			Project: &bitbucketserver.Project{
				Key: "foo",
			},
		},
	}

	awsCodeCommitService := repos.ExternalService{
		ID:          4,
		Kind:        "AWSCODECOMMIT",
		DisplayName: "AWS CodeCommit - Test",
		Config: formatJSON(`
		{
			"region": "us-west-1",
			"accessKeyID": "secret-accessKeyID",
			"secretAccessKey": "secret-secretAccessKey",
			"gitCredentials": {"username": "user", "password": "pw"},
		}`),
	}

	awsCodeCommitRepo := repos.Repo{
		Name:    "git-codecommit.us-west-1.amazonaws.com/stripe-go",
		Enabled: false,
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "f001337a-3450-46fd-b7d2-650c0EXAMPLE",
			ServiceType: "awscodecommit",
			ServiceID:   "arn:aws:codecommit:us-west-1:999999999999:",
		},
		Sources: map[string]*repos.SourceInfo{},
		Metadata: &awscodecommit.Repository{
			ID:   "f001337a-3450-46fd-b7d2-650c0EXAMPLE",
			Name: "stripe-go",
		},
	}

	gitoliteService := repos.ExternalService{
		ID:          5,
		Kind:        "GITOLITE",
		DisplayName: "Gitolite - Test",
		Config: formatJSON(`
		{
			// Some comment
			"prefix": "/",
			"host": "git@gitolite.mycorp.com"
		}`),
	}

	gitoliteRepo := repos.Repo{
		Name:    "gitolite.mycorp.com/bar",
		Enabled: false,
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "bar",
			ServiceType: "gitolite",
			ServiceID:   "git@gitolite.mycorp.com",
		},
		Sources: map[string]*repos.SourceInfo{},
		Metadata: &gitolite.Repo{
			Name: "bar",
			URL:  "git@gitolite.mycorp.com:bar.git",
		},
	}

	services := repos.ExternalServices{
		&githubService,
		&gitlabService,
		&bitbucketServerService,
		&awsCodeCommitService,
		&gitoliteService,
	}

	repositories := repos.Repos{
		&githubRepo,
		&gitlabRepo,
		&bitbucketServerRepo,
		&awsCodeCommitRepo,
		&gitoliteRepo,
	}

	type stored struct {
		services repos.ExternalServices
		repos    repos.Repos
	}

	type assert struct {
		services repos.ExternalServicesAssertion
		repos    repos.ReposAssertion
	}

	type testCase struct {
		name   string
		stored stored
		assert assert
	}

	testCases := []testCase{{
		name: "initialRepositoryEnablement gets deleted",
		stored: stored{
			services: services.With(func(e *repos.ExternalService) {
				var err error
				e.Config, err = jsonc.Edit(e.Config, false, "initialRepositoryEnablement")
				if err != nil {
					panic(err)
				}
			}),
		},
		assert: assert{
			services: repos.Assert.ExternalServicesEqual(services.With(
				repos.Opt.ExternalServiceModifiedAt(now),
			)...),
		},
	}, {
		// When the repos don't have any sources, we exclude them from
		// all external services of the same kind.
		name: "disabled repos are excluded from all sources",
		stored: stored{
			repos:    repositories,
			services: services,
		},
		assert: assert{
			repos: repos.Assert.ReposEqual(), // All disabled repos are deleted.
			services: repos.Assert.ExternalServicesEqual(services.With(
				repos.Opt.ExternalServiceModifiedAt(now),
				func(e *repos.ExternalService) {
					if err := e.Exclude(repositories...); err != nil {
						panic(err)
					}
				},
			)...),
		},
	}, {
		name: "disabled repos are excluded in all of its existing sources",
		stored: stored{
			repos: repos.Repos{
				githubRepo.With(repos.Opt.RepoSources(githubService.URN())),
				gitlabRepo.With(repos.Opt.RepoSources(gitlabService.URN())),
				bitbucketServerRepo.With(repos.Opt.RepoSources(bitbucketServerService.URN())),
				awsCodeCommitRepo.With(repos.Opt.RepoSources(awsCodeCommitService.URN())),
				gitoliteRepo.With(repos.Opt.RepoSources(gitoliteService.URN())),
			},
			services: services,
		},
		assert: assert{
			repos: repos.Assert.ReposEqual(), // All disabled repos are deleted.
			services: repos.Assert.ExternalServicesEqual(services.With(
				repos.Opt.ExternalServiceModifiedAt(now),
				func(e *repos.ExternalService) {
					if err := e.Exclude(repositories...); err != nil {
						panic(err)
					}
				},
			)...),
		},
	}, {
		name: "enabled repos are ignored",
		stored: stored{
			repos:    repositories.With(repos.Opt.RepoEnabled(true)),
			services: services,
		},
		assert: assert{
			repos: repos.Assert.ReposEqual(
				repositories.With(repos.Opt.RepoEnabled(true))...,
			),
			services: repos.Assert.ExternalServicesEqual(services...),
		},
	}}

	return func(t *testing.T) {
		t.Helper()

		for _, tc := range testCases {
			tc := tc
			ctx := context.Background()

			t.Run(tc.name, transact(ctx, store, func(t testing.TB, tx repos.Store) {
				if err := tx.UpsertRepos(ctx, tc.stored.repos.Clone()...); err != nil {
					t.Fatalf("failed to prepare store: %v", err)
				}

				if err := tx.UpsertExternalServices(ctx, tc.stored.services.Clone()...); err != nil {
					t.Fatalf("failed to prepare store: %v", err)
				}

				err := repos.EnabledStateDeprecationMigration(clock.Now).Run(ctx, tx)
				if err != nil {
					t.Fatalf("error: %v", err)
				}

				if tc.assert.services != nil {
					svcs, err := tx.ListExternalServices(ctx, repos.StoreListExternalServicesArgs{})
					if err != nil {
						t.Fatal(err)
					}
					tc.assert.services(t, svcs)
				}

				if tc.assert.repos != nil {
					rs, err := tx.ListRepos(ctx, repos.StoreListReposArgs{})
					if err != nil {
						t.Fatal(err)
					}
					tc.assert.repos(t, rs)
				}
			}))
		}

	}
}

func TestGithubSetDefaultRepositoryQueryMigration(t *testing.T) {
	t.Parallel()
	testGithubSetDefaultRepositoryQueryMigration(new(repos.FakeStore))(t)
}

func testGithubSetDefaultRepositoryQueryMigration(store repos.Store) func(*testing.T) {
	githubDotCom := repos.ExternalService{
		Kind:        "GITHUB",
		DisplayName: "Github.com - Test",
		Config: formatJSON(`
			{
				// Some comment
				"url": "https://github.com"
			}
		`),
	}

	githubNone := repos.ExternalService{
		Kind:        "GITHUB",
		DisplayName: "Github.com - Test",
		Config: formatJSON(`
			{
				// Some comment
				"url": "https://github.com",
				"repositoryQuery": ["none"]
			}
		`),
	}

	githubEnterprise := repos.ExternalService{
		Kind:        "GITHUB",
		DisplayName: "Github Enterprise - Test",
		Config: formatJSON(`
			{
				// Some comment
				"url": "https://github.mycorp.com"
			}
		`),
	}

	gitlab := repos.ExternalService{
		Kind:        "GITLAB",
		DisplayName: "Gitlab - Test",
		Config:      formatJSON(`{"url": "https://gitlab.com"}`),
	}

	clock := repos.NewFakeClock(time.Now(), 0)

	return func(t *testing.T) {
		t.Helper()

		for _, tc := range []struct {
			name   string
			stored repos.ExternalServices
			assert repos.ExternalServicesAssertion
			err    string
		}{
			{
				name:   "no external services",
				stored: repos.ExternalServices{},
				assert: repos.Assert.ExternalServicesEqual(),
				err:    "<nil>",
			},
			{
				name:   "non-github services are left unchanged",
				stored: repos.ExternalServices{&gitlab},
				assert: repos.Assert.ExternalServicesEqual(&gitlab),
				err:    "<nil>",
			},
			{
				name:   "github services with repositoryQuery set are left unchanged",
				stored: repos.ExternalServices{&githubNone},
				assert: repos.Assert.ExternalServicesEqual(&githubNone),
				err:    "<nil>",
			},
			{
				name:   "github.com services are set to affiliated",
				stored: repos.ExternalServices{&githubDotCom},
				assert: repos.Assert.ExternalServicesEqual(
					githubDotCom.With(
						repos.Opt.ExternalServiceModifiedAt(clock.Time(0)),
						func(e *repos.ExternalService) {
							e.Config = formatJSON(`
								{
									// Some comment
									"url": "https://github.com",
									"repositoryQuery": ["affiliated"]
								}
							`)
						},
					),
				),
				err: "<nil>",
			},
			{
				name:   "github enterprise services are set to public and affiliated",
				stored: repos.ExternalServices{&githubEnterprise},
				assert: repos.Assert.ExternalServicesEqual(
					githubEnterprise.With(
						repos.Opt.ExternalServiceModifiedAt(clock.Time(0)),
						func(e *repos.ExternalService) {
							e.Config = formatJSON(`
								{
									// Some comment
									"url": "https://github.mycorp.com",
									"repositoryQuery": ["affiliated", "public"]
								}
							`)
						},
					),
				),
				err: "<nil>",
			},
		} {
			tc := tc
			ctx := context.Background()

			t.Run(tc.name, transact(ctx, store, func(t testing.TB, tx repos.Store) {
				if err := tx.UpsertExternalServices(ctx, tc.stored.Clone()...); err != nil {
					t.Fatalf("failed to prepare store: %v", err)
				}

				err := repos.GithubSetDefaultRepositoryQueryMigration(clock.Now).Run(ctx, tx)
				if have, want := fmt.Sprint(err), tc.err; have != want {
					t.Errorf("error:\nhave: %v\nwant: %v", have, want)
				}

				es, err := tx.ListExternalServices(ctx, repos.StoreListExternalServicesArgs{})
				if err != nil {
					t.Fatal(err)
				}

				if tc.assert != nil {
					tc.assert(t, es)
				}
			}))
		}

	}
}

func TestGitLabSetDefaultProjectQueryMigration(t *testing.T) {
	t.Parallel()
	testGitLabSetDefaultProjectQueryMigration(new(repos.FakeStore))(t)
}

func testGitLabSetDefaultProjectQueryMigration(store repos.Store) func(*testing.T) {
	github := repos.ExternalService{
		Kind:        "GITHUB",
		DisplayName: "Github.com - Test",
		Config: formatJSON(`
			{
				// Some comment
				"url": "https://github.com"
			}
		`),
	}

	gitlabNone := repos.ExternalService{
		Kind:        "GITLAB",
		DisplayName: "Gitlab - Test",
		Config: formatJSON(`
			{
				// Some comment
				"url": "https://gitlab.com",
				"projectQuery": ["none"]
			}
		`),
	}

	gitlab := repos.ExternalService{
		Kind:        "GITLAB",
		DisplayName: "Gitlab - Test",
		Config: formatJSON(`
			{
				// Some comment
				"url": "https://gitlab.com"
			}
		`),
	}

	clock := repos.NewFakeClock(time.Now(), 0)

	return func(t *testing.T) {
		t.Helper()

		for _, tc := range []struct {
			name   string
			stored repos.ExternalServices
			assert repos.ExternalServicesAssertion
			err    string
		}{
			{
				name:   "no external services",
				stored: repos.ExternalServices{},
				assert: repos.Assert.ExternalServicesEqual(),
				err:    "<nil>",
			},
			{
				name:   "non-gitlab services are left unchanged",
				stored: repos.ExternalServices{&github},
				assert: repos.Assert.ExternalServicesEqual(&github),
				err:    "<nil>",
			},
			{
				name:   "gitlab services with projectQuery set are left unchanged",
				stored: repos.ExternalServices{&gitlabNone},
				assert: repos.Assert.ExternalServicesEqual(&gitlabNone),
				err:    "<nil>",
			},
			{
				name:   "gitlab services are set to ?membership=true",
				stored: repos.ExternalServices{&gitlab},
				assert: repos.Assert.ExternalServicesEqual(
					gitlab.With(
						repos.Opt.ExternalServiceModifiedAt(clock.Time(0)),
						func(e *repos.ExternalService) {
							e.Config = formatJSON(`
								{
									// Some comment
									"url": "https://gitlab.com",
									"projectQuery": ["?membership=true"]
								}
							`)
						},
					),
				),
				err: "<nil>",
			},
		} {
			tc := tc
			ctx := context.Background()

			t.Run(tc.name, transact(ctx, store, func(t testing.TB, tx repos.Store) {
				if err := tx.UpsertExternalServices(ctx, tc.stored.Clone()...); err != nil {
					t.Fatalf("failed to prepare store: %v", err)
				}

				err := repos.GitLabSetDefaultProjectQueryMigration(clock.Now).Run(ctx, tx)
				if have, want := fmt.Sprint(err), tc.err; have != want {
					t.Errorf("error:\nhave: %v\nwant: %v", have, want)
				}

				es, err := tx.ListExternalServices(ctx, repos.StoreListExternalServicesArgs{})
				if err != nil {
					t.Fatal(err)
				}

				if tc.assert != nil {
					tc.assert(t, es)
				}
			}))
		}

	}
}

func TestBitbucketServerSetDefaultRepositoryQueryMigration(t *testing.T) {
	t.Parallel()
	testBitbucketServerSetDefaultRepositoryQueryMigration(new(repos.FakeStore))(t)
}

func testBitbucketServerSetDefaultRepositoryQueryMigration(store repos.Store) func(*testing.T) {
	bitbucketsrv := repos.ExternalService{
		Kind:        "BITBUCKETSERVER",
		DisplayName: "BitbucketServer - Test",
		Config: formatJSON(`
			{
				// Some comment
				"url": "https://bitbucketserver.mycorp.com",
				"username": "admin",
				"token": "secret"
			}
		`),
	}

	bitbucketsrvNone := repos.ExternalService{
		Kind:        "BITBUCKETSERVER",
		DisplayName: "BitbucketServer - Test",
		Config: formatJSON(`
			{
				// Some comment
				"url": "https://bitbucketserver.mycorp.com",
				"username": "admin",
				"token": "secret",
				"repositoryQuery": ["none"]
			}
		`),
	}

	gitlab := repos.ExternalService{
		Kind:        "GITLAB",
		DisplayName: "Gitlab - Test",
		Config:      formatJSON(`{"url": "https://gitlab.com"}`),
	}

	clock := repos.NewFakeClock(time.Now(), 0)

	return func(t *testing.T) {
		t.Helper()

		for _, tc := range []struct {
			name   string
			stored repos.ExternalServices
			assert repos.ExternalServicesAssertion
			err    string
		}{
			{
				name:   "no external services",
				stored: repos.ExternalServices{},
				assert: repos.Assert.ExternalServicesEqual(),
				err:    "<nil>",
			},
			{
				name:   "non-bitbucketserver services are left unchanged",
				stored: repos.ExternalServices{&gitlab},
				assert: repos.Assert.ExternalServicesEqual(&gitlab),
				err:    "<nil>",
			},
			{
				name:   "bitbucketserver services with repositoryQuery set are left unchanged",
				stored: repos.ExternalServices{&bitbucketsrvNone},
				assert: repos.Assert.ExternalServicesEqual(&bitbucketsrvNone),
				err:    "<nil>",
			},
			{
				name:   "bitbucketserver services are migrated",
				stored: repos.ExternalServices{&bitbucketsrv},
				assert: repos.Assert.ExternalServicesEqual(
					bitbucketsrv.With(
						repos.Opt.ExternalServiceModifiedAt(clock.Time(0)),
						func(e *repos.ExternalService) {
							var err error
							e.Config, err = jsonc.Edit(e.Config,
								[]string{"?visibility=private", "?visibility=public"},
								"repositoryQuery",
							)

							if err != nil {
								panic(err)
							}
						},
					),
				),
				err: "<nil>",
			},
		} {
			tc := tc
			ctx := context.Background()

			t.Run(tc.name, transact(ctx, store, func(t testing.TB, tx repos.Store) {
				if err := tx.UpsertExternalServices(ctx, tc.stored.Clone()...); err != nil {
					t.Fatalf("failed to prepare store: %v", err)
				}

				err := repos.BitbucketServerSetDefaultRepositoryQueryMigration(clock.Now).Run(ctx, tx)
				if have, want := fmt.Sprint(err), tc.err; have != want {
					t.Errorf("error:\nhave: %v\nwant: %v", have, want)
				}

				es, err := tx.ListExternalServices(ctx, repos.StoreListExternalServicesArgs{})
				if err != nil {
					t.Fatal(err)
				}

				if tc.assert != nil {
					tc.assert(t, es)
				}
			}))
		}

	}
}

func TestBitbucketServerUsernameMigration(t *testing.T) {
	t.Parallel()
	testBitbucketServerUsernameMigration(new(repos.FakeStore))(t)
}

func testBitbucketServerUsernameMigration(store repos.Store) func(*testing.T) {
	bitbucketsrv := repos.ExternalService{
		Kind:        "BITBUCKETSERVER",
		DisplayName: "BitbucketServer - Test",
		Config: formatJSON(`
			{
				// Some comment
				"url": "https://admin@bitbucketserver.mycorp.com",
				"token": "secret"
			}
		`),
	}

	bitbucketsrvNoUsername := repos.ExternalService{
		Kind:        "BITBUCKETSERVER",
		DisplayName: "BitbucketServer - Test",
		Config: formatJSON(`
			{
				// Some comment
				"url": "https://bitbucketserver.mycorp.com",
				"token": "secret"
			}
		`),
	}

	bitbucketsrvWithUsername := repos.ExternalService{
		Kind:        "BITBUCKETSERVER",
		DisplayName: "BitbucketServer - Test",
		Config: formatJSON(`
			{
				// Some comment
				"url": "https://bitbucketserver.mycorp.com",
				"username": "admin",
				"token": "secret",
			}
		`),
	}

	gitlab := repos.ExternalService{
		Kind:        "GITLAB",
		DisplayName: "Gitlab - Test",
		Config:      formatJSON(`{"url": "https://gitlab.com"}`),
	}

	clock := repos.NewFakeClock(time.Now(), 0)

	return func(t *testing.T) {
		t.Helper()

		for _, tc := range []struct {
			name   string
			stored repos.ExternalServices
			assert repos.ExternalServicesAssertion
			err    string
		}{
			{
				name:   "no external services",
				stored: repos.ExternalServices{},
				assert: repos.Assert.ExternalServicesEqual(),
				err:    "<nil>",
			},
			{
				name:   "non-bitbucketserver services are left unchanged",
				stored: repos.ExternalServices{&gitlab},
				assert: repos.Assert.ExternalServicesEqual(&gitlab),
				err:    "<nil>",
			},
			{
				name:   "bitbucketserver services with username set are left unchanged",
				stored: repos.ExternalServices{&bitbucketsrvWithUsername},
				assert: repos.Assert.ExternalServicesEqual(&bitbucketsrvWithUsername),
				err:    "<nil>",
			},
			{
				name:   "bitbucketserver services without username in url are left unchanged",
				stored: repos.ExternalServices{&bitbucketsrvNoUsername},
				assert: repos.Assert.ExternalServicesEqual(&bitbucketsrvNoUsername),
				err:    "<nil>",
			},
			{
				name:   "bitbucketserver services without username are migrated",
				stored: repos.ExternalServices{&bitbucketsrv},
				assert: repos.Assert.ExternalServicesEqual(
					bitbucketsrv.With(
						repos.Opt.ExternalServiceModifiedAt(clock.Time(0)),
						func(e *repos.ExternalService) {
							var err error
							e.Config, err = jsonc.Edit(e.Config, "admin", "username")
							if err != nil {
								panic(err)
							}
						},
					),
				),
				err: "<nil>",
			},
		} {
			tc := tc
			ctx := context.Background()

			t.Run(tc.name, transact(ctx, store, func(t testing.TB, tx repos.Store) {
				if err := tx.UpsertExternalServices(ctx, tc.stored.Clone()...); err != nil {
					t.Fatalf("failed to prepare store: %v", err)
				}

				err := repos.BitbucketServerUsernameMigration(clock.Now).Run(ctx, tx)
				if have, want := fmt.Sprint(err), tc.err; have != want {
					t.Errorf("error:\nhave: %v\nwant: %v", have, want)
				}

				es, err := tx.ListExternalServices(ctx, repos.StoreListExternalServicesArgs{})
				if err != nil {
					t.Fatal(err)
				}

				if tc.assert != nil {
					tc.assert(t, es)
				}
			}))
		}

	}
}
func formatJSON(s string) string {
	formatted, err := jsonc.Format(s, true, 2)
	if err != nil {
		panic(err)
	}
	return formatted
}
