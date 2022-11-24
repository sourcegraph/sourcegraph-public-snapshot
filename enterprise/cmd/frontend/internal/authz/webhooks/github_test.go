package webhooks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-github/v47/github"
	"github.com/sourcegraph/log/logtest"
	fewebhooks "github.com/sourcegraph/sourcegraph/cmd/frontend/webhooks"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func marshalJSON(t testing.TB, v any) string {
	t.Helper()

	bs, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	return string(bs)
}

func stringPointer(v string) *string {
	return &v
}

func int64Pointer(v int64) *int64 {
	return &v
}

func TestGitHubWebhooks(t *testing.T) {
	logger := logtest.Scoped(t)
	sqlDB := dbtest.NewDB(logger, t)
	db := database.NewDB(logger, sqlDB)
	ctx := context.Background()
	u, err := db.Users().Create(context.Background(), database.NewUser{
		Email:           "test@user.com",
		Username:        "testuser",
		EmailIsVerified: true,
	})
	require.NoError(t, err)

	require.NoError(t, db.UserExternalAccounts().Insert(ctx, u.ID, extsvc.AccountSpec{
		ServiceType: extsvc.TypeGitHub,
		ServiceID:   "https://github.com/",
		ClientID:    "",
		AccountID:   "123",
	}, extsvc.AccountData{}))

	err = db.Repos().Create(ctx, &types.Repo{
		ID:   1,
		Name: "github.com/sourcegraph/sourcegraph",
		URI:  "github.com/sourcegraph/sourcegraph",
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "1234",
			ServiceType: "github",
			ServiceID:   "https://github.com",
		},
	})

	require.NoError(t, err)

	webhooksStore := db.Webhooks(keyring.Default().WebhookKey)
	esStore := db.ExternalServices()
	extSvc := &types.ExternalService{
		Kind:        extsvc.KindGitHub,
		DisplayName: "GitHub",
		Config: extsvc.NewUnencryptedConfig(marshalJSON(t, &schema.GitHubConnection{
			Authorization: &schema.GitHubAuthorization{},
			Url:           "https://github.com",
			Token:         "fake",
			Repos:         []string{"sourcegraph/sourcegraph"},
		})),
	}

	err = esStore.Upsert(ctx, extSvc)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, "INSERT INTO external_service_repos(external_service_id, repo_id, clone_url) VALUES (1, 1, 'https://github.com/sourcegraph/sourcegraph');")
	require.NoError(t, err)

	externalServiceConfig := fmt.Sprintf(`
{
    // Some comment to mess with json decoding
    "url": "https://github.com",
    "token": "fake",
    "repos": ["sourcegraph/sourcegraph"]
}
`)

	require.NoError(t, esStore.Update(ctx, []schema.AuthProviders{}, 1, &database.ExternalServiceUpdate{Config: &externalServiceConfig}))

	ghWebhook := NewGitHubWebhook()

	wh, err := webhooksStore.Create(ctx, extsvc.KindGitHub, "https://github.com", u.ID, nil)
	require.NoError(t, err)

	hook := fewebhooks.GitHubWebhook{
		WebhookRouter: &fewebhooks.WebhookRouter{
			DB: db,
		},
	}

	ghWebhook.Register(hook.WebhookRouter)

	t.Run("repository event", func(t *testing.T) {
		webhookCalled := false
		repoupdater.MockSchedulePermsSync = func(ctx context.Context, args protocol.PermsSyncRequest) error {
			webhookCalled = args.RepoIDs[0] == 1
			return nil
		}
		t.Cleanup(func() { repoupdater.MockSchedulePermsSync = nil })

		payload, err := json.Marshal(github.RepositoryEvent{
			Action: stringPointer("privatized"),
			Repo: &github.Repository{
				CloneURL: stringPointer("https://github.com/sourcegraph/sourcegraph.git"),
			},
		})
		require.NoError(t, err)
		req, err := http.NewRequest("POST", fmt.Sprintf("/.api/webhooks/%v", wh.UUID), bytes.NewBuffer(payload))
		req.Header.Add("X-Github-Event", "repository")
		req.Header.Set("Content-Type", "application/json")

		responseRecorder := httptest.NewRecorder()
		hook.ServeHTTP(responseRecorder, req)
		assert.True(t, webhookCalled)
	})

	t.Run("member event added", func(t *testing.T) {
		webhookCalled := false
		repoupdater.MockSchedulePermsSync = func(ctx context.Context, args protocol.PermsSyncRequest) error {
			webhookCalled = args.UserIDs[0] == 1
			return nil
		}
		t.Cleanup(func() { repoupdater.MockSchedulePermsSync = nil })

		payload, err := json.Marshal(github.MemberEvent{
			Action: stringPointer("added"),
			Member: &github.User{
				ID: int64Pointer(123),
			},
			Repo: &github.Repository{
				CloneURL: stringPointer("https://github.com/sourcegraph/sourcegraph.git"),
			},
		})
		require.NoError(t, err)
		req, err := http.NewRequest("POST", fmt.Sprintf("/.api/webhooks/%v", wh.UUID), bytes.NewBuffer(payload))
		req.Header.Add("X-Github-Event", "member")
		req.Header.Set("Content-Type", "application/json")

		responseRecorder := httptest.NewRecorder()
		hook.ServeHTTP(responseRecorder, req)
		assert.True(t, webhookCalled)
	})

	t.Run("member event removed", func(t *testing.T) {
		webhookCalled := false
		repoupdater.MockSchedulePermsSync = func(ctx context.Context, args protocol.PermsSyncRequest) error {
			webhookCalled = args.UserIDs[0] == 1
			return nil
		}
		t.Cleanup(func() { repoupdater.MockSchedulePermsSync = nil })

		payload, err := json.Marshal(github.MemberEvent{
			Action: stringPointer("removed"),
			Member: &github.User{
				ID: int64Pointer(123),
			},
			Repo: &github.Repository{
				CloneURL: stringPointer("https://github.com/sourcegraph/sourcegraph.git"),
			},
		})
		require.NoError(t, err)
		req, err := http.NewRequest("POST", fmt.Sprintf("/.api/webhooks/%v", wh.UUID), bytes.NewBuffer(payload))
		req.Header.Add("X-Github-Event", "member")
		req.Header.Set("Content-Type", "application/json")

		responseRecorder := httptest.NewRecorder()
		hook.ServeHTTP(responseRecorder, req)
		assert.True(t, webhookCalled)
	})

	t.Run("organization event member added", func(t *testing.T) {
		webhookCalled := false
		repoupdater.MockSchedulePermsSync = func(ctx context.Context, args protocol.PermsSyncRequest) error {
			webhookCalled = args.UserIDs[0] == 1
			return nil
		}
		t.Cleanup(func() { repoupdater.MockSchedulePermsSync = nil })

		payload, err := json.Marshal(github.OrganizationEvent{
			Action: stringPointer("member_added"),
			Membership: &github.Membership{User: &github.User{
				ID: int64Pointer(123),
			}},
		})
		require.NoError(t, err)
		req, err := http.NewRequest("POST", fmt.Sprintf("/.api/webhooks/%v", wh.UUID), bytes.NewBuffer(payload))
		req.Header.Add("X-Github-Event", "organization")
		req.Header.Set("Content-Type", "application/json")

		responseRecorder := httptest.NewRecorder()
		hook.ServeHTTP(responseRecorder, req)
		assert.True(t, webhookCalled)
	})

	t.Run("organization event member removed", func(t *testing.T) {
		webhookCalled := false
		repoupdater.MockSchedulePermsSync = func(ctx context.Context, args protocol.PermsSyncRequest) error {
			webhookCalled = args.UserIDs[0] == 1
			return nil
		}
		t.Cleanup(func() { repoupdater.MockSchedulePermsSync = nil })

		payload, err := json.Marshal(github.OrganizationEvent{
			Action: stringPointer("member_removed"),
			Membership: &github.Membership{User: &github.User{
				ID: int64Pointer(123),
			}},
		})
		require.NoError(t, err)
		req, err := http.NewRequest("POST", fmt.Sprintf("/.api/webhooks/%v", wh.UUID), bytes.NewBuffer(payload))
		req.Header.Add("X-Github-Event", "organization")
		req.Header.Set("Content-Type", "application/json")

		responseRecorder := httptest.NewRecorder()
		hook.ServeHTTP(responseRecorder, req)
		assert.True(t, webhookCalled)
	})

	t.Run("membership event added", func(t *testing.T) {
		webhookCalled := false
		repoupdater.MockSchedulePermsSync = func(ctx context.Context, args protocol.PermsSyncRequest) error {
			webhookCalled = args.UserIDs[0] == 1
			return nil
		}
		t.Cleanup(func() { repoupdater.MockSchedulePermsSync = nil })

		payload, err := json.Marshal(github.MembershipEvent{
			Action: stringPointer("added"),
			Member: &github.User{
				ID: int64Pointer(123),
			},
		})
		require.NoError(t, err)
		req, err := http.NewRequest("POST", fmt.Sprintf("/.api/webhooks/%v", wh.UUID), bytes.NewBuffer(payload))
		req.Header.Add("X-Github-Event", "membership")
		req.Header.Set("Content-Type", "application/json")

		responseRecorder := httptest.NewRecorder()
		hook.ServeHTTP(responseRecorder, req)
		assert.True(t, webhookCalled)
	})

	t.Run("membership event removed", func(t *testing.T) {
		webhookCalled := false
		repoupdater.MockSchedulePermsSync = func(ctx context.Context, args protocol.PermsSyncRequest) error {
			webhookCalled = args.UserIDs[0] == 1
			return nil
		}
		t.Cleanup(func() { repoupdater.MockSchedulePermsSync = nil })

		payload, err := json.Marshal(github.MembershipEvent{
			Action: stringPointer("removed"),
			Member: &github.User{
				ID: int64Pointer(123),
			},
		})
		require.NoError(t, err)
		req, err := http.NewRequest("POST", fmt.Sprintf("/.api/webhooks/%v", wh.UUID), bytes.NewBuffer(payload))
		req.Header.Add("X-Github-Event", "membership")
		req.Header.Set("Content-Type", "application/json")

		responseRecorder := httptest.NewRecorder()
		hook.ServeHTTP(responseRecorder, req)
		assert.True(t, webhookCalled)
	})

	t.Run("team event added to repository", func(t *testing.T) {
		webhookCalled := false
		repoupdater.MockSchedulePermsSync = func(ctx context.Context, args protocol.PermsSyncRequest) error {
			webhookCalled = args.RepoIDs[0] == 1
			return nil
		}
		t.Cleanup(func() { repoupdater.MockSchedulePermsSync = nil })

		payload, err := json.Marshal(github.TeamEvent{
			Action: stringPointer("added_to_repository"),
			Repo: &github.Repository{
				CloneURL: stringPointer("https://github.com/sourcegraph/sourcegraph.git"),
			},
		})
		require.NoError(t, err)
		req, err := http.NewRequest("POST", fmt.Sprintf("/.api/webhooks/%v", wh.UUID), bytes.NewBuffer(payload))
		req.Header.Add("X-Github-Event", "team")
		req.Header.Set("Content-Type", "application/json")

		responseRecorder := httptest.NewRecorder()
		hook.ServeHTTP(responseRecorder, req)
		assert.True(t, webhookCalled)
	})

	t.Run("team event removed from repository", func(t *testing.T) {
		webhookCalled := false
		repoupdater.MockSchedulePermsSync = func(ctx context.Context, args protocol.PermsSyncRequest) error {
			webhookCalled = args.RepoIDs[0] == 1
			return nil
		}
		t.Cleanup(func() { repoupdater.MockSchedulePermsSync = nil })

		payload, err := json.Marshal(github.TeamEvent{
			Action: stringPointer("removed_from_repository"),
			Repo: &github.Repository{
				CloneURL: stringPointer("https://github.com/sourcegraph/sourcegraph.git"),
			},
		})
		require.NoError(t, err)
		req, err := http.NewRequest("POST", fmt.Sprintf("/.api/webhooks/%v", wh.UUID), bytes.NewBuffer(payload))
		req.Header.Add("X-Github-Event", "team")
		req.Header.Set("Content-Type", "application/json")

		responseRecorder := httptest.NewRecorder()
		hook.ServeHTTP(responseRecorder, req)
		assert.True(t, webhookCalled)
	})
}
