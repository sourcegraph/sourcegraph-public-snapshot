package httpapi

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	gh "github.com/google/go-github/v28/github"

	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/schema"
)

var dsn = flag.String("dsn", "", "Database connection string to use in integration tests")

func TestDispatchSuccess(t *testing.T) {
	h := GithubWebhook{}
	var called bool
	h.Register("test-event-1", func(ctx context.Context, svc *repos.ExternalService, payload interface{}) error {
		called = true
		return nil
	})

	ctx := context.Background()
	if err := h.Dispatch(ctx, "test-event-1", nil, nil); err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	if !called {
		t.Errorf("Expected called to be true, was false")
	}
}

func TestDispatchNoHandler(t *testing.T) {
	h := GithubWebhook{}
	ctx := context.Background()
	// no op
	if err := h.Dispatch(ctx, "test-event-1", nil, nil); err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
}

func TestDispatchSuccessMultiple(t *testing.T) {
	h := GithubWebhook{}
	var called int
	h.Register("test-event-1", func(ctx context.Context, svc *repos.ExternalService, payload interface{}) error {
		called++
		return nil
	})
	h.Register("test-event-1", func(ctx context.Context, svc *repos.ExternalService, payload interface{}) error {
		called++
		return nil
	})

	ctx := context.Background()
	if err := h.Dispatch(ctx, "test-event-1", nil, nil); err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	if called != 2 {
		t.Errorf("Expected called to be 2, got %v", called)
	}
}

func TestDispatchError(t *testing.T) {
	h := GithubWebhook{}
	var called int
	h.Register("test-event-1", func(ctx context.Context, svc *repos.ExternalService, payload interface{}) error {
		called++
		return fmt.Errorf("oh dear")
	})
	h.Register("test-event-1", func(ctx context.Context, svc *repos.ExternalService, payload interface{}) error {
		called++
		return nil
	})

	ctx := context.Background()
	if err := h.Dispatch(ctx, "test-event-1", nil, nil); errString(err) != "oh dear" {
		t.Errorf("Expected 'oh no', got %s", err)
	}
	if called != 1 {
		t.Errorf("Expected called to be 1, got %v", called)
	}
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func TestExternalServices(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	dbtesting.SetupGlobalTestDB(t)
	db := dbtest.NewDB(t, *dsn)

	ctx := context.Background()

	secret := "secret"
	repoStore := repos.NewDBStore(db, sql.TxOptions{})
	extSvc := &repos.ExternalService{
		Kind:        extsvc.KindGitHub,
		DisplayName: "GitHub",
		Config: marshalJSON(t, &schema.GitHubConnection{
			Url:      "https://github.com",
			Token:    os.Getenv("GITHUB_TOKEN"),
			Repos:    []string{"sourcegraph/sourcegraph"},
			Webhooks: []*schema.GitHubWebhook{{Org: "sourcegraph", Secret: secret}},
		}),
	}

	err := repoStore.UpsertExternalServices(ctx, extSvc)
	if err != nil {
		t.Fatal(t)
	}

	hook := GithubWebhook{
		Repos: repoStore,
	}

	var called bool
	hook.Register("public", func(ctx context.Context, extSvc *repos.ExternalService, payload interface{}) error {
		evt, ok := payload.(*gh.PublicEvent)
		if !ok {
			t.Errorf("Expected *gh.PublicEvent event, got %T", payload)
		}
		if evt.GetRepo().GetFullName() != "sourcegraph/sourcegraph" {
			t.Errorf("Expected 'sourcegraph/sourcegraph', got %s", evt.GetRepo().GetFullName())
		}
		called = true
		return nil
	})

	u := extsvc.WebhookURL(extsvc.TypeGitHub, extSvc.ID, "https://example.com/")

	req, err := http.NewRequest("POST", u, bytes.NewReader(eventPayload))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("X-Github-Event", "public")
	req.Header.Set("X-Hub-Signature", sign(t, eventPayload, []byte(secret)))

	rec := httptest.NewRecorder()
	hook.ServeHTTP(rec, req)
	resp := rec.Result()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Non 200 code: %v", resp.StatusCode)
	}

	if !called {
		t.Fatalf("Expected called to be true, got false (webhook handler was not called)")
	}
}

func insertTestUser(t *testing.T, db *sql.DB) (userID int32) {
	t.Helper()

	err := db.QueryRow("INSERT INTO users (username) VALUES ('bbs-admin') RETURNING id").Scan(&userID)
	if err != nil {
		t.Fatal(err)
	}

	return userID
}

func truncateTables(t *testing.T, db *sql.DB, tables ...string) {
	t.Helper()

	_, err := db.Exec("TRUNCATE " + strings.Join(tables, ", ") + " RESTART IDENTITY")
	if err != nil {
		t.Fatal(err)
	}
}

func marshalJSON(t testing.TB, v interface{}) string {
	t.Helper()

	bs, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	return string(bs)
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

var eventPayload = []byte(`{
  "repository": {
    "id": 310572870,
    "node_id": "MDEwOlJlcG9zaXRvcnkzMTA1NzI4NzA=",
    "name": "sourcegraph",
    "full_name": "sourcegraph/sourcegraph",
    "private": false,
    "owner": {
      "login": "sourcegraph",
      "id": 74051180,
      "node_id": "MDEyOk9yZ2FuaXphdGlvbjc0MDUxMTgw",
      "type": "Organization",
      "site_admin": false
    },
    "html_url": "https://github.com/sourcegraph",
    "created_at": "2020-11-06T11:02:56Z",
    "updated_at": "2020-11-09T15:06:34Z",
    "pushed_at": "2020-11-06T11:02:58Z",
    "default_branch": "main"
  },
  "organization": {
    "login": "sourcegraph",
    "id": 74051180,
    "node_id": "MDEyOk9yZ2FuaXphdGlvbjc0MDUxMTgw",
    "description": null
  },
  "sender": {
    "login": "sourcegraph",
    "id": 5236823,
    "node_id": "MDQ6VXNlcjUyMzY4MjM=",
    "type": "User",
    "site_admin": false
  }
}`)
