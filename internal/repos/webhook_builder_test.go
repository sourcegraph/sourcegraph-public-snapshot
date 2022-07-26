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

		repoName := "ghe.sgdev.org/milton/test"
		err := godotenv.Load("./.env")
		if err != nil {
			t.Fatal(err)
		}
		token := os.Getenv("ACCESS_TOKEN")

		esStore := store.ExternalServiceStore()
		repoStore := store.RepoStore()

		ghConn := &schema.GitHubConnection{
			Url:      extsvc.KindGitHub,
			Token:    token,
			Repos:    []string{repoName},
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

		repo := &types.Repo{
			ID:       1,
			Name:     api.RepoName(repoName),
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
		gh := repos.NewGitHubWebhookAPI(client)

		success, err := gh.Client.TestPushSyncWebhook(ctx, repoName, id)
		if err != nil || !success {
			t.Fatal(err)
		}

		defer func() {
			deleted, err := gh.Client.DeleteSyncWebhook(ctx, repoName, id)
			if err != nil || !deleted {
				t.Fatal(err)
			}
		}()

		handler := webhooks.GitHubWebhook{
			ExternalServices: store.ExternalServiceStore(),
		}
		gh.Register(&handler)

		payloadData := []byte(`
		{
			"ref": "refs/heads/main",
			"before": "0000000000000000000000000000000000000000",
			"after": "9cae21b79a239d5b02a33f8323c6770c418d16a0",
			"repository": {
			  "id": 442310,
			  "node_id": "MDEwOlJlcG9zaXRvcnk0NDIzMTA=",
			  "name": "test",
			  "full_name": "milton/test",
			  "private": false,
			  "owner": {
				"name": "milton",
				"email": "dev@sourcegraph.com",
				"login": "milton",
				"id": 3,
				"node_id": "MDQ6VXNlcjM=",
				"avatar_url": "https://ghe.sgdev.org/avatars/u/3?",
				"gravatar_id": "",
				"url": "https://ghe.sgdev.org/api/v3/users/milton",
				"html_url": "https://ghe.sgdev.org/milton",
				"followers_url": "https://ghe.sgdev.org/api/v3/users/milton/followers",
				"following_url": "https://ghe.sgdev.org/api/v3/users/milton/following{/other_user}",
				"gists_url": "https://ghe.sgdev.org/api/v3/users/milton/gists{/gist_id}",
				"starred_url": "https://ghe.sgdev.org/api/v3/users/milton/starred{/owner}{/repo}",
				"subscriptions_url": "https://ghe.sgdev.org/api/v3/users/milton/subscriptions",
				"organizations_url": "https://ghe.sgdev.org/api/v3/users/milton/orgs",
				"repos_url": "https://ghe.sgdev.org/api/v3/users/milton/repos",
				"events_url": "https://ghe.sgdev.org/api/v3/users/milton/events{/privacy}",
				"received_events_url": "https://ghe.sgdev.org/api/v3/users/milton/received_events",
				"type": "User",
				"site_admin": true
			  },
			  "html_url": "https://ghe.sgdev.org/milton/test",
			  "description": null,
			  "fork": false,
			  "url": "https://ghe.sgdev.org/milton/test",
			  "forks_url": "https://ghe.sgdev.org/api/v3/repos/milton/test/forks",
			  "keys_url": "https://ghe.sgdev.org/api/v3/repos/milton/test/keys{/key_id}",
			  "collaborators_url": "https://ghe.sgdev.org/api/v3/repos/milton/test/collaborators{/collaborator}",
			  "teams_url": "https://ghe.sgdev.org/api/v3/repos/milton/test/teams",
			  "hooks_url": "https://ghe.sgdev.org/api/v3/repos/milton/test/hooks",
			  "issue_events_url": "https://ghe.sgdev.org/api/v3/repos/milton/test/issues/events{/number}",
			  "events_url": "https://ghe.sgdev.org/api/v3/repos/milton/test/events",
			  "assignees_url": "https://ghe.sgdev.org/api/v3/repos/milton/test/assignees{/user}",
			  "branches_url": "https://ghe.sgdev.org/api/v3/repos/milton/test/branches{/branch}",
			  "tags_url": "https://ghe.sgdev.org/api/v3/repos/milton/test/tags",
			  "blobs_url": "https://ghe.sgdev.org/api/v3/repos/milton/test/git/blobs{/sha}",
			  "git_tags_url": "https://ghe.sgdev.org/api/v3/repos/milton/test/git/tags{/sha}",
			  "git_refs_url": "https://ghe.sgdev.org/api/v3/repos/milton/test/git/refs{/sha}",
			  "trees_url": "https://ghe.sgdev.org/api/v3/repos/milton/test/git/trees{/sha}",
			  "statuses_url": "https://ghe.sgdev.org/api/v3/repos/milton/test/statuses/{sha}",
			  "languages_url": "https://ghe.sgdev.org/api/v3/repos/milton/test/languages",
			  "stargazers_url": "https://ghe.sgdev.org/api/v3/repos/milton/test/stargazers",
			  "contributors_url": "https://ghe.sgdev.org/api/v3/repos/milton/test/contributors",
			  "subscribers_url": "https://ghe.sgdev.org/api/v3/repos/milton/test/subscribers",
			  "subscription_url": "https://ghe.sgdev.org/api/v3/repos/milton/test/subscription",
			  "commits_url": "https://ghe.sgdev.org/api/v3/repos/milton/test/commits{/sha}",
			  "git_commits_url": "https://ghe.sgdev.org/api/v3/repos/milton/test/git/commits{/sha}",
			  "comments_url": "https://ghe.sgdev.org/api/v3/repos/milton/test/comments{/number}",
			  "issue_comment_url": "https://ghe.sgdev.org/api/v3/repos/milton/test/issues/comments{/number}",
			  "contents_url": "https://ghe.sgdev.org/api/v3/repos/milton/test/contents/{+path}",
			  "compare_url": "https://ghe.sgdev.org/api/v3/repos/milton/test/compare/{base}...{head}",
			  "merges_url": "https://ghe.sgdev.org/api/v3/repos/milton/test/merges",
			  "archive_url": "https://ghe.sgdev.org/api/v3/repos/milton/test/{archive_format}{/ref}",
			  "downloads_url": "https://ghe.sgdev.org/api/v3/repos/milton/test/downloads",
			  "issues_url": "https://ghe.sgdev.org/api/v3/repos/milton/test/issues{/number}",
			  "pulls_url": "https://ghe.sgdev.org/api/v3/repos/milton/test/pulls{/number}",
			  "milestones_url": "https://ghe.sgdev.org/api/v3/repos/milton/test/milestones{/number}",
			  "notifications_url": "https://ghe.sgdev.org/api/v3/repos/milton/test/notifications{?since,all,participating}",
			  "labels_url": "https://ghe.sgdev.org/api/v3/repos/milton/test/labels{/name}",
			  "releases_url": "https://ghe.sgdev.org/api/v3/repos/milton/test/releases{/id}",
			  "deployments_url": "https://ghe.sgdev.org/api/v3/repos/milton/test/deployments",
			  "created_at": 1657136088,
			  "updated_at": "2022-07-06T19:34:50Z",
			  "pushed_at": 1657136088,
			  "git_url": "git://ghe.sgdev.org/milton/test.git",
			  "ssh_url": "git@ghe.sgdev.org:milton/test.git",
			  "clone_url": "https://ghe.sgdev.org/milton/test.git",
			  "svn_url": "https://ghe.sgdev.org/milton/test",
			  "homepage": null,
			  "size": 0,
			  "stargazers_count": 0,
			  "watchers_count": 0,
			  "language": null,
			  "has_issues": true,
			  "has_projects": true,
			  "has_downloads": true,
			  "has_wiki": true,
			  "has_pages": false,
			  "forks_count": 0,
			  "mirror_url": null,
			  "archived": false,
			  "disabled": false,
			  "open_issues_count": 0,
			  "license": null,
			  "allow_forking": true,
			  "is_template": false,
			  "visibility": "public",
			  "forks": 0,
			  "open_issues": 0,
			  "watchers": 0,
			  "default_branch": "main",
			  "stargazers": 0,
			  "master_branch": "main"
			},
			"pusher": {
			  "name": "milton",
			  "email": "dev@sourcegraph.com"
			},
			"enterprise": {
			  "id": 1,
			  "slug": "sourcegraph",
			  "name": "Sourcegraph",
			  "node_id": "MDEwOkVudGVycHJpc2Ux",
			  "avatar_url": "https://ghe.sgdev.org/avatars/b/1?",
			  "description": "",
			  "website_url": "",
			  "html_url": "https://ghe.sgdev.org/enterprises/sourcegraph",
			  "created_at": "2019-05-23T19:36:13Z",
			  "updated_at": "2021-12-06T15:58:48Z"
			},
			"sender": {
			  "login": "milton",
			  "id": 3,
			  "node_id": "MDQ6VXNlcjM=",
			  "avatar_url": "https://ghe.sgdev.org/avatars/u/3?",
			  "gravatar_id": "",
			  "url": "https://ghe.sgdev.org/api/v3/users/milton",
			  "html_url": "https://ghe.sgdev.org/milton",
			  "followers_url": "https://ghe.sgdev.org/api/v3/users/milton/followers",
			  "following_url": "https://ghe.sgdev.org/api/v3/users/milton/following{/other_user}",
			  "gists_url": "https://ghe.sgdev.org/api/v3/users/milton/gists{/gist_id}",
			  "starred_url": "https://ghe.sgdev.org/api/v3/users/milton/starred{/owner}{/repo}",
			  "subscriptions_url": "https://ghe.sgdev.org/api/v3/users/milton/subscriptions",
			  "organizations_url": "https://ghe.sgdev.org/api/v3/users/milton/orgs",
			  "repos_url": "https://ghe.sgdev.org/api/v3/users/milton/repos",
			  "events_url": "https://ghe.sgdev.org/api/v3/users/milton/events{/privacy}",
			  "received_events_url": "https://ghe.sgdev.org/api/v3/users/milton/received_events",
			  "type": "User",
			  "site_admin": true
			},
			"created": true,
			"deleted": false,
			"forced": false,
			"base_ref": null,
			"compare": "https://ghe.sgdev.org/milton/test/commit/9cae21b79a23",
			"commits": [
			  {
				"id": "9cae21b79a239d5b02a33f8323c6770c418d16a0",
				"tree_id": "7713cad4d6de429f9bfc9bf9efcc09f9481d26a1",
				"distinct": true,
				"message": "Initial commit",
				"timestamp": "2022-07-06T15:34:48-04:00",
				"url": "https://ghe.sgdev.org/milton/test/commit/9cae21b79a239d5b02a33f8323c6770c418d16a0",
				"author": {
				  "name": "milton",
				  "email": "dev@sourcegraph.com",
				  "username": "milton"
				},
				"committer": {
				  "name": "GitHub Enterprise",
				  "email": "noreply@35.232.213.103"
				},
				"added": [
				  "README.md"
				],
				"removed": [],
				"modified": []
			  }
			],
			"head_commit": {
			  "id": "9cae21b79a239d5b02a33f8323c6770c418d16a0",
			  "tree_id": "7713cad4d6de429f9bfc9bf9efcc09f9481d26a1",
			  "distinct": true,
			  "message": "Initial commit",
			  "timestamp": "2022-07-06T15:34:48-04:00",
			  "url": "https://ghe.sgdev.org/milton/test/commit/9cae21b79a239d5b02a33f8323c6770c418d16a0",
			  "author": {
				"name": "milton",
				"email": "dev@sourcegraph.com",
				"username": "milton"
			  },
			  "committer": {
				"name": "GitHub Enterprise",
				"email": "noreply@35.232.213.103"
			  },
			  "added": [
				"README.md"
			  ],
			  "removed": [],
			  "modified": []
			}
		  }
		`)

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

		httpSrv := httpserver.NewFromAddr("127.0.0.1:8080", &http.Server{
			ReadTimeout:  75 * time.Second,
			WriteTimeout: 10 * time.Minute,
			Handler:      server.Handler(),
		})

		go func() {
			httpSrv.Start()
			defer httpSrv.Stop()
		}()

		// mock the Github PushEvent
		// e.g. user makes a commit
		url := fmt.Sprintf("%s/github-webhooks", globals.ExternalURL())
		req, err := http.NewRequest("POST", url, bytes.NewReader(payloadData))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("X-Github-Event", "push")
		req.Header.Set("X-Hub-Signature", sign(t, payloadData, []byte("secret")))

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
		h.jobChan <- -1
		return errors.Errorf("expected repos.WhBuildJob, got %T", record)
	}

	switch wbj.ExtsvcKind {
	case "GITHUB":
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
		gh := repos.NewGitHubWebhookAPI(client)

		id, err := gh.Client.FindSyncWebhook(ctx, wbj.RepoName)
		if err != nil {
			return errors.Wrap(err, "find webhook")
		}
		secret := randstr.Hex(32)

		if err := addSecretToExtSvc(svc, "someOrg", secret); err != nil {
			return errors.Wrap(err, "add secret to External Service")
		}

		id, err = gh.Client.CreateSyncWebhook(ctx, wbj.RepoName, fmt.Sprintf("https://%s", globals.ExternalURL().Host), secret)
		if err != nil {
			return errors.Wrap(err, "create webhook")
		}
		h.jobChan <- id
	}

	return nil
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
	fmt.Println("svc:", svc)

	return nil
}
