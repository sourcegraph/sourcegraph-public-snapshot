package repos_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/types/typestest"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func testSyncWebhookWorker(db database.DB) func(t *testing.T) {
	return func(t *testing.T) {
		ctx := context.Background()
		testRepo := &types.Repo{
			ID:       33,
			Name:     "github.com/susantoscott/Task-Tracker",
			Metadata: &github.Repository{},
			ExternalRepo: api.ExternalRepoSpec{
				ID:          "hi-mom-12345",
				ServiceID:   "https://github.com/",
				ServiceType: extsvc.TypeGitHub,
			},
		}
		err := db.Repos().Create(ctx, testRepo)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Printf("testRepo:%+v\n", testRepo)

		q1 := sqlf.Sprintf(`insert into create_webhook_jobs (repo_id, repo_name) values (%d, %s);`, testRepo.ID, testRepo.Name)
		result, err := db.Handle().ExecContext(ctx, q1.Query(sqlf.PostgresBindVar), q1.Args()...)
		if err != nil {
			t.Fatal(err)
		}
		PrintRows(db, ctx)
		rowsAffected, err := result.RowsAffected()
		if rowsAffected != 1 {
			t.Fatalf("Expected 1 row to be affected, got %d", rowsAffected)
		}

		jobChan := make(chan *repos.CreateWebhookJob)

		h := &fakeWebhookCreationHandler{
			jobChan: jobChan,
		}
		worker, _ := repos.NewWebhookCreatingWorker(ctx, db.Handle(), h, repos.WebhookCreatingWorkerOpts{
			NumHandlers:    1,
			WorkerInterval: 1 * time.Millisecond,
		})
		go worker.Start()
		// go resetter.Start()

		defer worker.Stop()
		// defer resetter.Stop()

		var job *repos.CreateWebhookJob
		select {
		case job = <-jobChan:
			fmt.Println("Job received")
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout")
		}
		fmt.Printf("job:%+v\n", job)

	}
}

func PrintRows(db database.DB, ctx context.Context) {
	q := sqlf.Sprintf(`select * from create_webhook_jobs;`)
	rows, err := db.Handle().QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		fmt.Println("error printing rows")
	}
	var jobs []repos.CreateWebhookJob

	for rows.Next() {
		var job repos.CreateWebhookJob
		var executionLogs *[]any
		rows.Scan(
			&job.ID,
			&job.State,
			&job.FailureMessage,
			&job.StartedAt,
			&job.FinishedAt,
			&job.ProcessAfter,
			&job.NumResets,
			&job.NumFailures,
			&executionLogs,
			&job.RepoID,
			&job.RepoName,
			&job.QueuedAt,
		)
		jobs = append(jobs, job)
	}
	if err := rows.Err(); err != nil {
		fmt.Println("err printing:", err)
	}
}

func testOldSyncWebhookWorker(s repos.Store) func(*testing.T) {
	return func(t *testing.T) {
		servicesPerKind := createExternalServices(t, s)
		githubService := servicesPerKind[extsvc.KindGitHub]
		githubRepo1 := (*&types.Repo{
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
		githubRepo2 := (*&types.Repo{
			Name:     "github.com/susantoscott/Password-Cracking",
			Metadata: &github.Repository{},
			ExternalRepo: api.ExternalRepoSpec{
				ID:          "hi-mom-6789",
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
			{repo: githubRepo1, svc: githubService},
			{repo: githubRepo2, svc: githubService},
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
			fmt.Println("Testing:", tc.name)
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

				parts := strings.Split(tc.name, "/")
				reponame := parts[2]

				befores := repos.ListWebhooks(reponame)
				fmt.Println("Before:")
				if len(befores) == 0 {
					fmt.Println("[]")
				}
				for _, before := range befores {
					fmt.Printf("%+v\n", before)
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
					// fmt.Printf("svc:%+v\n", svc)
					before, err := st.ExternalServiceStore().GetByID(ctx, svc.ID)
					if err != nil {
						t.Fatal(err)
					}
					// fmt.Printf("before:%+v\n", before)

					err = syncer.SyncExternalService(ctx, svc.ID, time.Millisecond)
					if have, want := fmt.Sprint(err), tc.err; !strings.Contains(have, want) {
						t.Errorf("erroq %q doesn't contain %q", have, want)
					}

					after, err := st.ExternalServiceStore().GetByID(ctx, svc.ID)
					if err != nil {
						t.Fatal(err)
					}
					// fmt.Printf("after:%+v\n", after)

					if before.LastSyncAt == after.LastSyncAt {
						t.Log(before.LastSyncAt, after.LastSyncAt)
						t.Errorf("Service %q last_synced was not updated", svc.DisplayName)
					}
				}

				toDeletes := make([]int, 0)

				afters := repos.ListWebhooks(reponame)
				fmt.Println("After:")
				for _, after := range afters {
					fmt.Printf("%+v\n", after)
					toDeletes = append(toDeletes, after.ID)
				}

				for _, toDelete := range toDeletes {
					repos.DeleteWebhook(reponame, toDelete)
				}

				// var want, have types.Repos
				// want.Concat(tc.diff.Added, tc.diff.Modified, tc.diff.Unmodified)
				// have, _ = st.RepoStore().List(ctx, database.ReposListOptions{})
				// want = want.With(typestest.Opt.RepoID(0))
				// have = have.With(typestest.Opt.RepoID(0))
				// sort.Sort(want)
				// sort.Sort(have)
				// fmt.Printf("want:%+v\n", want)
				// fmt.Printf("have:%+v\n", have)

				// typestest.Assert.ReposEqual(want...)(t, have)
			}))
		}
	}
}

type fakeWebhookCreationHandler struct {
	jobChan chan *repos.CreateWebhookJob
}

func (h *fakeWebhookCreationHandler) Handle(ctx context.Context, logger log.Logger, record workerutil.Record) error {
	fmt.Println("in the fake Handlerr")
	cwj, ok := record.(*repos.CreateWebhookJob)
	if !ok {
		return errors.Errorf("expected repos.CreateWebhookJob, got %T", record)
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case h.jobChan <- cwj:
		fmt.Println("putting creation job in channel")
		return repos.CreateSyncWebhook(cwj.RepoName, "secret", "token")
	}
}
