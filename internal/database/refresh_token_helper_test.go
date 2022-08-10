package database

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/oauthutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type mockDoer struct {
	do func(*http.Request) (*http.Response, error)
}

func (c *mockDoer) Do(r *http.Request) (*http.Response, error) {
	return c.do(r)
}

func TestRefreshToken_ExternalServices(t *testing.T) {
	ctx := context.Background()
	ctxOauth := oauthutil.OauthContext{}
	db := NewMockDB()

	externalServices := NewMockExternalServiceStore()
	extSvc := &types.ExternalService{
		ID:          2,
		Kind:        extsvc.KindGitLab,
		DisplayName: "gitlab",
		Config: `{
			"url": "gitlab.com",
			"token": "access-token",
			"token.type": "oauth",
			"token.oauth.refresh": "refresh-token",
			"token.oauth.expiry": "123",
			"projectQuery": ["projects?id_before=0"]
		}`,
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
		var result map[string]interface{}
		_ = json.Unmarshal([]byte(extSvc[0].Config), &result)

		if result["token.oauth.refresh"] != expectedRefreshToken {
			t.Fatalf("got %v, want %v", result["token.oauth.refresh"], expectedRefreshToken)
		}

		return nil
	})

	h := &RefreshTokenHelperForExternalService{DB: db, ExternalServiceID: 2, OauthRefreshToken: "refresh_token"}
	newToken, err := h.RefreshToken(ctx, doer, ctxOauth)

	if newToken != expectedNewToken {
		t.Fatalf("got %v, want %v", newToken, expectedNewToken)
	}
	if err != nil {
		t.Fatal(err)
	}
}

func TestRefreshToken_ExternalAccounts(t *testing.T) {
	ctx := context.Background()
	ctxOauth := oauthutil.OauthContext{}
	db := NewMockDB()

	externalAccounts := NewMockUserExternalAccountsStore()
	extAccts := []*extsvc.Account{{
		AccountSpec: extsvc.AccountSpec{
			ServiceType: extsvc.TypeGitLab,
			ServiceID:   "https://gitlab.com/",
			AccountID:   "accountId",
		},
	}}

	externalAccounts.ListFunc.SetDefaultReturn(
		extAccts,
		nil,
	)

	externalAccounts.GetFunc.SetDefaultReturn(extAccts[0], nil)
	externalAccounts.LookupUserAndSaveFunc.SetDefaultHook(func(ctx context.Context, spec extsvc.AccountSpec, data extsvc.AccountData) (int32, error) {
		return 1, nil
	})

	db.UserExternalAccountsFunc.SetDefaultReturn(externalAccounts)

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

	h := &RefreshTokenHelperForExternalAccount{DB: db, ExternalAccountID: 1, OauthRefreshToken: "refresh_token"}
	newToken, err := h.RefreshToken(ctx, doer, ctxOauth)

	if newToken != expectedNewToken {
		t.Fatalf("got %v, want %v", newToken, expectedNewToken)
	}
	if err != nil {
		t.Fatal(err)
	}
}
