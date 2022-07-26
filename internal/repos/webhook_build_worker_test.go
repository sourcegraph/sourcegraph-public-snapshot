package repos_test

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/thanhpk/randstr"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/webhooks"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/httpserver"
	"github.com/sourcegraph/sourcegraph/internal/httptestutil"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	ru "github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

var updateWebhooks = flag.Bool("updateWebhooks", false, "update testdata for webhook build worker integration test")

func testWebhookBuilder(store repos.Store, db database.DB) func(*testing.T) {
	return func(t *testing.T) {
		ctx := context.Background()
		logger := logtest.Scoped(t)

		err := godotenv.Load("./.env")
		if err != nil {
			t.Fatal(err)
		}
		token := os.Getenv("ACCESS_TOKEN")

		esStore := store.ExternalServiceStore()
		repoStore := store.RepoStore()

		repo := &types.Repo{
			ID:       1,
			Name:     api.RepoName("ghe.sgdev.org/milton/test"),
			Metadata: &github.Repository{},
			ExternalRepo: api.ExternalRepoSpec{
				ID:          "hi-mom-12345",
				ServiceID:   "https://ghe.sgdev.org",
				ServiceType: extsvc.TypeGitHub,
			},
		}
		if err := repoStore.Create(ctx, repo); err != nil {
			t.Fatal(err)
		}

		ghConn := &schema.GitHubConnection{
			Url:      extsvc.KindGitHub,
			Token:    token,
			Repos:    []string{string(repo.Name)},
			Webhooks: []*schema.GitHubWebhook{{Org: "ghe.sgdev.org", Secret: "secret"}},
		}
		bs, err := json.Marshal(ghConn)
		if err != nil {
			t.Fatal(err)
		}

		config := string(bs)
		svc := &types.ExternalService{
			Kind:        extsvc.KindGitHub,
			DisplayName: "TestService",
			Config:      config,
		}
		if err := esStore.Upsert(ctx, svc); err != nil {
			t.Fatal(err)
		}

		sourcer := repos.NewFakeSourcer(nil, repos.NewFakeSource(svc, nil, repo))
		syncer := &repos.Syncer{
			Logger:  logger,
			Sourcer: sourcer,
			Store:   store,
			Now:     time.Now,
		}

		conf.Get().ExperimentalFeatures.EnableWebhookSyncing = true
		if err := syncer.SyncExternalService(ctx, svc.ID, time.Millisecond); err != nil {
			t.Fatal(err)
		}

		accountData := json.RawMessage(`{}`)
		authData := json.RawMessage(fmt.Sprintf(`
			{
				"access_token":"%s",
				"token_type":"Bearer",
				"refresh_token":"",
				"expiry":"%s"
			}`,
			token, time.Now().Add(time.Hour).Format(time.RFC3339)))
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
				Data:     &accountData,
			},
		}

		if _, err := store.UserExternalAccountsStore().CreateUserAndSave(ctx, database.NewUser{
			Email:                 "USCtrojan@sourcegraph.com",
			Username:              "susantoscott",
			Password:              "saltedPassword!@#$%",
			EmailVerificationCode: "123456",
		}, extAccount.AccountSpec, extAccount.AccountData); err != nil {
			t.Fatal(err)
		}

		testName := "sync-webhook-integration"
		cf, save := httptestutil.NewGitHubRecorderFactory(t, *updateWebhooks, testName)
		defer save()

		doer, err := cf.Doer()
		if err != nil {
			t.Fatal(err)
		}

		jobChan := make(chan int)
		whBuildHandler := &fakeWhBuildHandler{
			store:              store,
			doer:               &doer,
			jobChan:            jobChan,
			minWhBuildInterval: func() time.Duration { return time.Minute },
		}

		whBuildWorker, _ := repos.NewWebhookBuildWorker(ctx, store.Handle(), whBuildHandler, repos.WebhookBuildOptions{
			NumHandlers:    3,
			WorkerInterval: 1 * time.Millisecond,
		})

		go whBuildWorker.Start()
		defer whBuildWorker.Stop()

		var id int

	loop:
		select {
		case id = <-jobChan:
			logger.Info(fmt.Sprintf("Channel received: %d", id))
			break loop
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout")
		}

		// this is in response to line 1514 in
		// internal/extsvc/github/common.go
		baseURL, err := url.Parse("")
		if err != nil {
			t.Fatal(err)
		}

		client := github.NewV3Client(logger, svc.URN(), baseURL, &auth.OAuthBearerToken{Token: token}, doer)
		g := repos.NewGitHubWebhookAPI(client)

		// Test Webhook
		// Delete webhook

		handler := webhooks.GitHubWebhook{
			ExternalServices: store.ExternalServiceStore(),
		}
		g.Register(&handler)

		payloadBytes, err := os.ReadFile(filepath.Join("testdata", "webhook-build-worker.json"))
		if err != nil {
			t.Fatal(err)
		}

		// setting up the internal server
		// that repoupdater listens to
		server := &repoupdater.Server{
			Logger:                logger,
			Store:                 store,
			Scheduler:             repos.NewUpdateScheduler(logger, db),
			GitserverClient:       gitserver.NewClient(db),
			SourcegraphDotComMode: envvar.SourcegraphDotComMode(),
			RateLimitSyncer:       repos.NewRateLimitSyncer(ratelimit.DefaultRegistry, store.ExternalServiceStore(), repos.RateLimitSyncerOpts{}),
		}

		serverURL := "127.0.0.1:8080"

		ru.DefaultClient = newClient(fmt.Sprintf("http://%s", serverURL), doer)
		httpSrv := httpserver.NewFromAddr(serverURL, &http.Server{
			ReadTimeout:  75 * time.Second,
			WriteTimeout: 10 * time.Minute,
			Handler:      server.Handler(),
		})

		go httpSrv.Start()
		defer httpSrv.Stop()

		// mock the Github PushEvent
		// e.g. user makes a commit
		url := fmt.Sprintf("%s/github-webhooks", globals.ExternalURL())
		req, err := http.NewRequest("POST", url, bytes.NewReader([]byte(payloadBytes)))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("X-Github-Event", "push")
		req.Header.Set("X-Hub-Signature", sign(t, payloadBytes, []byte("secret")))

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		resp := rec.Result()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			t.Fatalf("Non-200 status code: %v", resp.StatusCode)
		}
	}
}

type fakeWhBuildHandler struct {
	store              repos.Store
	doer               *httpcli.Doer
	jobChan            chan int
	minWhBuildInterval func() time.Duration
}

func (h *fakeWhBuildHandler) Handle(ctx context.Context, logger log.Logger, record workerutil.Record) error {
	wbj, ok := record.(*repos.WebhookBuildJob)
	if !ok {
		h.jobChan <- 0
		return errors.Errorf("expected repos.WhBuildJob, got %T", record)
	}

	switch wbj.ExtSvcKind {
	case "GITHUB":
		if err := activateGitHubHandler(ctx, logger, h, wbj); err != nil {
			return err
		}
	}

	return nil
}

func activateGitHubHandler(ctx context.Context, logger log.Logger, h *fakeWhBuildHandler, wbj *repos.WebhookBuildJob) error {
	svcs, err := h.store.ExternalServiceStore().List(ctx, database.ExternalServicesListOptions{})
	if err != nil || len(svcs) != 1 {
		return errors.Wrap(err, "get external service")
	}
	svc := svcs[0]

	baseURL, err := url.Parse("")
	if err != nil {
		return errors.Wrap(err, "parse base URL")
	}

	accts, err := h.store.UserExternalAccountsStore().List(ctx, database.ExternalAccountsListOptions{})
	if err != nil {
		return errors.Wrap(err, "get accounts")
	}

	_, token, err := github.GetExternalAccountData(&accts[0].AccountData)
	if err != nil {
		return errors.Wrap(err, "get token")
	}

	client := github.NewV3Client(logger, svc.URN(), baseURL, &auth.OAuthBearerToken{Token: token.AccessToken}, *h.doer)
	g := repos.NewGitHubWebhookAPI(client)

	id, err := g.FindSyncWebhook(ctx, wbj.RepoName)
	if err != nil && err.Error() != "unable to find webhook" {
		return errors.Wrap(err, "find webhook")
	}

	if id != 0 {
		h.jobChan <- id
		return nil
	}

	secret := randstr.Hex(32)
	if err := addSecretToExtSvc(svc, "someOrg", secret); err != nil {
		return errors.Wrap(err, "add secret to External Service")
	}

	id, err = g.CreateSyncWebhook(ctx, wbj.RepoName, fmt.Sprintf("https://%s", globals.ExternalURL().Host), secret)
	if err != nil {
		return errors.Wrap(err, "create webhook")
	}

	h.jobChan <- id
	return nil
}

func addSecretToExtSvc(svc *types.ExternalService, org, secret string) error {
	var config schema.GitHubConnection
	err := json.Unmarshal([]byte(svc.Config), &config)
	if err != nil {
		return errors.Wrap(err, "unmarshal config")
	}

	config.Webhooks = append(config.Webhooks, &schema.GitHubWebhook{
		Org: org, Secret: secret,
	})

	newConfig, err := json.Marshal(config)
	if err != nil {
		return errors.Wrap(err, "marshal config")
	}

	svc.Config = string(newConfig)

	return nil
}

func newClient(serverURL string, doer httpcli.Doer) *ru.Client {
	return &ru.Client{
		URL:        serverURL,
		HTTPClient: doer,
	}
}

func webhookURLBuilderWithID(repoName string, hookID int) (string, error) {
	repoName = fmt.Sprintf("//%s", repoName)
	u, err := url.Parse(repoName)
	if err != nil {
		return "", errors.Newf("error parsing URL:", err)
	}

	if u.Host == "github.com" {
		return fmt.Sprintf("https://api.github.com/repos%s/hooks/%d", u.Path, hookID), nil
	}
	return fmt.Sprintf("https://%s/api/v3/repos%s/hooks/%d", u.Host, u.Path, hookID), nil
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
