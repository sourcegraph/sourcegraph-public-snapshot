package syncwebhooks

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/sourcegraph/log/logtest"

	"github.com/google/go-github/github"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/webhooks"
	rp "github.com/sourcegraph/sourcegraph/cmd/repo-updater/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/httptestutil"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

var update = flag.Bool("update", false, "update sync testdata")

func testGitHubSyncHooks(db database.DB, userID int32) func(*testing.T) {
	return func(t *testing.T) {
		ctx := context.Background()
		logger := logtest.Scoped(t)

		svc := types.ExternalService{
			Kind: extsvc.KindGitHub,
			Config: `{
				"URL": "https://github.com",
				"Token": "secret-token"
				}`,
		}

		repo := types.Repo{
			ID:   1,
			Name: "Hello-World",
			ExternalRepo: api.ExternalRepoSpec{
				ID:          "hi mom",
				ServiceType: extsvc.TypeGitHub,
				ServiceID:   "https://github.com",
			},
			Metadata: new(github.Repository),
		}

		initStore := func(db database.DB) repos.Store {
			store := repos.NewStore(logger, db)
			if err := store.ExternalServiceStore().Upsert(ctx, &svc); err != nil {
				fmt.Println("err:", err)
				t.Fatal(err)
			}
			if err := store.RepoStore().Create(ctx, &repo); err != nil {
				fmt.Println("repostore", err)
				t.Fatal(err)
			}
			return store
		}

		type testCase struct {
			name string
			repo api.RepoName
			res  *protocol.RepoUpdateResponse
			err  string
			init func(database.DB) repos.Store
		}

		tcs := []testCase{
			{
				name: "valid repo",
				repo: repo.Name,
				init: initStore,
				res: &protocol.RepoUpdateResponse{
					ID:   repo.ID,
					Name: string(repo.Name),
				},
			},
		}

		for _, tc := range tcs {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				sqlDB := dbtest.NewDB(t)
				db := database.NewDB(sqlDB)
				store := tc.init(db)
				s := &rp.Server{
					Logger:    logger,
					Store:     store,
					Scheduler: repos.NewUpdateScheduler(logger, db),
				}
				server := httptest.NewServer(s.Handler())
				defer server.Close()
				activatePushWebhook(db, server.URL, t)
			})
		}
	}
}

func activatePushWebhook(db database.DB, url string, t *testing.T) {
	ctx := context.Background()
	rcache.SetupForTest(t)

	cf, save := httptestutil.NewGitHubRecorderFactory(t, *update, "github-webhooks")
	defer save()

	secret := "secret"
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		token = "no-GITHUB_TOKEN-set"
	}

	repoStore := db.Repos()
	esStore := db.ExternalServices()
	extSvc := &types.ExternalService{
		Kind:        extsvc.KindGitHub,
		DisplayName: "GitHub",
		Config: MarshalJSON(t, &schema.GitHubConnection{
			Url:      "https://github.com",
			Token:    token,
			Repos:    []string{"sourcegraph/sourcegraph"},
			Webhooks: []*schema.GitHubWebhook{{Org: "sourcegraph", Secret: secret}},
		}),
	}

	err := esStore.Upsert(ctx, extSvc)
	if err != nil {
		t.Fatal(err)
	}

	githubSrc, err := repos.NewGithubSource(db.ExternalServices(), extSvc, cf)
	if err != nil {
		t.Fatal(err)
	}

	githubRepo, err := githubSrc.GetRepo(ctx, "sourcegraph/sourcegraph")
	if err != nil {
		t.Fatal(err)
	}
	err = repoStore.Create(ctx, githubRepo)
	if err != nil {
		t.Fatal(err)
	}

	sc := SyncWebhook{}

	bs, err := os.ReadFile(filepath.Join("testdata", "github.json"))
	if err != nil {
		t.Fatal(err)
	}

	var event github.PushEvent
	if err := json.Unmarshal(bs, &event); err != nil {
		t.Fatal(err)
	}

	handler := webhooks.GitHubWebhook{
		ExternalServices: esStore,
	}
	sc.Register(&handler)
	Url = url // for testing purposes

	body, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("X-Github-Event", "push")
	req.Header.Set("X-Hub-Signature", sign(t, body, []byte(secret)))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	resp := rec.Result()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Non 200 code: %v", resp.StatusCode)
	}

	if *update {
		// do something
	}

}

// missing:
// push_id *int64,
// head *string,
// size *int,
// distinct_size *int,
// installation *Installation
