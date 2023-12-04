package oauthtoken

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/oauthutil"
)

type mockDoer struct {
	do func(*http.Request) (*http.Response, error)
}

func (c *mockDoer) Do(r *http.Request) (*http.Response, error) {
	return c.do(r)
}

func TestExternalAccountTokenRefresher(t *testing.T) {
	ctx := context.Background()

	externalAccounts := dbmocks.NewMockUserExternalAccountsStore()
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
	externalAccounts.UpdateFunc.SetDefaultHook(func(ctx context.Context, acct *extsvc.Account) (*extsvc.Account, error) {
		return &extsvc.Account{UserID: 1}, nil
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
