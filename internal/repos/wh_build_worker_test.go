package repos_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/log/logtest"

	ru "github.com/sourcegraph/sourcegraph/cmd/repo-updater/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/httpserver"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	githubwebhook "github.com/sourcegraph/sourcegraph/internal/repos/webhooks"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

func testSyncWebhookWorker(s repos.Store) func(*testing.T) {
	return func(t *testing.T) {
		ctx := context.Background()
		esStore := s.ExternalServiceStore()

		ghConn := &schema.GitHubConnection{
			Url:   extsvc.KindGitHub,
			Token: "token",
			Repos: []string{"susantoscott/Task-Tracker"},
		}
		bs, err := json.Marshal(ghConn)
		if err != nil {
			t.Fatal(err)
		}

		config := string(bs)
		testSvc := &types.ExternalService{
			Kind:        extsvc.KindGitHub,
			DisplayName: "testingTesting123",
			Config:      config,
		}
		if err := esStore.Upsert(ctx, testSvc); err != nil {
			t.Fatal(err)
		}

		testRepo := &types.Repo{
			ID:       33,
			Name:     "susantoscott/Task-Tracker",
			Metadata: &github.Repository{},
			ExternalRepo: api.ExternalRepoSpec{
				ID:          "hi-mom-12345",
				ServiceID:   "https://github.com/",
				ServiceType: extsvc.TypeGitHub,
			},
		}
		err = s.RepoStore().Create(ctx, testRepo)
		if err != nil {
			t.Fatal(err)
		}

		sourcer := repos.NewFakeSourcer(nil, repos.NewFakeSource(testSvc, nil, testRepo))
		syncer := &repos.Syncer{
			Logger:  logtest.Scoped(t),
			Sourcer: sourcer,
			Store:   s,
			Now:     time.Now,
		}

		if err := syncer.SyncExternalService(ctx, testSvc.ID, time.Millisecond); err != nil {
			t.Fatal(err)
		}

		data := json.RawMessage(`{}`)
		authData := json.RawMessage(fmt.Sprintf(`
			{
				"access_token":"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",
				"token_type":"Bearer",
				"refresh_token":"",
				"expiry":"%s"
			}`,
			time.Now().Add(time.Hour).Format(time.RFC3339)))
		extAccount := extsvc.Account{
			ID:     0,
			UserID: 777,
			AccountSpec: extsvc.AccountSpec{
				ServiceID:   "serviceID",
				ServiceType: "testService",
				ClientID:    "clientID",
				AccountID:   "accountID",
			},
			AccountData: extsvc.AccountData{
				AuthData: &authData,
				Data:     &data,
			},
		}

		_, err = s.UserExternalAccountsStore().CreateUserAndSave(ctx, database.NewUser{
			Email:                 "secret@usc.edu",
			Username:              "susantoscott",
			Password:              "saltedPassword",
			EmailVerificationCode: "123456",
		}, extAccount.AccountSpec, extAccount.AccountData)
		if err != nil {
			t.Fatal(err)
		}

		jobChan := make(chan string)
		whBuildHandler := &fakeWhBuildHandler{
			store:              s,
			jobChan:            jobChan,
			minWhBuildInterval: func() time.Duration { return time.Minute },
		}

		whBuildWorker, _ := repos.NewWhBuildWorker(ctx, s.Handle(), whBuildHandler, repos.WhBuildOptions{
			NumHandlers:    1,
			WorkerInterval: 1 * time.Millisecond,
		})
		go whBuildWorker.Start()
		defer whBuildWorker.Stop()

	loop:
		select {
		case <-jobChan:
			fmt.Println("finished build job")
			break loop
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout")
		}

		type event struct {
			PayloadType string
			Data        json.RawMessage
		}

		db := database.NewDB(dbtest.NewDB(t))
		logger := log.Scoped("service", "repo-updater service")

		fmt.Println("Creating server...")
		server := &ru.Server{
			Logger:                logger,
			Store:                 s,
			Scheduler:             repos.NewUpdateScheduler(logger, db),
			GitserverClient:       gitserver.NewClient(db),
			SourcegraphDotComMode: false,
			RateLimitSyncer:       repos.NewRateLimitSyncer(ratelimit.DefaultRegistry, esStore, repos.RateLimitSyncerOpts{}),
		}

		var handler http.Handler
		{
			m := ru.NewHandlerMetrics()
			m.MustRegister(prometheus.DefaultRegisterer)

			handler = ru.ObservedHandler(
				logger,
				m,
				opentracing.GlobalTracer(),
			)(server.Handler())
		}

		httpServer := httpserver.NewFromAddr("localhost:8080", &http.Server{
			ReadTimeout:  75 * time.Second,
			WriteTimeout: 10 * time.Minute,
			Handler:      handler,
		})

		go func() {
			httpServer.Start()
			defer httpServer.Stop()
		}()
		fmt.Println("Server is live...")

		repoName := "susantoscott/Task-Tracker"
		secret := "secret"
		token := "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"
		id, err := githubwebhook.CreateSyncWebhook(repoName, secret, token)
		// I still have to add it to the store
		if err != nil {
			t.Fatal("create:", err)
		}

		success, err := githubwebhook.TestPushSyncWebhook(repoName, id, token)
		if err != nil {
			t.Fatal(err)
		}
		if !success {
			t.Fatal("failed")
		}

		deleted, err := githubwebhook.DeleteSyncWebhook(repoName, id, token)
		if err != nil {
			t.Fatal(err)
		}
		if !deleted {
			t.Fatal("not deleted")
		}

		time.Sleep(5 * time.Second)
	}
}

type fakeWhBuildHandler struct {
	store              repos.Store
	jobChan            chan string
	minWhBuildInterval func() time.Duration
}

func (h *fakeWhBuildHandler) Handle(ctx context.Context, logger log.Logger, record workerutil.Record) error {
	fmt.Println("in fake handler")
	wbj, ok := record.(*repos.WhBuildJob)
	if !ok {
		h.jobChan <- "wrong type"
		return errors.Errorf("expected repos.WhBuildJob, got %T", record)
	}
	fmt.Printf("Job:%+v\n", wbj)

	switch wbj.ExtsvcKind {
	case "GITHUB":
		accts, err := h.store.UserExternalAccountsStore().List(ctx, database.ExternalAccountsListOptions{})
		if err != nil {
			return errors.Newf("Error getting user accounts,", err)
		}

		_, token, err := github.GetExternalAccountData(&accts[0].AccountData)
		if err != nil {
			fmt.Println("token error:", err)
		}

		foundSyncWebhook := githubwebhook.FindSyncWebhook(wbj.RepoName, token.AccessToken)
		if !foundSyncWebhook {
			fmt.Println("not found")
			githubwebhook.CreateSyncWebhook(wbj.RepoName, "secret", token.AccessToken)
		}
		h.jobChan <- "done!"
	}

	// how will we know if a repo has been deleted?
	return nil
}

func PrintRows(s repos.Store, ctx context.Context) {
	fmt.Println("Printing rows...")
	q := sqlf.Sprintf(`select * from webhook_build_jobs;`)
	rows, err := s.RepoStore().Handle().QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		fmt.Println("error printing rows")
	}
	var jobs []repos.WhBuildJob

	for rows.Next() {
		var job repos.WhBuildJob
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

	fmt.Println("Len:", len(jobs))
	for _, j := range jobs {
		fmt.Printf("Job:%+v\n", j)
	}
}

func sign(t *testing.T, message, secret []byte) string {
	t.Helper()

	mac := hmac.New(sha256.New, secret)

	_, err := mac.Write(message)
	if err != nil {
		t.Fatalf("writing hmac message failed: %s", err)
	}

	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}
