package repos_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/github"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/gitlab"
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

	included := func(rs ...*repos.Repo) func(*repos.ExternalService) {
		return func(e *repos.ExternalService) {
			if err := e.Include(rs...); err != nil {
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
			"token": "secret"
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
		Sources:  map[string]*repos.SourceInfo{},
		Metadata: new(github.Repository),
	}

	gitlabService := repos.ExternalService{
		ID:          2,
		Kind:        "GITLAB",
		DisplayName: "gitlab.com - test",
		Config: formatJSON(`
		{
			// Some comment
			"url": "https://gitlab.com",
			"token": "secret"
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
		Sources:  map[string]*repos.SourceInfo{},
		Metadata: new(gitlab.Project),
	}

	bitbucketServerService := repos.ExternalService{
		ID:          2,
		Kind:        "BITBUCKETSERVER",
		DisplayName: "Bitbucket Server - Test",
		Config: formatJSON(`
		{
			// Some comment
			"url": "https://bitbucketserver.mycorp.com",
			"token": "secret"
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
		Sources:  map[string]*repos.SourceInfo{},
		Metadata: new(bitbucketserver.Repo),
	}

	var testCases []testCase
	for _, k := range []struct {
		svc  repos.ExternalService
		repo repos.Repo
	}{
		{svc: githubService, repo: githubRepo},
		{svc: gitlabService, repo: gitlabRepo},
		{svc: bitbucketServerService, repo: bitbucketServerRepo},
	} {
		repo, svc := k.repo, k.svc
		testCases = append(testCases,
			testCase{
				name:   "disabled: was deleted, got added, then excluded",
				stored: repos.Repos{repo.With(repos.Opt.RepoDeletedAt(now))},
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
				name:    "disabled: was deleted and is still deleted, got excluded",
				stored:  repos.Repos{repo.With(repos.Opt.RepoDeletedAt(now))},
				sourcer: repos.NewFakeSourcer(nil, repos.NewFakeSource(svc.Clone(), nil)),
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
				name: "enabled: was not deleted and got deleted, then included",
				stored: repos.Repos{repo.With(func(r *repos.Repo) {
					r.Enabled = true
				})},
				sourcer: repos.NewFakeSourcer(nil, repos.NewFakeSource(svc.Clone(), nil)),
				svcs: repos.Assert.ExternalServicesEqual(svc.With(
					repos.Opt.ExternalServiceModifiedAt(now),
					included(&repo),
				)),
				err: "<nil>",
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
				name: "enabled: repo is included in all of its sources",
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
					svc.With(
						repos.Opt.ExternalServiceModifiedAt(now),
						included(&repo),
					),
					svc.With(
						repos.Opt.ExternalServiceID(23),
						repos.Opt.ExternalServiceModifiedAt(now),
						included(&repo),
					),
				),
				err: "<nil>",
			},
			testCase{
				name:    "disabled: repos are deleted",
				stored:  repos.Repos{&repo},
				sourcer: repos.NewFakeSourcer(nil, repos.NewFakeSource(svc.Clone(), nil)),
				repos: repos.Assert.ReposEqual(repo.With(
					func(r *repos.Repo) {
						r.DeletedAt = now
						r.Enabled = true
					},
				)),
				err: "<nil>",
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
					t.Errorf("failed to prepare store: %v", err)
					return
				}

				err := repos.EnabledStateDeprecationMigration(tc.sourcer, clock.Now).Run(ctx, tx)
				if have, want := fmt.Sprint(err), tc.err; have != want {
					t.Errorf("error:\nhave: %v\nwant: %v", have, want)
					return
				}

				if tc.svcs != nil {
					svcs, err := tx.ListExternalServices(ctx)
					if err != nil {
						t.Error(err)
						return
					}
					tc.svcs(t, svcs)
				}

				if tc.repos != nil {
					rs, err := tx.ListRepos(ctx)
					if err != nil {
						t.Error(err)
						return
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
					t.Errorf("failed to prepare store: %v", err)
					return
				}

				err := repos.GithubSetDefaultRepositoryQueryMigration(clock.Now).Run(ctx, tx)
				if have, want := fmt.Sprint(err), tc.err; have != want {
					t.Errorf("error:\nhave: %v\nwant: %v", have, want)
				}

				es, err := tx.ListExternalServices(ctx)
				if err != nil {
					t.Error(err)
					return
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
					t.Errorf("failed to prepare store: %v", err)
					return
				}

				err := repos.GitLabSetDefaultProjectQueryMigration(clock.Now).Run(ctx, tx)
				if have, want := fmt.Sprint(err), tc.err; have != want {
					t.Errorf("error:\nhave: %v\nwant: %v", have, want)
				}

				es, err := tx.ListExternalServices(ctx)
				if err != nil {
					t.Error(err)
					return
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
				"url": "https://bitbucketserver.mycorp.com"
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
					t.Errorf("failed to prepare store: %v", err)
					return
				}

				err := repos.BitbucketServerSetDefaultRepositoryQueryMigration(clock.Now).Run(ctx, tx)
				if have, want := fmt.Sprint(err), tc.err; have != want {
					t.Errorf("error:\nhave: %v\nwant: %v", have, want)
				}

				es, err := tx.ListExternalServices(ctx)
				if err != nil {
					t.Error(err)
					return
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
