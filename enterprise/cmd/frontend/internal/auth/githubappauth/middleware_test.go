package githubapp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/github_apps/store"
	ghtypes "github.com/sourcegraph/sourcegraph/enterprise/internal/github_apps/types"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestGenerateRedirectURL(t *testing.T) {
	reposDomain := "repos"
	batchesDomain := "batches"
	invalidDomain := "invalid"

	testCases := []struct {
		name           string
		domain         *string
		installationID int
		appID          int
		expectedURL    string
		expectedError  error
	}{
		{
			name:           "repos domain",
			domain:         &reposDomain,
			installationID: 1,
			appID:          2,
			expectedURL:    "/site-admin/github-apps/R2l0SHViQXBwOjI=?installation_id=1",
		},
		{
			name:           "batches domain",
			domain:         &batchesDomain,
			installationID: 1,
			appID:          2,
			expectedURL:    "/site-admin/batch-changes",
		},
		{
			name:          "invalid domain",
			domain:        &invalidDomain,
			expectedError: errors.Errorf("invalid domain: %s", invalidDomain),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url, err := generateRedirectURL(tc.domain, tc.installationID, tc.appID)
			if tc.expectedError != nil {
				require.EqualError(t, err, tc.expectedError.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedURL, url)
			}
		})
	}
}

func TestGithubAppAuthMiddleware(t *testing.T) {
	webhookUUID := uuid.New()

	mockUserStore := database.NewMockUserStore()
	mockUserStore.GetByCurrentAuthUserFunc.SetDefaultHook(func(ctx context.Context) (*types.User, error) {
		a := actor.FromContext(ctx)
		return &types.User{
			ID:        a.UID,
			SiteAdmin: a.UID == 2,
		}, nil
	})

	mockWebhookStore := database.NewMockWebhookStore()
	mockWebhookStore.CreateFunc.SetDefaultHook(func(ctx context.Context, name, kind, urn string, actorUID int32, e *encryption.Encryptable) (*types.Webhook, error) {
		return &types.Webhook{
			ID:              1,
			UUID:            webhookUUID,
			Name:            name,
			CodeHostKind:    kind,
			CreatedByUserID: actorUID,
			UpdatedByUserID: actorUID,
		}, nil
	})
	mockWebhookStore.GetByUUIDFunc.SetDefaultReturn(&types.Webhook{
		ID:   1,
		UUID: webhookUUID,
		Name: "test-github-app",
	}, nil)
	mockWebhookStore.UpdateFunc.SetDefaultHook(func(ctx context.Context, w *types.Webhook) (*types.Webhook, error) {
		return w, nil
	})

	mockGitHubAppsStore := store.NewMockGitHubAppsStore()
	mockGitHubAppsStore.CreateFunc.SetDefaultReturn(1, nil)
	mockGitHubAppsStore.GetByIDFunc.SetDefaultHook(func(ctx context.Context, id int) (*ghtypes.GitHubApp, error) {
		return &ghtypes.GitHubApp{
			ID: id,
		}, nil
	})

	db := edb.NewMockEnterpriseDB()

	db.UsersFunc.SetDefaultReturn(mockUserStore)
	db.WebhooksFunc.SetDefaultReturn(mockWebhookStore)
	db.GitHubAppsFunc.SetDefaultReturn(mockGitHubAppsStore)

	rcache.SetupForTest(t)
	cache := rcache.NewWithTTL("test_cache", 5)

	mux := newServeMux(db, "/githubapp", cache)

	t.Run("/state", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/githubapp/state", nil)

		t.Run("regular user", func(t *testing.T) {
			req = req.WithContext(actor.WithActor(req.Context(), &actor.Actor{
				UID: 1,
			}))

			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)

			if w.Code != http.StatusForbidden {
				t.Fatalf("expected status code %d but got %d", http.StatusForbidden, w.Code)
			}
		})

		t.Run("site-admin", func(t *testing.T) {
			req = req.WithContext(actor.WithActor(req.Context(), &actor.Actor{
				UID: 2,
			}))

			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("expected status code %d but got %d", http.StatusOK, w.Code)
			}

			state := w.Body.String()
			if state == "" {
				t.Fatal("expected non-empty state in response")
			}

			cachedState, ok := cache.Get(state)
			if !ok {
				t.Fatal("expected state to be cached")
			}

			var stateDetails gitHubAppStateDetails
			if err := json.Unmarshal(cachedState, &stateDetails); err != nil {
				t.Fatalf("unexpected error unmarshalling cached state: %s", err.Error())
			}

			if stateDetails.AppID != 0 {
				t.Fatal("expected AppID to be 0 for empty state")
			}
		})
	})

	t.Run("/new-app-state", func(t *testing.T) {
		webhookURN := "https://example.com"
		appName := "TestApp"
		domain := "batches"
		req := httptest.NewRequest("GET", fmt.Sprintf("/githubapp/new-app-state?webhookURN=%s&appName=%s&domain=%s", webhookURN, appName, domain), nil)

		t.Run("normal user", func(t *testing.T) {
			req = req.WithContext(actor.WithActor(req.Context(), &actor.Actor{
				UID: 1,
			}))

			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)

			if w.Code != http.StatusForbidden {
				t.Fatalf("expected status code %d but got %d", http.StatusForbidden, w.Code)
			}
		})

		t.Run("site-admin", func(t *testing.T) {
			req = req.WithContext(actor.WithActor(req.Context(), &actor.Actor{
				UID: 2,
			}))

			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("expected status code %d but got %d", http.StatusOK, w.Code)
			}

			var resp struct {
				State       string `json:"state"`
				WebhookUUID string `json:"webhookUUID,omitempty"`
			}
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("unexpected error decoding response: %s", err.Error())
			}

			if resp.State == "" {
				t.Fatal("expected non-empty state in response")
			}
			if resp.WebhookUUID == "" {
				t.Fatal("expected non-empty webhookUUID in response")
			}

			cachedState, ok := cache.Get(resp.State)
			if !ok {
				t.Fatal("expected state to be cached")
			}

			var stateDetails gitHubAppStateDetails
			if err := json.Unmarshal(cachedState, &stateDetails); err != nil {
				t.Fatalf("unexpected error unmarshalling cached state: %s", err.Error())
			}

			if stateDetails.WebhookUUID != resp.WebhookUUID {
				t.Fatal("expected webhookUUID in state details to match response")
			}
			if stateDetails.Domain != domain {
				t.Fatal("expected domain in state details to match request param")
			}
		})
	})

	t.Run("/redirect", func(t *testing.T) {
		baseURL := "/githubapp/redirect"
		code := "2644896245sasdsf6dsd"
		state := uuid.New()
		domain := types.BatchesDomain

		t.Run("normal user", func(t *testing.T) {
			req := httptest.NewRequest("GET", baseURL, nil)
			req = req.WithContext(actor.WithActor(req.Context(), &actor.Actor{
				UID: 1,
			}))

			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)

			if w.Code != http.StatusForbidden {
				t.Fatalf("expected status code %d but got %d", http.StatusForbidden, w.Code)
			}
		})

		t.Run("without state", func(t *testing.T) {
			req := httptest.NewRequest("GET", fmt.Sprintf("%s?code=%s", baseURL, code), nil)
			req = req.WithContext(actor.WithActor(req.Context(), &actor.Actor{
				UID: 2,
			}))

			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected status code %d but got %d", http.StatusBadRequest, w.Code)
			}
		})

		t.Run("without code", func(t *testing.T) {
			req := httptest.NewRequest("GET", fmt.Sprintf("%s?state=%s", baseURL, state), nil)
			req = req.WithContext(actor.WithActor(req.Context(), &actor.Actor{
				UID: 2,
			}))

			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected status code %d but got %d", http.StatusBadRequest, w.Code)
			}
		})

		t.Run("success", func(t *testing.T) {
			MockCreateGitHubApp = func(conversionURL string, domain types.GitHubAppDomain) (*ghtypes.GitHubApp, error) {
				return &ghtypes.GitHubApp{}, nil
			}
			req := httptest.NewRequest("GET", fmt.Sprintf("%s?state=%s&code=%s", baseURL, state, code), nil)
			req.Header.Set("Referer", "https://example.com")
			req = req.WithContext(actor.WithActor(req.Context(), &actor.Actor{
				UID: 2,
			}))

			stateDeets, err := json.Marshal(gitHubAppStateDetails{
				WebhookUUID: webhookUUID.String(),
				Domain:      string(domain),
			})
			require.NoError(t, err)
			cache.Set(state.String(), stateDeets)

			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)

			if w.Code != http.StatusSeeOther {
				t.Fatalf("expected status code %d but got %d", http.StatusOK, w.Code)
			}
		})
	})

	t.Run("/setup", func(t *testing.T) {
		baseURL := "/githubapp/setup"
		state := uuid.New()
		installationID := 232034243
		domain := types.BatchesDomain

		t.Run("normal user", func(t *testing.T) {
			req := httptest.NewRequest("GET", baseURL, nil)
			req = req.WithContext(actor.WithActor(req.Context(), &actor.Actor{
				UID: 1,
			}))

			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)

			if w.Code != http.StatusForbidden {
				t.Fatalf("expected status code %d but got %d", http.StatusForbidden, w.Code)
			}
		})

		t.Run("without state", func(t *testing.T) {
			req := httptest.NewRequest("GET", fmt.Sprintf("%s?installation_id=%d", baseURL, installationID), nil)
			req = req.WithContext(actor.WithActor(req.Context(), &actor.Actor{
				UID: 2,
			}))

			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)

			if w.Code != http.StatusFound {
				t.Fatalf("expected status code %d but got %d", http.StatusBadRequest, w.Code)
			}
		})

		t.Run("without installation_id", func(t *testing.T) {
			req := httptest.NewRequest("GET", fmt.Sprintf("%s?state=%s", baseURL, state), nil)
			req = req.WithContext(actor.WithActor(req.Context(), &actor.Actor{
				UID: 2,
			}))

			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)

			if w.Code != http.StatusFound {
				t.Fatalf("expected status code %d but got %d", http.StatusBadRequest, w.Code)
			}
		})

		t.Run("without setup_action", func(t *testing.T) {
			req := httptest.NewRequest("GET", fmt.Sprintf("%s?installation_id=%d&state=%s", baseURL, installationID, state), nil)
			req = req.WithContext(actor.WithActor(req.Context(), &actor.Actor{
				UID: 2,
			}))

			stateDeets, err := json.Marshal(gitHubAppStateDetails{})
			require.NoError(t, err)
			cache.Set(state.String(), stateDeets)

			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected status code %d but got %d", http.StatusBadRequest, w.Code)
			}
		})

		t.Run("success", func(t *testing.T) {
			req := httptest.NewRequest("GET", fmt.Sprintf("%s?installation_id=%d&state=%s&setup_action=install", baseURL, installationID, state), nil)
			req = req.WithContext(actor.WithActor(req.Context(), &actor.Actor{
				UID: 2,
			}))

			stateDeets, err := json.Marshal(gitHubAppStateDetails{
				Domain: string(domain),
			})
			require.NoError(t, err)
			cache.Set(state.String(), stateDeets)

			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)

			if w.Code != http.StatusFound {
				t.Fatalf("expected status code %d but got %d", http.StatusOK, w.Code)
			}
		})
	})
}
