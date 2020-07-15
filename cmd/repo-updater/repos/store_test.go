package repos_test

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/google/go-cmp/cmp"
	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/awscodecommit"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitolite"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

func TestFakeStore(t *testing.T) {
	t.Parallel()

	lg := log15.New()
	lg.SetHandler(log15.DiscardHandler())

	for _, tc := range []struct {
		name string
		test func(repos.Store) func(*testing.T)
	}{
		{"ListExternalServices", testStoreListExternalServices},
		{"UpsertExternalServices", testStoreUpsertExternalServices},
		{"ListRepos", testStoreListRepos},
		{"ListRepos_Pagination", testStoreListReposPagination},
		{"UpsertRepos", testStoreUpsertRepos},
		{"SetClonedRepos", testStoreSetClonedRepos},
	} {
		t.Run(tc.name, tc.test(repos.NewObservedStore(
			new(repos.FakeStore),
			lg,
			repos.NewStoreMetrics(),
			trace.Tracer{Tracer: opentracing.GlobalTracer()},
		)))
	}
}

func testStoreListExternalServicesByRepos(store repos.Store) func(*testing.T) {
	return func(t *testing.T) {
		t.Helper()

		ctx := context.Background()
		clock := repos.NewFakeClock(time.Now(), 0)
		now := clock.Now()

		t.Run("", transact(ctx, store, func(t testing.TB, tx repos.Store) {
			github := repos.ExternalService{
				Kind:        extsvc.KindGitHub,
				DisplayName: "Github - Test",
				Config:      `{"url": "https://github.com"}`,
				CreatedAt:   now,
				UpdatedAt:   now,
			}

			gitlab := repos.ExternalService{
				Kind:        extsvc.KindGitLab,
				DisplayName: "GitLab - Test",
				Config:      `{"url": "https://gitlab.com"}`,
				CreatedAt:   now,
				UpdatedAt:   now,
			}

			svcs := repos.ExternalServices{&github, &gitlab}

			if err := tx.UpsertExternalServices(ctx, svcs...); err != nil {
				t.Fatalf("failed to setup store: %v", err)
			}

			repositories := repos.Repos{
				{
					Name: "github.com/foo/bar",
					Sources: map[string]*repos.SourceInfo{
						fmt.Sprintf("extsvc:github:%d", github.ID): {},
					},
					ExternalRepo: api.ExternalRepoSpec{
						ID:          "bar",
						ServiceType: "github",
						ServiceID:   "http://github.com",
					},
				},
				{
					Name: "github.com/foo/baz",
					Sources: map[string]*repos.SourceInfo{
						fmt.Sprintf("extsvc:github:%d", github.ID): {},
					},
					ExternalRepo: api.ExternalRepoSpec{
						ID:          "baz",
						ServiceType: "github",
						ServiceID:   "http://github.com",
					},
				},
				{
					Name: "gitlab.com/foo/bar",
					Sources: map[string]*repos.SourceInfo{
						fmt.Sprintf("extsvc:gitlab:%d", gitlab.ID): {},
					},
					ExternalRepo: api.ExternalRepoSpec{
						ID:          "bar",
						ServiceType: extsvc.TypeGitLab,
						ServiceID:   "http://gitlab.com",
					},
				},
			}

			if err := tx.UpsertRepos(ctx, repositories...); err != nil {
				t.Fatalf("failed to setup store: %v", err)
			}

			opts := repos.StoreListExternalServicesArgs{
				RepoIDs: repositories.IDs(),
			}

			have, err := tx.ListExternalServices(ctx, opts)
			if err != nil {
				t.Fatal(err)
			}

			repos.Assert.ExternalServicesEqual(svcs...)(t, have)
		}))
	}
}

func testStoreListExternalServices(store repos.Store) func(*testing.T) {
	clock := repos.NewFakeClock(time.Now(), 0)
	now := clock.Now()

	github := repos.ExternalService{
		Kind:        extsvc.KindGitHub,
		DisplayName: "Github - Test",
		Config:      `{"url": "https://github.com"}`,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	gitlab := repos.ExternalService{
		Kind:        extsvc.KindGitLab,
		DisplayName: "GitLab - Test",
		Config:      `{"url": "https://gitlab.com"}`,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	bitbucketServer := repos.ExternalService{
		Kind:        extsvc.KindBitbucketServer,
		DisplayName: "Bitbucket Server - Test",
		Config:      `{"url": "https://bitbucketserver.mycorp.com"}`,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	awsCodeCommit := repos.ExternalService{
		Kind:        extsvc.KindAWSCodeCommit,
		DisplayName: "AWS CodeCommit - Test",
		Config:      `{"region": "us-west-1"}`,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	otherService := repos.ExternalService{
		Kind:        extsvc.KindOther,
		DisplayName: "Other code hosts",
		Config:      `{"url": "https://git-host.mycorp.com"}`,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	gitoliteService := repos.ExternalService{
		Kind:        extsvc.KindGitolite,
		DisplayName: "Gitolite Server - Test",
		Config:      `{"prefix": "/", "host": "git@gitolite.mycorp.com"}`,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	phabricatorService := repos.ExternalService{
		Kind:        extsvc.KindPhabricator,
		DisplayName: "Phabricator - Test",
		Config:      `{"url": "https://phab.org", "token": "foo"}`,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	svcs := repos.ExternalServices{
		&github,
		&gitlab,
		&bitbucketServer,
		&awsCodeCommit,
		&otherService,
		&gitoliteService,
		&phabricatorService,
	}

	type testCase struct {
		name   string
		args   func(stored repos.ExternalServices) repos.StoreListExternalServicesArgs
		stored repos.ExternalServices
		assert repos.ExternalServicesAssertion
		err    error
	}

	var testCases []testCase
	testCases = append(testCases,
		testCase{
			name: "returned kind is uppercase",
			args: func(repos.ExternalServices) repos.StoreListExternalServicesArgs {
				return repos.StoreListExternalServicesArgs{
					Kinds: svcs.Kinds(),
				}
			},
			stored: svcs,
			assert: repos.Assert.ExternalServicesEqual(svcs...),
		},
		testCase{
			name: "case-insensitive kinds",
			args: func(repos.ExternalServices) (args repos.StoreListExternalServicesArgs) {
				for _, kind := range svcs.Kinds() {
					args.Kinds = append(args.Kinds, strings.ToLower(kind))
				}
				return args
			},
			stored: svcs,
			assert: repos.Assert.ExternalServicesEqual(svcs...),
		},
		testCase{
			name:   "excludes soft deleted external services by default",
			stored: svcs.With(repos.Opt.ExternalServiceDeletedAt(now)),
			assert: repos.Assert.ExternalServicesEqual(),
		},
		testCase{
			name:   "results are in ascending order by id",
			stored: mkExternalServices(7, svcs...),
			assert: repos.Assert.ExternalServicesOrderedBy(
				func(a, b *repos.ExternalService) bool {
					return a.ID < b.ID
				},
			),
		},
		testCase{
			name:   "excludes phabricator by default",
			stored: svcs,
			assert: repos.Assert.ExternalServicesEqual(func() (es repos.ExternalServices) {
				for _, e := range svcs {
					if e.Kind != extsvc.KindPhabricator {
						es = append(es, e)
					}
				}
				return es
			}()...),
		},
		testCase{
			name:   "includes phabricator if specified in Kinds",
			stored: svcs,
			args: func(repos.ExternalServices) (args repos.StoreListExternalServicesArgs) {
				args.Kinds = []string{extsvc.KindPhabricator}
				return args
			},
			assert: repos.Assert.ExternalServicesEqual(&phabricatorService),
		},
	)

	testCases = append(testCases, testCase{
		name:   "returns svcs by their ids",
		stored: svcs,
		args: func(stored repos.ExternalServices) repos.StoreListExternalServicesArgs {
			return repos.StoreListExternalServicesArgs{
				IDs: []int64{stored[0].ID, stored[1].ID},
			}
		},
		assert: repos.Assert.ExternalServicesEqual(svcs[:2].Clone()...),
	})

	return func(t *testing.T) {
		t.Helper()

		for _, tc := range testCases {
			tc := tc
			ctx := context.Background()

			t.Run(tc.name, transact(ctx, store, func(t testing.TB, tx repos.Store) {
				stored := tc.stored.Clone()
				if err := tx.UpsertExternalServices(ctx, stored...); err != nil {
					t.Fatalf("failed to setup store: %v", err)
				}

				var args repos.StoreListExternalServicesArgs
				if tc.args != nil {
					args = tc.args(stored)
				}

				es, err := tx.ListExternalServices(ctx, args)
				if have, want := fmt.Sprint(err), fmt.Sprint(tc.err); have != want {
					t.Errorf("error:\nhave: %v\nwant: %v", have, want)
				}

				if tc.assert != nil {
					tc.assert(t, es)
				}
			}))
		}
	}
}

func testStoreUpsertExternalServices(store repos.Store) func(*testing.T) {
	clock := repos.NewFakeClock(time.Now(), 0)
	now := clock.Now()

	return func(t *testing.T) {
		t.Helper()

		github := repos.ExternalService{
			Kind:        extsvc.KindGitHub,
			DisplayName: "Github - Test",
			Config:      `{"url": "https://github.com"}`,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		gitlab := repos.ExternalService{
			Kind:        extsvc.KindGitLab,
			DisplayName: "GitLab - Test",
			Config:      `{"url": "https://gitlab.com"}`,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		bitbucketServer := repos.ExternalService{
			Kind:        extsvc.KindBitbucketServer,
			DisplayName: "Bitbucket Server - Test",
			Config:      `{"url": "https://bitbucketserver.mycorp.com"}`,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		awsCodeCommit := repos.ExternalService{
			Kind:        extsvc.KindAWSCodeCommit,
			DisplayName: "AWS CodeCommit - Test",
			Config:      `{"region": "us-west-1"}`,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		otherService := repos.ExternalService{
			Kind:        extsvc.KindOther,
			DisplayName: "Other code hosts",
			Config:      `{"url": "https://git-host.mycorp.com"}`,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		gitoliteService := repos.ExternalService{
			Kind:        extsvc.KindGitolite,
			DisplayName: "Gitolite Server - Test",
			Config:      `{"prefix": "/", "host": "git@gitolite.mycorp.com"}`,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		svcs := repos.ExternalServices{
			&github,
			&gitlab,
			&bitbucketServer,
			&awsCodeCommit,
			&otherService,
			&gitoliteService,
		}

		ctx := context.Background()

		t.Run("no external services", func(t *testing.T) {
			if err := store.UpsertExternalServices(ctx); err != nil {
				t.Fatalf("UpsertExternalServices error: %s", err)
			}
		})

		t.Run("many external services", transact(ctx, store, func(t testing.TB, tx repos.Store) {
			want := mkExternalServices(7, svcs...)

			if err := tx.UpsertExternalServices(ctx, want...); err != nil {
				t.Fatalf("UpsertExternalServices error: %s", err)
			}

			for _, e := range want {
				if e.Kind != strings.ToUpper(e.Kind) {
					t.Errorf("external service kind didn't get upper-cased: %q", e.Kind)
					break
				}
			}

			sort.Sort(want)

			have, err := tx.ListExternalServices(ctx, repos.StoreListExternalServicesArgs{
				Kinds: svcs.Kinds(),
			})
			if err != nil {
				t.Fatalf("ListExternalServices error: %s", err)
			}

			if diff := cmp.Diff(have, []*repos.ExternalService(want), cmpopts.EquateEmpty()); diff != "" {
				t.Fatalf("ListExternalServices:\n%s", diff)
			}

			now := clock.Now()
			suffix := "-updated"
			for _, r := range want {
				r.DisplayName += suffix
				r.Kind += suffix
				r.Config += suffix
				r.UpdatedAt = now
				r.CreatedAt = now
			}

			if err = tx.UpsertExternalServices(ctx, want...); err != nil {
				t.Errorf("UpsertExternalServices error: %s", err)
			} else if have, err = tx.ListExternalServices(ctx, repos.StoreListExternalServicesArgs{}); err != nil {
				t.Errorf("ListExternalServices error: %s", err)
			} else if diff := cmp.Diff(have, []*repos.ExternalService(want), cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("ListExternalServices:\n%s", diff)
			}

			want.Apply(repos.Opt.ExternalServiceDeletedAt(now))
			args := repos.StoreListExternalServicesArgs{}

			if err = tx.UpsertExternalServices(ctx, want.Clone()...); err != nil {
				t.Errorf("UpsertExternalServices error: %s", err)
			} else if have, err = tx.ListExternalServices(ctx, args); err != nil {
				t.Errorf("ListExternalServices error: %s", err)
			} else if diff := cmp.Diff(have, []*repos.ExternalService(nil), cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("ListExternalServices:\n%s", diff)
			}
		}))
	}
}

func testStoreUpsertRepos(store repos.Store) func(*testing.T) {
	clock := repos.NewFakeClock(time.Now(), 0)
	now := clock.Now()

	return func(t *testing.T) {
		t.Helper()

		kinds := []string{
			extsvc.KindGitHub,
			extsvc.KindGitLab,
			extsvc.KindBitbucketServer,
			extsvc.KindAWSCodeCommit,
			extsvc.KindOther,
			extsvc.KindGitolite,
		}

		github := repos.Repo{
			Name:        "github.com/foo/bar",
			URI:         "github.com/foo/bar",
			Description: "The description",
			Language:    "barlang",
			CreatedAt:   now,
			ExternalRepo: api.ExternalRepoSpec{
				ID:          "AAAAA==",
				ServiceType: "github",
				ServiceID:   "http://github.com",
			},
			Sources: map[string]*repos.SourceInfo{
				"extsvc:1": {
					ID:       "extsvc:1",
					CloneURL: "git@github.com:foo/bar.git",
				},
			},
			Metadata: new(github.Repository),
		}

		gitlab := repos.Repo{
			Name:        "gitlab.com/foo/bar",
			URI:         "gitlab.com/foo/bar",
			Description: "The description",
			Language:    "barlang",
			CreatedAt:   now,
			ExternalRepo: api.ExternalRepoSpec{
				ID:          "1234",
				ServiceType: extsvc.TypeGitLab,
				ServiceID:   "http://gitlab.com",
			},
			Sources: map[string]*repos.SourceInfo{
				"extsvc:2": {
					ID:       "extsvc:2",
					CloneURL: "git@gitlab.com:foo/bar.git",
				},
			},
			Metadata: new(gitlab.Project),
		}

		bitbucketServer := repos.Repo{
			Name:        "bitbucketserver.mycorp.com/foo/bar",
			URI:         "bitbucketserver.mycorp.com/foo/bar",
			Description: "The description",
			Language:    "barlang",
			CreatedAt:   now,
			ExternalRepo: api.ExternalRepoSpec{
				ID:          "1234",
				ServiceType: "bitbucketServer",
				ServiceID:   "http://bitbucketserver.mycorp.com",
			},
			Sources: map[string]*repos.SourceInfo{
				"extsvc:3": {
					ID:       "extsvc:3",
					CloneURL: "git@bitbucketserver.mycorp.com:foo/bar.git",
				},
			},
			Metadata: new(bitbucketserver.Repo),
		}

		awsCodeCommit := repos.Repo{
			Name:        "git-codecommit.us-west-1.amazonaws.com/stripe-go",
			URI:         "git-codecommit.us-west-1.amazonaws.com/stripe-go",
			Description: "The description",
			Language:    "barlang",
			CreatedAt:   now,
			ExternalRepo: api.ExternalRepoSpec{
				ID:          "f001337a-3450-46fd-b7d2-650c0EXAMPLE",
				ServiceType: extsvc.TypeAWSCodeCommit,
				ServiceID:   "arn:aws:codecommit:us-west-1:999999999999:",
			},
			Sources: map[string]*repos.SourceInfo{
				"extsvc:4": {
					ID:       "extsvc:4",
					CloneURL: "git@git-codecommit.us-west-1.amazonaws.com/v1/repos/stripe-go",
				},
			},
			Metadata: new(awscodecommit.Repository),
		}

		otherRepo := repos.Repo{
			Name: "git-host.com/org/foo",
			URI:  "git-host.com/org/foo",
			ExternalRepo: api.ExternalRepoSpec{
				ID:          "git-host.com/org/foo",
				ServiceID:   "https://git-host.com/",
				ServiceType: extsvc.TypeOther,
			},
			Sources: map[string]*repos.SourceInfo{
				"extsvc:5": {
					ID:       "extsvc:5",
					CloneURL: "https://git-host.com/org/foo",
				},
			},
		}

		gitoliteRepo := repos.Repo{
			Name:      "gitolite.mycorp.com/bar",
			URI:       "gitolite.mycorp.com/bar",
			CreatedAt: now,
			ExternalRepo: api.ExternalRepoSpec{
				ID:          "bar",
				ServiceType: extsvc.TypeGitolite,
				ServiceID:   "git@gitolite.mycorp.com",
			},
			Sources: map[string]*repos.SourceInfo{
				"extsvc:5": {
					ID:       "extsvc:5",
					CloneURL: "git@gitolite.mycorp.com:bar.git",
				},
			},
			Metadata: new(gitolite.Repo),
		}

		repositories := repos.Repos{
			&github,
			&gitlab,
			&bitbucketServer,
			&awsCodeCommit,
			&otherRepo,
			&gitoliteRepo,
		}

		ctx := context.Background()

		t.Run("no repos", func(t *testing.T) {
			if err := store.UpsertRepos(ctx); err != nil {
				t.Fatalf("UpsertRepos error: %s", err)
			}
		})

		t.Run("many repos", transact(ctx, store, func(t testing.TB, tx repos.Store) {
			want := mkRepos(7, repositories...)

			if err := tx.UpsertRepos(ctx, want...); err != nil {
				t.Fatalf("UpsertRepos error: %s", err)
			}

			sort.Sort(want)

			if noID := want.Filter(hasNoID); len(noID) > 0 {
				t.Fatalf("UpsertRepos didn't assign an ID to all repos: %v", noID.Names())
			}

			have, err := tx.ListRepos(ctx, repos.StoreListReposArgs{
				Kinds: kinds,
			})
			if err != nil {
				t.Fatalf("ListRepos error: %s", err)
			}

			if diff := cmp.Diff(have, []*repos.Repo(want), cmpopts.EquateEmpty()); diff != "" {
				t.Fatalf("ListRepos:\n%s", diff)
			}

			suffix := "-updated"
			now := clock.Now()
			for _, r := range want {
				r.Name += suffix
				r.URI += suffix
				r.Description += suffix
				r.Language += suffix
				r.UpdatedAt = now
				r.CreatedAt = now
				r.Archived = !r.Archived
				r.Fork = !r.Fork
			}

			if err = tx.UpsertRepos(ctx, want.Clone()...); err != nil {
				t.Errorf("UpsertRepos error: %s", err)
			} else if have, err = tx.ListRepos(ctx, repos.StoreListReposArgs{}); err != nil {
				t.Errorf("ListRepos error: %s", err)
			} else if diff := cmp.Diff(have, []*repos.Repo(want), cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("ListRepos:\n%s", diff)
			}

			deleted := want.Clone().With(repos.Opt.RepoDeletedAt(now))
			args := repos.StoreListReposArgs{}

			if err = tx.UpsertRepos(ctx, deleted...); err != nil {
				t.Fatalf("UpsertRepos error: %s", err)
			} else if have, err = tx.ListRepos(ctx, args); err != nil {
				t.Errorf("ListRepos error: %s", err)
			} else if diff := cmp.Diff(have, []*repos.Repo(nil), cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("ListRepos:\n%s", diff)
			}

			// Insert previously soft-deleted repos. Ensure we get back the same ID.
			if err = tx.UpsertRepos(ctx, want.Clone().With(repos.Opt.RepoID(0))...); err != nil {
				t.Errorf("UpsertRepos error: %s", err)
			} else if have, err = tx.ListRepos(ctx, repos.StoreListReposArgs{}); err != nil {
				t.Errorf("ListRepos error: %s", err)
			} else if diff := cmp.Diff(have, []*repos.Repo(want), cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("ListRepos:\n%s", diff)
			}

			// Delete all again, then try insert repos with different external
			// IDs but same name. Check we get new IDs.
			for _, r := range want {
				r.ID = 0
				r.ExternalRepo.ID += "-different"
			}
			if err = tx.UpsertRepos(ctx, deleted...); err != nil {
				t.Fatalf("UpsertRepos deleted error: %s", err)
			} else if err = tx.UpsertRepos(ctx, want...); err != nil {
				t.Fatalf("UpsertRepos want error: %s", err)
			} else if have, err = tx.ListRepos(ctx, repos.StoreListReposArgs{}); err != nil {
				t.Errorf("ListRepos error: %s", err)
			} else if diff := cmp.Diff(have, []*repos.Repo(want), cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("ListRepos:\n%s", diff)
			} else if sameIDs := want.Filter(hasID(deleted.IDs()...)); len(sameIDs) > 0 {
				t.Errorf("ListRepos returned IDs of soft deleted repos: %v", sameIDs.Names())
			}
		}))

		t.Run("many repos soft-deleted and single repo reinserted", transact(ctx, store, func(t testing.TB, tx repos.Store) {
			all := mkRepos(7, repositories...)

			if err := tx.UpsertRepos(ctx, all...); err != nil {
				t.Fatalf("UpsertRepos error: %s", err)
			}

			sort.Sort(all)

			if noID := all.Filter(hasNoID); len(noID) > 0 {
				t.Fatalf("UpsertRepos didn't assign an ID to all repos: %v", noID.Names())
			}

			have, err := tx.ListRepos(ctx, repos.StoreListReposArgs{
				Kinds: kinds,
			})
			if err != nil {
				t.Fatalf("ListRepos error: %s", err)
			}

			if diff := cmp.Diff(have, []*repos.Repo(all), cmpopts.EquateEmpty()); diff != "" {
				t.Fatalf("ListRepos:\n%s", diff)
			}

			allDeleted := all.Clone().With(repos.Opt.RepoDeletedAt(now))
			args := repos.StoreListReposArgs{}

			if err = tx.UpsertRepos(ctx, allDeleted...); err != nil {
				t.Fatalf("UpsertRepos error: %s", err)
			} else if have, err = tx.ListRepos(ctx, args); err != nil {
				t.Errorf("ListRepos error: %s", err)
			} else if diff := cmp.Diff(have, []*repos.Repo(nil), cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("ListRepos:\n%s", diff)
			}

			// Insert one of the previously soft-deleted repos. Ensure ID on upserted repo is set and we get back the same ID.
			want := repos.Repos{all[0]}
			upsert := want.Clone().With(repos.Opt.RepoID(0))
			if err = tx.UpsertRepos(ctx, upsert...); err != nil {
				t.Fatalf("UpsertRepos error: %s", err)
			}
			if upsert[0].ID == 0 {
				t.Fatalf("Repo ID is zero")
			}

			if have, err = tx.ListRepos(ctx, repos.StoreListReposArgs{}); err != nil {
				t.Fatalf("ListRepos error: %s", err)
			}
			if diff := cmp.Diff(have, []*repos.Repo(want), cmpopts.EquateEmpty()); diff != "" {
				t.Fatalf("ListRepos:\n%s", diff)
			}
		}))

		t.Run("it shouldn't modify the cloned column", transact(ctx, store, func(t testing.TB, tx repos.Store) {
			// UpsertRepos shouldn't set the cloned column to true
			r := mkRepos(1, repositories...)[0]
			r.Cloned = true
			if err := tx.UpsertRepos(ctx, r); err != nil {
				t.Fatalf("UpsertRepos error: %s", err)
			}

			count, err := tx.CountNotClonedRepos(ctx)
			if err != nil {
				t.Fatalf("CountNotClonedRepos error: %s", err)
			}
			if count != 1 {
				t.Fatalf("Wrong number of not cloned repos: %d", count)
			}

			// UpsertRepos shouldn't set the cloned column to false either
			if err := tx.SetClonedRepos(ctx, r.Name); err != nil {
				t.Fatalf("SetClonedRepos error: %s", err)
			}
			r = r.Clone()
			r.Cloned = false
			if err := tx.UpsertRepos(ctx, r); err != nil {
				t.Fatalf("UpsertRepos error: %s", err)
			}

			count, err = tx.CountNotClonedRepos(ctx)
			if err != nil {
				t.Fatalf("CountNotClonedRepos error: %s", err)
			}
			if count != 0 {
				t.Fatalf("Wrong number of not cloned repos: %d", count)
			}
		}))
	}
}

func isCloned(r *repos.Repo) bool {
	return r.Cloned
}

func testStoreSetClonedRepos(store repos.Store) func(*testing.T) {
	return func(t *testing.T) {
		t.Helper()

		var repositories repos.Repos
		for i := 0; i < 3; i++ {
			repositories = append(repositories, &repos.Repo{
				Name:   fmt.Sprintf("github.com/%d/%d", i, i),
				URI:    fmt.Sprintf("github.com/%d/%d", i, i),
				Cloned: false,
				ExternalRepo: api.ExternalRepoSpec{
					ID:          fmt.Sprintf("%d", i),
					ServiceType: extsvc.TypeGitHub,
					ServiceID:   "http://github.com",
				},
				Sources: map[string]*repos.SourceInfo{
					"extsvc:3": {
						ID:       "extsvc:3",
						CloneURL: "git@github.com:foo/bar.git",
					},
				},
				Metadata: new(github.Repository),
			})
		}

		check := func(t testing.TB, ctx context.Context, tx repos.Store, wantNames []string) {
			t.Helper()

			res, err := tx.ListRepos(ctx, repos.StoreListReposArgs{})
			if err != nil {
				t.Fatalf("ListRepos error: %s", err)
			}

			cloned := repos.Repos(res).Filter(isCloned).Names()
			sort.Strings(cloned)

			if got, want := cloned, wantNames; !cmp.Equal(got, want) {
				t.Fatalf("got=%v, want=%v: %s", got, want, cmp.Diff(got, want))
			}
		}

		ctx := context.Background()

		t.Run("no repo name", func(t *testing.T) {
			if err := store.SetClonedRepos(ctx); err != nil {
				t.Fatalf("SetClonedRepos error: %s", err)
			}
		})

		t.Run("many repo names", transact(ctx, store, func(t testing.TB, tx repos.Store) {
			stored := mkRepos(9, repositories...)

			if err := tx.UpsertRepos(ctx, stored...); err != nil {
				t.Fatalf("UpsertRepos error: %s", err)
			}

			sort.Sort(stored)

			names := stored[:3].Names()
			sort.Strings(names)

			if err := tx.SetClonedRepos(ctx, names...); err != nil {
				t.Fatalf("SetClonedRepos error: %s", err)
			}
			check(t, ctx, tx, names)

			// setClonedRepositories should be idempotent and have the same behavior
			// when called with the same repos
			if err := tx.SetClonedRepos(ctx, names...); err != nil {
				t.Fatalf("SetClonedRepos error: %s", err)
			}
			check(t, ctx, tx, names)

			// when adding another repo to the list, the other repos must be set as well
			names = stored[:4].Names()
			sort.Strings(names)
			if err := tx.SetClonedRepos(ctx, names...); err != nil {
				t.Fatalf("SetClonedRepos error: %s", err)
			}

			check(t, ctx, tx, names)
		}))

		t.Run("repo names in mixed case", transact(ctx, store, func(t testing.TB, tx repos.Store) {
			stored := mkRepos(9, repositories...)
			for i := range stored {
				if i%2 == 0 {
					stored[i].Name = strings.ToUpper(stored[i].Name)
				}
			}

			if err := tx.UpsertRepos(ctx, stored...); err != nil {
				t.Fatalf("UpsertRepos error: %s", err)
			}

			sort.Sort(stored)

			originalNames := stored.Names()
			sort.Strings(originalNames)

			lowerCaseNames := make([]string, 0, len(originalNames))
			for _, n := range originalNames {
				lowerCaseNames = append(lowerCaseNames, strings.ToLower(n))
			}

			if err := tx.SetClonedRepos(ctx, lowerCaseNames...); err != nil {
				t.Fatalf("SetClonedRepos error: %s", err)
			}

			check(t, ctx, tx, originalNames)
		}))
	}
}

func testStoreCountNotClonedRepos(store repos.Store) func(*testing.T) {
	return func(t *testing.T) {
		t.Helper()

		var repositories repos.Repos
		for i := 0; i < 3; i++ {
			repositories = append(repositories, &repos.Repo{
				Name:   fmt.Sprintf("github.com/%d/%d", i, i),
				URI:    fmt.Sprintf("github.com/%d/%d", i, i),
				Cloned: false,
				ExternalRepo: api.ExternalRepoSpec{
					ID:          fmt.Sprintf("%d", i),
					ServiceType: extsvc.TypeGitHub,
					ServiceID:   "http://github.com",
				},
				Sources: map[string]*repos.SourceInfo{
					"extsvc:3": {
						ID:       "extsvc:3",
						CloneURL: "git@github.com:foo/bar.git",
					},
				},
				Metadata: new(github.Repository),
			})
		}

		ctx := context.Background()

		t.Run("no cloned repos", func(t *testing.T) {
			count, err := store.CountNotClonedRepos(ctx)
			if err != nil {
				t.Fatalf("CountNotClonedRepos error: %s", err)
			}
			if diff := cmp.Diff(count, uint64(0)); diff != "" {
				t.Fatalf("CountNotClonedRepos:\n%s", diff)
			}
		})

		t.Run("multiple cloned repos", transact(ctx, store, func(t testing.TB, tx repos.Store) {
			stored := mkRepos(10, repositories...)

			if err := tx.UpsertRepos(ctx, stored...); err != nil {
				t.Fatalf("UpsertRepos error: %s", err)
			}

			sort.Sort(stored)
			cloned := stored[:3].Names()

			if err := tx.SetClonedRepos(ctx, cloned...); err != nil {
				t.Fatalf("SetClonedRepos error: %s", err)
			}

			sort.Strings(cloned)

			count, err := tx.CountNotClonedRepos(ctx)
			if err != nil {
				t.Fatalf("CountNotClonedRepos error: %s", err)
			}
			if diff := cmp.Diff(count, uint64(7)); diff != "" {
				t.Fatalf("CountNotClonedRepos:\n%s", diff)
			}
		}))

		t.Run("deleted non cloned repos", transact(ctx, store, func(t testing.TB, tx repos.Store) {
			stored := mkRepos(10, repositories...)

			if err := tx.UpsertRepos(ctx, stored...); err != nil {
				t.Fatalf("UpsertRepos error: %s", err)
			}

			sort.Sort(stored)
			cloned := stored[:3].Names()

			if err := tx.SetClonedRepos(ctx, cloned...); err != nil {
				t.Fatalf("SetClonedRepos error: %s", err)
			}

			sort.Strings(cloned)
			deletedCloned := stored[8:].With(func(r *repos.Repo) {
				r.DeletedAt = time.Now()
			})

			if err := tx.UpsertRepos(ctx, deletedCloned...); err != nil {
				t.Fatalf("UpsertRepos error: %s", err)
			}

			count, err := tx.CountNotClonedRepos(ctx)
			if err != nil {
				t.Fatalf("CountNotClonedRepos error: %s", err)
			}
			if diff := cmp.Diff(count, uint64(5)); diff != "" {
				t.Fatalf("CountNotClonedRepos:\n%s", diff)
			}
		}))
	}
}

func hasNoID(r *repos.Repo) bool {
	return r.ID == 0
}

func hasID(ids ...api.RepoID) func(r *repos.Repo) bool {
	return func(r *repos.Repo) bool {
		for _, id := range ids {
			if r.ID == id {
				return true
			}
		}
		return false
	}
}

func testStoreListRepos(store repos.Store) func(*testing.T) {
	clock := repos.NewFakeClock(time.Now(), 0)
	now := clock.Now()

	unmanaged := repos.Repo{
		Name:     "unmanaged",
		Sources:  map[string]*repos.SourceInfo{},
		Metadata: new(github.Repository),
		ExternalRepo: api.ExternalRepoSpec{
			ServiceType: "non_existent_kind",
			ServiceID:   "https://example.com/",
			ID:          "unmanaged",
		},
	}

	github := repos.Repo{
		Name: "github.com/bar/foo",
		Sources: map[string]*repos.SourceInfo{
			"extsvc:123": {
				ID:       "extsvc:123",
				CloneURL: "git@github.com:bar/foo.git",
			},
		},
		Metadata: new(github.Repository),
		ExternalRepo: api.ExternalRepoSpec{
			ServiceType: "github",
			ServiceID:   "https://github.com/",
			ID:          "foo",
		},
	}

	gitlab := repos.Repo{
		Name:    "gitlab.com/bar/foo",
		Private: true,
		Sources: map[string]*repos.SourceInfo{
			"extsvc:123": {
				ID:       "extsvc:123",
				CloneURL: "git@gitlab.com:bar/foo.git",
			},
		},
		Metadata: new(gitlab.Project),
		ExternalRepo: api.ExternalRepoSpec{
			ServiceType: extsvc.TypeGitLab,
			ServiceID:   "https://gitlab.com/",
			ID:          "123",
		},
	}

	bitbucketServer := repos.Repo{
		Name: "bitbucketserver.mycorp.com/foo/bar",
		Sources: map[string]*repos.SourceInfo{
			"extsvc:123": {
				ID:       "extsvc:123",
				CloneURL: "git@bitbucketserver.mycorp.com:foo/bar.git",
			},
		},
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "1234",
			ServiceType: "bitbucketServer",
			ServiceID:   "http://bitbucketserver.mycorp.com",
		},
		Metadata: new(bitbucketserver.Repo),
	}

	awsCodeCommit := repos.Repo{
		Name: "git-codecommit.us-west-1.amazonaws.com/stripe-go",
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "f001337a-3450-46fd-b7d2-650c0EXAMPLE",
			ServiceType: extsvc.TypeAWSCodeCommit,
			ServiceID:   "arn:aws:codecommit:us-west-1:999999999999:",
		},
		Sources: map[string]*repos.SourceInfo{
			"extsvc:4": {
				ID:       "extsvc:4",
				CloneURL: "git@git-codecommit.us-west-1.amazonaws.com/v1/repos/stripe-go",
			},
		},
		Metadata: new(awscodecommit.Repository),
	}

	otherRepo := repos.Repo{
		Name: "git-host.com/org/foo",
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "git-host.com/org/foo",
			ServiceID:   "https://git-host.com/",
			ServiceType: extsvc.TypeOther,
		},
		Sources: map[string]*repos.SourceInfo{
			"extsvc:4": {
				ID:       "extsvc:4",
				CloneURL: "https://git-host.com/org/foo",
			},
		},
	}

	gitoliteRepo := repos.Repo{
		Name:      "gitolite.mycorp.com/bar",
		CreatedAt: now,
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "bar",
			ServiceType: extsvc.TypeGitolite,
			ServiceID:   "git@gitolite.mycorp.com",
		},
		Sources: map[string]*repos.SourceInfo{
			"extsvc:5": {
				ID:       "extsvc:5",
				CloneURL: "git@gitolite.mycorp.com:bar.git",
			},
		},
		Metadata: new(gitolite.Repo),
	}

	repositories := repos.Repos{
		&github,
		&gitlab,
		&bitbucketServer,
		&awsCodeCommit,
		&otherRepo,
		&gitoliteRepo,
	}

	kinds := []string{
		extsvc.KindGitHub,
		extsvc.KindGitLab,
		extsvc.KindBitbucketServer,
		extsvc.KindAWSCodeCommit,
		extsvc.KindOther,
		extsvc.KindGitolite,
	}

	type testCase struct {
		name   string
		args   func(stored repos.Repos) repos.StoreListReposArgs
		stored repos.Repos
		repos  repos.ReposAssertion
		err    error
	}

	var testCases []testCase
	{
		stored := repositories.With(func(r *repos.Repo) {
			r.ExternalRepo.ServiceType =
				strings.ToUpper(r.ExternalRepo.ServiceType)
		})

		testCases = append(testCases, testCase{
			name: "case-insensitive kinds",
			args: func(_ repos.Repos) (args repos.StoreListReposArgs) {
				for _, kind := range kinds {
					args.Kinds = append(args.Kinds, strings.ToUpper(kind))
				}
				return args
			},
			stored: stored,
			repos:  repos.Assert.ReposEqual(stored...),
		})
	}

	testCases = append(testCases, testCase{
		name: "ignores unmanaged",
		args: func(_ repos.Repos) repos.StoreListReposArgs {
			return repos.StoreListReposArgs{Kinds: kinds}
		},
		stored: repos.Repos{&github, &gitlab, &unmanaged}.Clone(),
		repos:  repos.Assert.ReposEqual(&github, &gitlab),
	})

	{
		stored := repositories.With(repos.Opt.RepoDeletedAt(now))
		testCases = append(testCases, testCase{
			name:   "excludes soft deleted repos by default",
			stored: stored,
			repos:  repos.Assert.ReposEqual(),
		})
	}

	testCases = append(testCases, testCase{
		name:   "returns repos in ascending order by id",
		stored: mkRepos(7, repositories...),
		repos: repos.Assert.ReposOrderedBy(func(a, b *repos.Repo) bool {
			return a.ID < b.ID
		}),
	})

	testCases = append(testCases, testCase{
		name:   "returns repos by their names",
		stored: repositories,
		args: func(_ repos.Repos) repos.StoreListReposArgs {
			return repos.StoreListReposArgs{
				Names: []string{github.Name, gitlab.Name},
			}
		},
		repos: repos.Assert.ReposEqual(&github, &gitlab),
	})

	testCases = append(testCases, testCase{
		name:   "returns repos by their ids",
		stored: repositories,
		args: func(stored repos.Repos) repos.StoreListReposArgs {
			return repos.StoreListReposArgs{
				IDs: []api.RepoID{stored[0].ID, stored[1].ID},
			}
		},
		repos: repos.Assert.ReposEqual(repositories[:2].Clone()...),
	})

	testCases = append(testCases, testCase{
		name:   "limits repos to the given kinds",
		stored: repositories,
		args: func(repos.Repos) repos.StoreListReposArgs {
			return repos.StoreListReposArgs{
				Kinds: []string{extsvc.KindGitHub, extsvc.KindGitLab},
			}
		},
		repos: repos.Assert.ReposEqual(&github, &gitlab),
	})

	testCases = append(testCases, testCase{
		name: "only include private",
		args: func(repos.Repos) repos.StoreListReposArgs {
			return repos.StoreListReposArgs{
				PrivateOnly: true,
			}
		},
		stored: repositories,
		repos:  repos.Assert.ReposEqual(&gitlab),
	})

	testCases = append(testCases, testCase{
		name:   "use or",
		stored: repositories,
		args: func(repos.Repos) repos.StoreListReposArgs {
			return repos.StoreListReposArgs{
				Names: []string{"gitlab.com/bar/foo"},
				Kinds: []string{"github"},
				UseOr: true,
			}
		},
		repos: repos.Assert.ReposEqual(&github, &gitlab),
	})

	testCases = append(testCases, testCase{
		name:   "use and",
		stored: repositories,
		args: func(repos.Repos) repos.StoreListReposArgs {
			return repos.StoreListReposArgs{
				Names: []string{"gitlab.com/bar/foo"},
				Kinds: []string{"github"},
				UseOr: false,
			}
		},
		repos: repos.Assert.ReposEqual(),
	})

	return func(t *testing.T) {
		t.Helper()

		ctx := context.Background()

		for _, tc := range testCases {
			tc := tc

			t.Run(tc.name, transact(ctx, store, func(t testing.TB, tx repos.Store) {
				stored := tc.stored.Clone()
				if err := tx.UpsertRepos(ctx, stored...); err != nil {
					t.Fatalf("failed to setup store: %v", err)
				}

				var args repos.StoreListReposArgs
				if tc.args != nil {
					args = tc.args(stored)
				}

				rs, err := tx.ListRepos(ctx, args)
				if have, want := fmt.Sprint(err), fmt.Sprint(tc.err); have != want {
					t.Errorf("error:\nhave: %v\nwant: %v", have, want)
				}

				if tc.repos != nil {
					tc.repos(t, rs)
				}
			}))
		}

		t.Run("only include cloned", transact(ctx, store, func(t testing.TB, tx repos.Store) {
			stored := mkRepos(5, repositories...).Clone()
			if err := tx.UpsertRepos(ctx, stored...); err != nil {
				t.Fatalf("failed to setup store: %v", err)
			}

			sort.Sort(stored)

			cloned := stored[:3]
			if err := tx.SetClonedRepos(ctx, cloned.Names()...); err != nil {
				t.Fatalf("failed to set cloned repos: %v", err)
			}

			args := repos.StoreListReposArgs{
				ClonedOnly: true,
			}

			rs, err := tx.ListRepos(ctx, args)
			if err != nil {
				t.Errorf("failed to list repos: %v", err)
			}

			want := cloned.With(func(r *repos.Repo) {
				r.Cloned = true
			})

			repos.Assert.ReposEqual(want...)(t, rs)
		}))
	}
}

func testStoreListReposPagination(store repos.Store) func(*testing.T) {
	clock := repos.NewFakeClock(time.Now(), 0)
	now := clock.Now()

	github := repos.Repo{
		Name:        "foo/bar",
		URI:         "github.com/foo/bar",
		Description: "The description",
		Language:    "barlang",
		CreatedAt:   now,
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "AAAAA==",
			ServiceType: "github",
			ServiceID:   "http://github.com",
		},
		Sources: map[string]*repos.SourceInfo{
			"extsvc:1": {
				ID:       "extsvc:1",
				CloneURL: "git@github.com:foo/bar.git",
			},
		},
		Metadata: new(github.Repository),
	}

	return func(t *testing.T) {
		ctx := context.Background()
		t.Run("", transact(ctx, store, func(t testing.TB, tx repos.Store) {
			stored := mkRepos(7, &github)
			if err := tx.UpsertRepos(ctx, stored...); err != nil {
				t.Fatalf("UpsertRepos error: %s", err)
			}

			sort.Sort(stored)

			lo, hi := -2, len(stored)+2
			for page := lo; page < hi; page++ {
				for limit := lo; limit < hi; limit++ {
					args := repos.StoreListReposArgs{
						PerPage: int64(page),
						Limit:   int64(limit),
					}

					listed, err := tx.ListRepos(ctx, args)
					if err != nil {
						t.Fatalf("unexpected error with page=%d, limit=%d: %v", page, limit, err)
					}

					var want repos.Repos
					if limit <= 0 || limit >= len(stored) {
						want = stored
					} else {
						want = stored[:limit]
					}

					if have := repos.Repos(listed); !reflect.DeepEqual(have, want) {
						t.Fatalf("page=%d, limit=%d: %s", page, limit, cmp.Diff(have, want))
					}
				}
			}
		}))
	}
}

func testDBStoreTransact(store *repos.DBStore) func(*testing.T) {
	return func(t *testing.T) {
		ctx := context.Background()

		txstore, err := store.Transact(ctx)
		if err != nil {
			t.Fatal("expected DBStore to support transactions", err)
		}
		defer txstore.Done()

		_, err = txstore.(repos.Transactor).Transact(ctx)
		have := fmt.Sprintf("%s", err)
		want := "dbstore: already in a transaction"
		if have != want {
			t.Errorf("error:\nhave: %v\nwant: %v", have, want)
		}
	}
}

func mkRepos(n int, base ...*repos.Repo) repos.Repos {
	if len(base) == 0 {
		return nil
	}

	rs := make(repos.Repos, 0, n)
	for i := 0; i < n; i++ {
		id := strconv.Itoa(i)
		r := base[i%len(base)].Clone()
		r.Name += id
		r.ExternalRepo.ID += id
		rs = append(rs, r)
	}
	return rs
}

func mkExternalServices(n int, base ...*repos.ExternalService) repos.ExternalServices {
	if len(base) == 0 {
		return nil
	}
	es := make(repos.ExternalServices, 0, n)
	for i := 0; i < n; i++ {
		id := strconv.Itoa(i)
		r := base[i%len(base)].Clone()
		r.DisplayName += id
		es = append(es, r)
	}
	return es
}

func transact(ctx context.Context, s repos.Store, test func(testing.TB, repos.Store)) func(*testing.T) {
	return func(t *testing.T) {
		t.Helper()

		tr, ok := s.(repos.Transactor)

		if ok {
			txstore, err := tr.Transact(ctx)
			if err != nil {
				t.Fatalf("failed to start transaction: %v", err)
			}
			defer txstore.Done(&errRollback)
			s = &noopTxStore{TB: t, Store: txstore}
		}

		test(t, s)
	}
}

type noopTxStore struct {
	testing.TB
	repos.Store
	count int
}

func (tx *noopTxStore) Transact(context.Context) (repos.TxStore, error) {
	if tx.count != 0 {
		return nil, fmt.Errorf("noopTxStore: %d current transactions", tx.count)
	}
	tx.count++
	// noop
	return tx, nil
}

func (tx *noopTxStore) Done(errs ...*error) {
	tx.Helper()

	if tx.count != 1 {
		tx.Fatal("no current transactions")
	}
	if len(errs) > 0 && *errs[0] != nil {
		tx.Fatal(fmt.Sprintf("unexpected error in noopTxStore: %v", *errs[0]))
	}
	tx.count--
}
