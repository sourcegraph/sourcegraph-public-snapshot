package webhooks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/google/go-github/v47/github"
	"github.com/sourcegraph/log/logtest"
	fewebhooks "github.com/sourcegraph/sourcegraph/cmd/frontend/webhooks"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
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

func waitUntil(t *testing.T, condition chan bool) {
	t.Helper()
	select {
	case ret := <-condition:
		if !ret {
			t.Fatal("Expected condition to be true")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Timed out while waiting for condition")
	}
}

func TestGitHubWebhooks(t *testing.T) {
	sleepTime = 0
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	whStore := db.Webhooks(keyring.Default().WebhookKey)
	esStore := db.ExternalServices()

	u, err := db.Users().Create(context.Background(), database.NewUser{
		Username:        "testuser",
		EmailIsVerified: true,
	})
	require.NoError(t, err)

	accountID := int64(123)
	err = db.UserExternalAccounts().Insert(ctx, u.ID, extsvc.AccountSpec{
		ServiceType: extsvc.TypeGitHub,
		ServiceID:   "https://github.com/",
		AccountID:   strconv.Itoa(int(accountID)),
	}, extsvc.AccountData{})
	require.NoError(t, err)

	es := &types.ExternalService{
		Kind:        extsvc.KindGitHub,
		DisplayName: "GitHub",
		Config: extsvc.NewUnencryptedConfig(marshalJSON(t, &schema.GitHubConnection{
			Authorization: &schema.GitHubAuthorization{},
			Url:           "https://github.com/",
			Token:         "fake",
			Repos:         []string{"sourcegraph/sourcegraph"},
		})),
	}
	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}
	err = esStore.Create(ctx, confGet, es)
	require.NoError(t, err)

	repo := &types.Repo{
		Name: "github.com/sourcegraph/sourcegraph",
		URI:  "github.com/sourcegraph/sourcegraph",
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "R_kgDOIOwtPQ",
			ServiceType: extsvc.TypeGitHub,
			ServiceID:   "https://github.com/",
		},
		Metadata: map[string]any{"ID": "R_kgDOIOwtPQ"},
		Sources: map[string]*types.SourceInfo{
			es.URN(): {
				ID:       "R_kgDOIOwtPQ",
				CloneURL: "https://github.com/sourcegraph/sourcegraph",
			},
		},
	}

	ghWebhook := NewGitHubWebhook(logger)

	reposStore := repos.NewStore(logger, db)
	reposStore.CreateExternalServiceRepo(ctx, es, repo)

	wh, err := whStore.Create(ctx, "test-webhook", extsvc.KindGitHub, "https://github.com", u.ID, nil)
	require.NoError(t, err)

	hook := fewebhooks.GitHubWebhook{
		WebhookRouter: &fewebhooks.WebhookRouter{
			DB: db,
		},
	}

	ghWebhook.Register(hook.WebhookRouter)

	newReq := func(t *testing.T, event string, payload []byte) *http.Request {
		t.Helper()
		req, err := http.NewRequest("POST", fmt.Sprintf("/.api/webhooks/%v", wh.UUID), bytes.NewBuffer(payload))
		require.NoError(t, err)
		req.Header.Add("X-Github-Event", event)
		req.Header.Set("Content-Type", "application/json")
		return req
	}

	ghCloneURL := github.String("https://github.com/sourcegraph/sourcegraph.git")

	webhookTests := []struct {
		name      string
		eventType string
		event     any
		matchRepo bool // if false, match user
	}{
		{
			name:      "repository event",
			eventType: "repository",
			event: github.RepositoryEvent{
				Action: github.String("privatized"),
				Repo: &github.Repository{
					CloneURL: ghCloneURL,
				},
			},
			matchRepo: true,
		},
		{
			name:      "member event added",
			eventType: "member",
			event: github.MemberEvent{
				Action: github.String("added"),
				Member: &github.User{
					ID: github.Int64(accountID),
				},
				Repo: &github.Repository{
					CloneURL: ghCloneURL,
				},
			},
			matchRepo: false,
		},
		{
			name:      "member event removed",
			eventType: "member",
			event: github.MemberEvent{
				Action: github.String("removed"),
				Member: &github.User{
					ID: github.Int64(accountID),
				},
				Repo: &github.Repository{
					CloneURL: ghCloneURL,
				},
			},
			matchRepo: false,
		},
		{
			name:      "organization event member added",
			eventType: "organization",
			event: github.OrganizationEvent{
				Action: github.String("member_added"),
				Membership: &github.Membership{
					User: &github.User{
						ID: github.Int64(accountID),
					},
				},
			},
			matchRepo: false,
		},
		{
			name:      "organization event member removed",
			eventType: "organization",
			event: github.OrganizationEvent{
				Action: github.String("member_removed"),
				Membership: &github.Membership{
					User: &github.User{
						ID: github.Int64(accountID),
					},
				},
			},
			matchRepo: false,
		},
		{
			name:      "membership event added",
			eventType: "membership",
			event: github.MembershipEvent{
				Action: github.String("added"),
				Member: &github.User{
					ID: github.Int64(accountID),
				},
			},
			matchRepo: false,
		},
		{
			name:      "membership event removed",
			eventType: "membership",
			event: github.MembershipEvent{
				Action: github.String("removed"),
				Member: &github.User{
					ID: github.Int64(accountID),
				},
			},
			matchRepo: false,
		},
		{
			name:      "team event added to repository",
			eventType: "team",
			event: github.TeamEvent{
				Action: github.String("added_to_repository"),
				Repo: &github.Repository{
					CloneURL: ghCloneURL,
				},
			},
			matchRepo: true,
		},
		{
			name:      "team event removed from repository",
			eventType: "team",
			event: github.TeamEvent{
				Action: github.String("removed_from_repository"),
				Repo: &github.Repository{
					CloneURL: ghCloneURL,
				},
			},
			matchRepo: true,
		},
	}

	for _, webhookTest := range webhookTests {
		t.Run(webhookTest.name, func(t *testing.T) {
			webhookCalled := make(chan bool)
			repoupdater.MockSchedulePermsSync = func(ctx context.Context, args protocol.PermsSyncRequest) error {
				if webhookTest.matchRepo {
					webhookCalled <- args.RepoIDs[0] == repo.ID
				} else {
					webhookCalled <- args.UserIDs[0] == u.ID
				}
				return nil
			}
			t.Cleanup(func() { repoupdater.MockSchedulePermsSync = nil })

			payload, err := json.Marshal(webhookTest.event)
			require.NoError(t, err)
			req := newReq(t, webhookTest.eventType, payload)

			responseRecorder := httptest.NewRecorder()
			hook.ServeHTTP(responseRecorder, req)
			waitUntil(t, webhookCalled)
		})
	}
}
