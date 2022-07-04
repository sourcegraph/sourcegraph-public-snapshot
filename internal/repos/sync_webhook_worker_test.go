package repos_test

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/types/typestest"
)

func testSyncWebhookWorker(s repos.Store) func(*testing.T) {
	return func(t *testing.T) {
		servicesPerKind := createExternalServices(t, s)
		githubService := servicesPerKind[extsvc.KindGitHub]
		githubRepo := (*&types.Repo{
			Name:     "github.com/susantoscott/Task-Tracker",
			Metadata: &github.Repository{},
			ExternalRepo: api.ExternalRepoSpec{
				ID:          "hi-mom-12345",
				ServiceID:   "https://github.com/",
				ServiceType: extsvc.TypeGitHub,
			},
		}).With(
			typestest.Opt.RepoSources(githubService.URN()),
		)

		clock := timeutil.NewFakeClock(time.Now(), 0)
		svcdup := types.ExternalService{
			Kind:        extsvc.KindGitHub,
			DisplayName: "Github - Test",
			Config:      `{"url": "https://github.com"}`,
			CreatedAt:   clock.Now(),
			UpdatedAt:   clock.Now(),
		}

		q := sqlf.Sprintf(`INSERT INTO users (id, username) VALUES (1, 'u')`)
		_, err := s.Handle().ExecContext(context.Background(), q.Query(sqlf.PostgresBindVar), q.Args()...)
		if err != nil {
			t.Fatal(err)
		}

		userAddedGithubSvc := githubService.With(func(service *types.ExternalService) {
			service.ID = 0
			service.NamespaceUserID = 1
		})

		if err := s.ExternalServiceStore().Upsert(context.Background(), &svcdup, userAddedGithubSvc); err != nil {
			t.Fatalf("failed to insert external service: %v", err)
		}

		// userAddedGithubRepo := githubRepo.With(func(r *types.Repo) {
		// 	r.Name += "-2"
		// 	r.ExternalRepo.ID += "-2"
		// },
		// 	typestest.Opt.RepoSources(userAddedGithubSvc.URN()),
		// )

		type testCase struct {
			name    string
			sourcer repos.Sourcer
			store   repos.Store
			stored  types.Repos
			svcs    []*types.ExternalService
			ctx     context.Context
			now     func() time.Time
			diff    repos.Diff
			err     string
		}

		var testCases []testCase
		for _, tc := range []struct {
			repo *types.Repo
			svc  *types.ExternalService
		}{
			{repo: githubRepo, svc: githubService},
			// {repo: userAddedGithubRepo, svc: userAddedGithubSvc},
		} {
			testCases = append(testCases,
				testCase{
					name: string(tc.repo.Name) + "/new repo",
					sourcer: repos.NewFakeSourcer(nil,
						repos.NewFakeSource(tc.svc.Clone(), nil, tc.repo.Clone()),
					),
					store:  s,
					stored: types.Repos{},
					now:    clock.Now,
					diff: repos.Diff{Added: types.Repos{tc.repo.With(
						typestest.Opt.RepoCreatedAt(clock.Time(1)),
						typestest.Opt.RepoSources(tc.svc.Clone().URN()),
					)}},
					svcs: []*types.ExternalService{tc.svc},
					err:  "<nil>",
				},
			)

			var diff repos.Diff
			if tc.svc.NamespaceUserID > 0 {
				diff.Deleted = append(diff.Deleted, tc.repo.With(
					typestest.Opt.RepoSources(tc.svc.URN()),
				))
			} else {
				diff.Unmodified = append(diff.Unmodified, tc.repo.With(
					typestest.Opt.RepoSources(tc.svc.URN()),
				))
			}
		}

		for _, tc := range testCases {
			if tc.name == "" {
				t.Error("Test case name is blank")
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
				if st == nil {
					t.Fatal("nil store")
				}
				now := tc.now
				if now == nil {
					clock := timeutil.NewFakeClock(time.Now(), time.Second)
					now = clock.Now
				}
				ctx := tc.ctx
				if ctx == nil {
					ctx = context.Background()
				}

				if len(tc.stored) > 0 {
					cloned := tc.stored.Clone()
					if err := st.RepoStore().Create(ctx, cloned...); err != nil {
						t.Fatalf("failed to prepare store: %v", err)
					}
				}

				syncer := &repos.Syncer{
					Logger:  logtest.Scoped(t),
					Sourcer: tc.sourcer,
					Store:   st,
					Now:     now,
				}

				for _, svc := range tc.svcs {
					before, err := st.ExternalServiceStore().GetByID(ctx, svc.ID)
					if err != nil {
						t.Fatal(err)
					}

					err = syncer.SyncExternalService(ctx, svc.ID, time.Millisecond)
					if have, want := fmt.Sprint(err), tc.err; !strings.Contains(have, want) {
						t.Errorf("erroq %q doesn't contain %q", have, want)
					}

					after, err := st.ExternalServiceStore().GetByID(ctx, svc.ID)
					if err != nil {
						t.Fatal(err)
					}

					if before.LastSyncAt == after.LastSyncAt {
						t.Log(before.LastSyncAt, after.LastSyncAt)
						t.Errorf("Service %q last_synced was not updated", svc.DisplayName)
					}
				}

				var want, have types.Repos
				want.Concat(tc.diff.Added, tc.diff.Modified, tc.diff.Unmodified)
				have, _ = st.RepoStore().List(ctx, database.ReposListOptions{})
				want = want.With(typestest.Opt.RepoID(0))
				have = have.With(typestest.Opt.RepoID(0))
				sort.Sort(want)
				sort.Sort(have)
				fmt.Printf("want:%+v\n", want)
				fmt.Printf("have:%+v\n", have)

				typestest.Assert.ReposEqual(want...)(t, have)
			}))
		}
	}
}
