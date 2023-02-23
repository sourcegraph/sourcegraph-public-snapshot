package webhooks

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/external/globals"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/webhooks"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketcloud"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	gitlabwebhooks "github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab/webhooks"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestGitHubHandler(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)

	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := repos.NewStore(logger, db)
	repoStore := store.RepoStore()
	esStore := store.ExternalServiceStore()

	repo := &types.Repo{
		ID:   1,
		Name: "ghe.sgdev.org/milton/test",
	}
	if err := repoStore.Create(ctx, repo); err != nil {
		t.Fatal(err)
	}

	conn := schema.GitHubConnection{
		Url:      "https://ghe.sgdev.org",
		Token:    "token",
		Repos:    []string{"milton/test"},
		Webhooks: []*schema.GitHubWebhook{{Org: "ghe.sgdev.org", Secret: "secret"}},
	}

	config, err := json.Marshal(conn)
	if err != nil {
		t.Fatal(err)
	}

	svc := &types.ExternalService{
		Kind:        extsvc.KindGitHub,
		DisplayName: "TestService",
		Config:      extsvc.NewUnencryptedConfig(string(config)),
	}
	if err := esStore.Upsert(ctx, svc); err != nil {
		t.Fatal(err)
	}

	handler := NewGitHubHandler()
	router := &webhooks.GitHubWebhook{
		Router: &webhooks.Router{
			DB: db,
		},
	}
	handler.Register(router.Router)

	mux := http.NewServeMux()
	mux.HandleFunc("/enqueue-repo-update", func(w http.ResponseWriter, r *http.Request) {
		reqBody, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		var req protocol.RepoUpdateRequest
		if err := json.Unmarshal(reqBody, &req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		repositories, err := repoStore.List(ctx, database.ReposListOptions{Names: []string{string(req.Repo)}})
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
		}
		if len(repositories) != 1 {
			http.Error(w, fmt.Sprintf("expected 1 repo, got %v", len(repositories)), http.StatusNotFound)
		}

		repo := repositories[0]
		res := &protocol.RepoUpdateResponse{
			ID:   repo.ID,
			Name: string(repo.Name),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(res)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	cf := httpcli.NewExternalClientFactory()
	opts := []httpcli.Opt{}
	doer, err := cf.Doer(opts...)
	if err != nil {
		t.Fatal(err)
	}

	repoupdater.DefaultClient = &repoupdater.Client{
		URL:        server.URL,
		HTTPClient: doer,
	}

	payload, err := os.ReadFile(filepath.Join("testdata", "github-push.json"))
	if err != nil {
		t.Fatal(err)
	}

	targetURL := fmt.Sprintf("%s/github-webhooks", globals.ExternalURL())
	req, err := http.NewRequest("POST", targetURL, bytes.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("X-GitHub-Event", "push")
	req.Header.Set("X-Hub-Signature", sign(t, payload, []byte("secret")))

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	resp := rec.Result()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status code: 200, got %v", resp.StatusCode)
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

func TestGitLabHandler(t *testing.T) {
	repoName := "gitlab.com/ryanslade/ryan-test-private"

	db := database.NewMockDB()
	repositories := database.NewMockRepoStore()
	repositories.GetFirstRepoNameByCloneURLFunc.SetDefaultHook(func(ctx context.Context, s string) (api.RepoName, error) {
		return api.RepoName(repoName), nil
	})
	db.ReposFunc.SetDefaultReturn(repositories)

	handler := NewGitLabHandler()
	data, err := os.ReadFile("testdata/gitlab-push.json")
	if err != nil {
		t.Fatal(err)
	}
	var payload gitlabwebhooks.PushEvent
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatal(err)
	}

	var updateQueued string
	repoupdater.MockEnqueueRepoUpdate = func(ctx context.Context, repo api.RepoName) (*protocol.RepoUpdateResponse, error) {
		updateQueued = string(repo)
		return &protocol.RepoUpdateResponse{
			ID:   1,
			Name: string(repo),
		}, nil
	}
	t.Cleanup(func() { repoupdater.MockEnqueueRepoUpdate = nil })

	if err := handler.handlePushEvent(context.Background(), db, &payload); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, repoName, updateQueued)
}

func TestBitbucketServerHandler(t *testing.T) {
	repoName := "bitbucket.sgdev.org/private/test-2020-06-01"

	db := database.NewMockDB()
	repositories := database.NewMockRepoStore()
	repositories.GetFirstRepoNameByCloneURLFunc.SetDefaultHook(func(ctx context.Context, s string) (api.RepoName, error) {
		return "bitbucket.sgdev.org/private/test-2020-06-01", nil
	})
	db.ReposFunc.SetDefaultReturn(repositories)

	handler := NewBitbucketServerHandler()
	data, err := os.ReadFile("testdata/bitbucket-server-push.json")
	if err != nil {
		t.Fatal(err)
	}
	var payload bitbucketserver.PushEvent
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatal(err)
	}

	var updateQueued string
	repoupdater.MockEnqueueRepoUpdate = func(ctx context.Context, repo api.RepoName) (*protocol.RepoUpdateResponse, error) {
		updateQueued = string(repo)
		return &protocol.RepoUpdateResponse{
			ID:   1,
			Name: string(repo),
		}, nil
	}
	t.Cleanup(func() { repoupdater.MockEnqueueRepoUpdate = nil })

	if err := handler.handlePushEvent(context.Background(), db, &payload); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, repoName, updateQueued)
}

func TestBitbucketCloudHandler(t *testing.T) {
	repoName := "bitbucket.org/sourcegraph-testing/sourcegraph"

	db := database.NewMockDB()
	repositories := database.NewMockRepoStore()
	repositories.GetFirstRepoNameByCloneURLFunc.SetDefaultHook(func(ctx context.Context, s string) (api.RepoName, error) {
		return "bitbucket.org/sourcegraph-testing/sourcegraph", nil
	})
	db.ReposFunc.SetDefaultReturn(repositories)

	handler := NewBitbucketCloudHandler()
	data, err := os.ReadFile("testdata/bitbucket-cloud-push.json")
	if err != nil {
		t.Fatal(err)
	}
	var payload bitbucketcloud.PushEvent
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatal(err)
	}

	var updateQueued string
	repoupdater.MockEnqueueRepoUpdate = func(ctx context.Context, repo api.RepoName) (*protocol.RepoUpdateResponse, error) {
		updateQueued = string(repo)
		return &protocol.RepoUpdateResponse{
			ID:   1,
			Name: string(repo),
		}, nil
	}
	t.Cleanup(func() { repoupdater.MockEnqueueRepoUpdate = nil })

	if err := handler.handlePushEvent(context.Background(), db, &payload); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, repoName, updateQueued)
}
