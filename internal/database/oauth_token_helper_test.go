package database

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/oauthutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type mockDoer struct {
	do func(*http.Request) (*http.Response, error)
}

func (c *mockDoer) Do(r *http.Request) (*http.Response, error) {
	return c.do(r)
}

func TestExternalServiceTokenRefresher(t *testing.T) {
	ctx := context.Background()
	db := NewMockDB()

	externalServices := NewMockExternalServiceStore()
	extSvc := &types.ExternalService{
		ID:          2,
		Kind:        extsvc.KindGitLab,
		DisplayName: "gitlab",
		Config: extsvc.NewUnencryptedConfig(`{
			"url": "gitlab.com",
			"token": "access-token",
			"token.type": "oauth",
			"token.oauth.refresh": "refresh-token",
			"token.oauth.expiry": "123",
			"projectQuery": ["projects?id_before=0"]
		}`),
	}

	db.ExternalServicesFunc.SetDefaultReturn(externalServices)

	doer := &mockDoer{
		do: func(r *http.Request) (*http.Response, error) {
			if r.Header.Get("Authorization") == "Bearer bad token" {
				return &http.Response{
					Status:     http.StatusText(http.StatusUnauthorized),
					StatusCode: http.StatusUnauthorized,
					Body:       io.NopCloser(bytes.NewReader([]byte(`{"error":"invalid_token","error_description":"Token is expired. You can either do re-authorization or token refresh."}`))),
				}, nil
			}

			body := `{"access_token": "new-token", "token_type": "Bearer", "expires_in":3600, "refresh_token":"new-refresh-token", "scope":"create"}`
			return &http.Response{
				Status:     http.StatusText(http.StatusOK),
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader([]byte(body))),
			}, nil
		},
	}

	externalServices.GetByIDFunc.SetDefaultHook(func(_ context.Context, id int64) (*types.ExternalService, error) {
		if id == 2 {
			return extSvc, nil
		}
		return nil, nil
	})

	expectedNewToken := "new-token"
	expectedRefreshToken := "new-refresh-token"

	externalServices.UpsertFunc.SetDefaultHook(func(ctx context.Context, extSvc ...*types.ExternalService) error {
		config, err := extSvc[0].Config.Decrypt(ctx)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal([]byte(config), &result)
		require.NoError(t, err)
		assert.Equal(t, expectedRefreshToken, result["token.oauth.refresh"])
		return nil
	})

	newToken, err := externalServiceTokenRefresher(db, 2, "refresh_token")(ctx, doer, oauthutil.OAuthContext{})
	require.NoError(t, err)
	assert.Equal(t, expectedNewToken, newToken.Token)
}

func TestExternalAccountTokenRefresher(t *testing.T) {
	ctx := context.Background()

	externalAccounts := NewMockUserExternalAccountsStore()
	originalToken := &auth.OAuthBearerToken{
		Token:        "expired",
		RefreshToken: "refresh_token",
	}
	extAccts := []*extsvc.Account{{
		AccountSpec: extsvc.AccountSpec{
			ServiceType: extsvc.TypeGitLab,
			ServiceID:   "https://gitlab.com/",
			AccountID:   "accountId",
		},
		AccountData: extsvc.AccountData{
			AuthData: extsvc.NewUnencryptedData([]byte(`{"access_token": "expired", "refresh_token": "refresh_token"}`)),
		},
	}}

	externalAccounts.ListFunc.SetDefaultReturn(
		extAccts,
		nil,
	)
	externalAccounts.TransactFunc.SetDefaultReturn(externalAccounts, nil)

	externalAccounts.GetFunc.SetDefaultReturn(extAccts[0], nil)
	externalAccounts.LookupUserAndSaveFunc.SetDefaultHook(func(ctx context.Context, spec extsvc.AccountSpec, data extsvc.AccountData) (int32, error) {
		return 1, nil
	})

	doer := &mockDoer{
		do: func(r *http.Request) (*http.Response, error) {
			if r.Header.Get("Authorization") == "Bearer bad token" {
				return &http.Response{
					Status:     http.StatusText(http.StatusUnauthorized),
					StatusCode: http.StatusUnauthorized,
					Body:       io.NopCloser(bytes.NewReader([]byte(`{"error":"invalid_token","error_description":"Token is expired. You can either do re-authorization or token refresh."}`))),
				}, nil
			}

			body := `{"access_token": "new-token", "token_type": "Bearer", "expires_in":3600, "refresh_token":"new-refresh-token", "scope":"create"}`
			return &http.Response{
				Status:     http.StatusText(http.StatusOK),
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader([]byte(body))),
			}, nil
		},
	}

	expectedNewToken := "new-token"
	newToken, err := externalAccountTokenRefresher(externalAccounts, 1, originalToken)(ctx, doer, oauthutil.OAuthContext{})
	require.NoError(t, err)
	assert.Equal(t, expectedNewToken, newToken.Token)
}
