package repos_test

import (
	"context"
	"encoding/json"
	"fmt"
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
				"access_token":"9cc46dcda66306277915a6919a90ac7972853317d9df385a828b17d9200b7d4c",
				"token_type":"Bearer",
				"refresh_token":"5fa56e21251f4c2295494ee29b6b66f7011dad92251ab988a376a23ef12ad041",
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

		// extAccounts := database.NewMockUserExternalAccountsStore()
		// extAccounts.ListFunc.SetDefaultReturn([]*extsvc.Account{&extAccount}, nil)

		// userAccountsDB := database.NewMockDB()
		// userAccountsDB.UserExternalAccountsFunc.SetDefaultReturn(extAccounts)

		user, err := s.UserExternalAccountsStore().CreateUserAndSave(ctx, database.NewUser{
			Email:                 "a@a.com",
			Username:              "u",
			Password:              "p",
			EmailVerificationCode: "c",
		}, extAccount.AccountSpec, extAccount.AccountData)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Printf("User:%+v\n", user)

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

		var job string
		select {
		case job = <-jobChan:
			fmt.Println("job received:", job)
			return
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout")
		}

		// create a server
		// use the server's URL
		// send a push event to the repo
		// assert that we enqueue a repo update

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
		fmt.Println("getting accounts...")
		accts, err := h.store.UserExternalAccountsStore().List(ctx, database.ExternalAccountsListOptions{})
		if err != nil {
			return errors.Newf("Error getting user accounts,", err)
		}
		fmt.Println("Accts:", len(accts))

		_, token, err := github.GetExternalAccountData(&accts[0].AccountData)
		if err != nil {
			fmt.Println("token error:", err)
		}
		fmt.Println("token:", token.AccessToken)

		foundSyncWebhook := githubwebhook.FindSyncWebhook(wbj.RepoName, token.AccessToken)
		if !foundSyncWebhook {
			githubwebhook.CreateSyncWebhook(wbj.RepoName, "secret", token.AccessToken)
		}
		h.jobChan <- "done!"
	}

	// how will we know if a repo has been deleted?
	return nil
}

// func (h *fakeWhBuildHandler) Handle2(ctx context.Context, logger log.Logger, record workerutil.Record) error {
// 	fmt.Println("in fake handler")
// 	wbj, ok := record.(*repos.WhBuildJob)
// 	if !ok {
// 		h.jobChan <- "wrong type"
// 		return errors.Errorf("expected repos.WhBuildJob, got %T", record)
// 	}
// 	fmt.Printf("Job:%+v\n", wbj)

// 	switch wbj.ExtsvcKind {
// 	case "GITHUB":
// 		fmt.Println("getting accounts...")
// 		accts, err := h.userAccountsDB.UserExternalAccounts().List(ctx, database.ExternalAccountsListOptions{
// 			UserID:         777,
// 			ExcludeExpired: true,
// 		})
// 		if err != nil {
// 			return errors.Newf("Error getting user accounts,", err)
// 		}
// 		fmt.Println("Accts:", len(accts))

// 		_, token, err := github.GetExternalAccountData(&accts[0].AccountData)
// 		if err != nil {
// 			fmt.Println("token error:", err)
// 		}

// 		fmt.Println("token:", token.AccessToken)
// 		webhookName := githubwebhook.FindSyncWebhook(wbj.RepoName, "secret", token.AccessToken)
// 		if webhookName != "web" {
// 			err := githubwebhook.CreateSyncWebhook(string(wbj.RepoName), "secret", token.AccessToken)
// 			if err != nil {
// 				return errors.Errorf("failed to create webhook for %s", wbj.RepoName)
// 			}
// 			h.jobChan <- "created new webhook"
// 			return nil
// 		} else {
// 			h.jobChan <- "webhook: " + webhookName
// 			return nil
// 		}
// 	}

// 	// how will we know if a repo has been deleted?
// 	return nil
// }

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
