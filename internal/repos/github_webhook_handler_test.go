package repos_test

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/external/globals"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/webhooks"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestGitHubWebhookHandle(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)

	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := repos.NewStore(logger, db)
	esStore := store.ExternalServiceStore()

	conn := *&schema.GitHubConnection{
		Url:      extsvc.KindGitHub,
		Token:    "token",
		Repos:    []string{},
		Webhooks: []*schema.GitHubWebhook{{Org: "ghe.sgdev.org", Secret: "secret"}},
	}

	config, err := json.Marshal(conn)
	if err != nil {
		t.Fatal(err)
	}

	svc := &types.ExternalService{
		Kind:        extsvc.KindGitHub,
		DisplayName: "TestService",
		Config:      string(config),
	}
	if err := esStore.Upsert(ctx, svc); err != nil {
		t.Fatal(err)
	}

	handler := repos.GitHubWebhookHandler{}
	router := &webhooks.GitHubWebhook{
		ExternalServices: esStore,
	}
	handler.Register(router)

	payload, err := os.ReadFile(filepath.Join("testdata", "github-ping.json"))
	if err != nil {
		t.Fatal(err)
	}

	targetURL := fmt.Sprintf("%s/github-webhooks", globals.ExternalURL())
	req, err := http.NewRequest("POST", targetURL, bytes.NewReader([]byte(payload)))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("X-GitHub-Event", "push")
	req.Header.Set("X-Hub-Signature", sign(t, payload, []byte("secret")))

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	resp := rec.Result()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if !(string(data) == `Post "http://repo-updater:3182/enqueue-repo-update": dial tcp: lookup repo-updater: no such host`+"\n") {
		t.Fatal(errors.Newf(`expected body: [%s], got [%s]`,
			`Post "http://repo-updater:3182/enqueue-repo-update": dial tcp: lookup repo-updater: no such host`,
			string(data),
		))
	}

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected status code: 500, got %v", resp.StatusCode)
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
