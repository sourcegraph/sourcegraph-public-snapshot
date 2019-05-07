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
	excluded := func(rs ...*repos.Repo) func(*repos.ExternalService) {
		return func(e *repos.ExternalService) {
			if err := e.Exclude(rs...); err != nil {
				panic(err)
			}
		}
	}

	clock := repos.NewFakeClock(time.Now(), 0)
	now := clock.Now()

	type testCase struct {
		name    string
		sourcer repos.Sourcer
		stored  repos.Repos
		svcs    repos.ExternalServicesAssertion
		repos   repos.ReposAssertion
		err     string
	}

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
		ID:          2,
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
		ID:          9,
		Kind:        "AWSCODECOMMIT",
		DisplayName: "AWS CodeCommit - Test",
		Config: formatJSON(`
		{
			"region": "us-west-1",
			"accessKeyID": "secret-accessKeyID",
			"secretAccessKey": "secret-secretAccessKey",
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
		Metadata: awscodecommit.Repository{
			ID:   "f001337a-3450-46fd-b7d2-650c0EXAMPLE",
			Name: "stripe-go",
		},
	}

	gitoliteService := repos.ExternalService{
		ID:          4,
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

	var testCases []testCase
	for _, k := range []struct {
		svc  repos.ExternalService
		repo repos.Repo
	}{
		{svc: githubService, repo: githubRepo},
		{svc: gitlabService, repo: gitlabRepo},
		{svc: bitbucketServerService, repo: bitbucketServerRepo},
		{svc: awsCodeCommitService, repo: awsCodeCommitRepo},
		{svc: gitoliteService, repo: gitoliteRepo},
	} {
		repo, svc := k.repo, k.svc
		testCases = append(testCases,
			testCase{
				name:   "enabled: was deleted, got added (enabled), not excluded",
				stored: repos.Repos{repo.With(repos.Opt.RepoDeletedAt(now))},
				sourcer: repos.NewFakeSourcer(nil,
					repos.NewFakeSource(svc.Clone(), nil, repo.With(repos.Opt.RepoEnabled(true))),
				),
				svcs: repos.Assert.ExternalServicesEqual(svc.Clone()),
				err:  "<nil>",
			},
			testCase{
				name:   "disabled: was not deleted and was not modified, got excluded",
				stored: repos.Repos{&repo},
				sourcer: repos.NewFakeSourcer(nil,
					repos.NewFakeSource(svc.Clone(), nil, repo.Clone()),
				),
				svcs: repos.Assert.ExternalServicesEqual(svc.With(
					repos.Opt.ExternalServiceModifiedAt(now),
					excluded(&repo),
				)),
				err: "<nil>",
			},
			testCase{
				name:   "disabled: was not deleted, got modified, then excluded",
				stored: repos.Repos{&repo},
				sourcer: repos.NewFakeSourcer(nil, repos.NewFakeSource(svc.Clone(), nil,
					repo.With(func(r *repos.Repo) {
						r.Description = "some updated description"
					})),
				),
				svcs: repos.Assert.ExternalServicesEqual(svc.With(
					repos.Opt.ExternalServiceModifiedAt(now),
					excluded(&repo),
				)),
				err: "<nil>",
			},
			testCase{
				name: "enabled: was not deleted and is still not deleted, not included",
				stored: repos.Repos{repo.With(func(r *repos.Repo) {
					r.Enabled = true
				})},
				sourcer: repos.NewFakeSourcer(nil, repos.NewFakeSource(svc.Clone(), nil,
					repo.With(func(r *repos.Repo) {
						r.Enabled = true
					})),
				),
				svcs: repos.Assert.ExternalServicesEqual(svc.Clone()),
				err:  "<nil>",
			},
			testCase{
				name: "enabled: was not deleted and got deleted, not included",
				stored: repos.Repos{repo.With(func(r *repos.Repo) {
					r.Enabled = true
				})},
				sourcer: repos.NewFakeSourcer(nil, repos.NewFakeSource(svc.Clone(), nil)),
				svcs:    repos.Assert.ExternalServicesEqual(svc.Clone()),
				err:     "<nil>",
			},
			testCase{
				name:   "enabled: got added for the first time, so not included",
				stored: repos.Repos{},
				sourcer: repos.NewFakeSourcer(nil, repos.NewFakeSource(svc.Clone(), nil,
					repo.With(func(r *repos.Repo) {
						r.Enabled = true
					})),
				),
				svcs: repos.Assert.ExternalServicesEqual(svc.Clone()),
				err:  "<nil>",
			},
			testCase{
				name: "initialRepositoryEnablement gets deleted",
				sourcer: repos.NewFakeSourcer(nil, repos.NewFakeSource(
					svc.With(func(e *repos.ExternalService) {
						var err error
						e.Config, err = jsonc.Edit(e.Config, false, "initialRepositoryEnablement")
						if err != nil {
							panic(err)
						}
					}), nil,
				)),
				svcs: repos.Assert.ExternalServicesEqual(svc.With(
					repos.Opt.ExternalServiceModifiedAt(now),
				)),
				err: "<nil>",
			},
			testCase{
				name:   "disabled: repo is excluded in all of its sources",
				stored: repos.Repos{repo.With(repos.Opt.RepoDeletedAt(now))},
				sourcer: repos.NewFakeSourcer(nil,
					repos.NewFakeSource(svc.Clone(), nil, repo.Clone()),
					repos.NewFakeSource(svc.With(repos.Opt.ExternalServiceID(23)), nil, repo.Clone()),
				),
				svcs: repos.Assert.ExternalServicesEqual(
					svc.With(
						repos.Opt.ExternalServiceModifiedAt(now),
						excluded(&repo),
					),
					svc.With(
						repos.Opt.ExternalServiceID(23),
						repos.Opt.ExternalServiceModifiedAt(now),
						excluded(&repo),
					),
				),
				err: "<nil>",
			},
			testCase{
				name: "enabled: deleted repo is not included in any of its sources",
				stored: repos.Repos{repo.With(
					repos.Opt.RepoSources(
						svc.URN(),
						svc.With(repos.Opt.ExternalServiceID(23)).URN(),
					),
					func(r *repos.Repo) { r.Enabled = true },
				)},
				sourcer: repos.NewFakeSourcer(nil,
					repos.NewFakeSource(svc.Clone(), nil),
					repos.NewFakeSource(svc.With(repos.Opt.ExternalServiceID(23)), nil),
				),
				svcs: repos.Assert.ExternalServicesEqual(
					svc.Clone(),
					svc.With(repos.Opt.ExternalServiceID(23)),
				),
				err: "<nil>",
			},
			testCase{
				name:    "disabled: repos are deleted",
				stored:  repos.Repos{&repo},
				sourcer: repos.NewFakeSourcer(nil, repos.NewFakeSource(svc.Clone(), nil)),
				repos:   repos.Assert.ReposEqual(),
				err:     "<nil>",
			},
		)
	}

	return func(t *testing.T) {
		t.Helper()

		for _, tc := range testCases {
			tc := tc
			ctx := context.Background()

			t.Run(tc.name, transact(ctx, store, func(t testing.TB, tx repos.Store) {
				if err := tx.UpsertRepos(ctx, tc.stored.Clone()...); err != nil {
					t.Fatalf("failed to prepare store: %v", err)
				}

				err := repos.EnabledStateDeprecationMigration(tc.sourcer, clock.Now).Run(ctx, tx)
				if have, want := fmt.Sprint(err), tc.err; have != want {
					t.Fatalf("error:\nhave: %v\nwant: %v", have, want)
				}

				if tc.svcs != nil {
					svcs, err := tx.ListExternalServices(ctx, repos.StoreListExternalServicesArgs{})
					if err != nil {
						t.Fatal(err)
					}
					tc.svcs(t, svcs)
				}

				if tc.repos != nil {
					rs, err := tx.ListRepos(ctx, repos.StoreListReposArgs{})
					if err != nil {
						t.Fatal(err)
					}
					tc.repos(t, rs)
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
