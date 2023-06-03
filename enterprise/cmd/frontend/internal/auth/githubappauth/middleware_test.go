package githubapp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
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
			expectedURL:    "/site-admin/batch-changes/github-apps/R2l0SHViQXBwOjI=?installation_id=1",
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
	db := database.NewMockEnterpriseDB()
	rcache.SetupForTest(t)
	cache := rcache.NewWithTTL("test_cache", 5)

	mux := newServeMux(db, "/githubapp", cache)

	t.Run("/state", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/githubapp/state", nil)

		t.Run("regular user", func(t *testing.T) {
			req = req.WithContext(actor.WithActor(req.Context(), &actor.Actor{}))

			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("expected status code %d but got %d", http.StatusOK, w.Code)
			}

			// state := w.Body.String()
			// if state == "" {
			// 	t.Fatal("expected non-empty state in response")
			// }

			// cachedState, ok := cache.Get(state)
			// if !ok {
			// 	t.Fatal("expected state to be cached")
			// }

			// var stateDetails gitHubAppStateDetails
			// if err := json.Unmarshal(cachedState, &stateDetails); err != nil {
			// 	t.Fatalf("unexpected error unmarshalling cached state: %s", err.Error())
			// }

			// if stateDetails.AppID != 0 {
			// 	t.Fatal("expected AppID to be 0 for empty state")
			// }
		})

		t.Run("site-admin", func(t *testing.T) {
			t.Skip()
			newCtx := actor.WithActor(req.Context(), &actor.Actor{
				UID: 1,
			})
			req = req.WithContext(newCtx)

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
		t.Skip()
		webhookURN := "https://example.com"
		appName := "TestApp"
		domain := "batches"

		fmt.Println(fmt.Sprintf("/githubapp/new-app-state?webhookURN=%s&appName=%s&domain=%s", webhookURN, appName, domain), "<====")
		req := httptest.NewRequest("GET", fmt.Sprintf("/githubapp/new-app-state?webhookURN=%s&appName=%s&domain=%s", webhookURN, appName, domain), nil)
		newCtx := actor.WithActor(req.Context(), &actor.Actor{
			Internal: true,
		})
		req = req.WithContext(newCtx)

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
}
