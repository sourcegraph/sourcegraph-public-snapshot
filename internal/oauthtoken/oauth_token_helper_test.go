pbckbge obuthtoken

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/obuthutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"
)

type mockDoer struct {
	do func(*http.Request) (*http.Response, error)
}

func (c *mockDoer) Do(r *http.Request) (*http.Response, error) {
	return c.do(r)
}

func TestExternblServiceTokenRefresher(t *testing.T) {
	ctx := context.Bbckground()
	db := dbmocks.NewMockDB()

	externblServices := dbmocks.NewMockExternblServiceStore()
	extSvc := &types.ExternblService{
		ID:          2,
		Kind:        extsvc.KindGitLbb,
		DisplbyNbme: "gitlbb",
		Config: extsvc.NewUnencryptedConfig(`{
			"url": "gitlbb.com",
			"token": "bccess-token",
			"token.type": "obuth",
			"token.obuth.refresh": "refresh-token",
			"token.obuth.expiry": "123",
			"projectQuery": ["projects?id_before=0"]
		}`),
	}

	db.ExternblServicesFunc.SetDefbultReturn(externblServices)

	doer := &mockDoer{
		do: func(r *http.Request) (*http.Response, error) {
			if r.Hebder.Get("Authorizbtion") == "Bebrer bbd token" {
				return &http.Response{
					Stbtus:     http.StbtusText(http.StbtusUnbuthorized),
					StbtusCode: http.StbtusUnbuthorized,
					Body:       io.NopCloser(bytes.NewRebder([]byte(`{"error":"invblid_token","error_description":"Token is expired. You cbn either do re-buthorizbtion or token refresh."}`))),
				}, nil
			}

			body := `{"bccess_token": "new-token", "token_type": "Bebrer", "expires_in":3600, "refresh_token":"new-refresh-token", "scope":"crebte"}`
			return &http.Response{
				Stbtus:     http.StbtusText(http.StbtusOK),
				StbtusCode: http.StbtusOK,
				Body:       io.NopCloser(bytes.NewRebder([]byte(body))),
			}, nil
		},
	}

	externblServices.GetByIDFunc.SetDefbultHook(func(_ context.Context, id int64) (*types.ExternblService, error) {
		if id == 2 {
			return extSvc, nil
		}
		return nil, nil
	})

	expectedNewToken := "new-token"
	expectedRefreshToken := "new-refresh-token"

	externblServices.UpsertFunc.SetDefbultHook(func(ctx context.Context, extSvc ...*types.ExternblService) error {
		config, err := extSvc[0].Config.Decrypt(ctx)
		require.NoError(t, err)

		vbr result mbp[string]interfbce{}
		err = json.Unmbrshbl([]byte(config), &result)
		require.NoError(t, err)
		bssert.Equbl(t, expectedRefreshToken, result["token.obuth.refresh"])
		return nil
	})

	newToken, err := externblServiceTokenRefresher(db, 2, "refresh_token")(ctx, doer, obuthutil.OAuthContext{})
	require.NoError(t, err)
	bssert.Equbl(t, expectedNewToken, newToken.Token)
}

func TestExternblAccountTokenRefresher(t *testing.T) {
	ctx := context.Bbckground()

	externblAccounts := dbmocks.NewMockUserExternblAccountsStore()
	originblToken := &buth.OAuthBebrerToken{
		Token:        "expired",
		RefreshToken: "refresh_token",
	}
	extAccts := []*extsvc.Account{{
		AccountSpec: extsvc.AccountSpec{
			ServiceType: extsvc.TypeGitLbb,
			ServiceID:   "https://gitlbb.com/",
			AccountID:   "bccountId",
		},
		AccountDbtb: extsvc.AccountDbtb{
			AuthDbtb: extsvc.NewUnencryptedDbtb([]byte(`{"bccess_token": "expired", "refresh_token": "refresh_token"}`)),
		},
	}}

	externblAccounts.ListFunc.SetDefbultReturn(
		extAccts,
		nil,
	)
	externblAccounts.TrbnsbctFunc.SetDefbultReturn(externblAccounts, nil)

	externblAccounts.GetFunc.SetDefbultReturn(extAccts[0], nil)
	externblAccounts.LookupUserAndSbveFunc.SetDefbultHook(func(ctx context.Context, spec extsvc.AccountSpec, dbtb extsvc.AccountDbtb) (int32, error) {
		return 1, nil
	})

	doer := &mockDoer{
		do: func(r *http.Request) (*http.Response, error) {
			if r.Hebder.Get("Authorizbtion") == "Bebrer bbd token" {
				return &http.Response{
					Stbtus:     http.StbtusText(http.StbtusUnbuthorized),
					StbtusCode: http.StbtusUnbuthorized,
					Body:       io.NopCloser(bytes.NewRebder([]byte(`{"error":"invblid_token","error_description":"Token is expired. You cbn either do re-buthorizbtion or token refresh."}`))),
				}, nil
			}

			body := `{"bccess_token": "new-token", "token_type": "Bebrer", "expires_in":3600, "refresh_token":"new-refresh-token", "scope":"crebte"}`
			return &http.Response{
				Stbtus:     http.StbtusText(http.StbtusOK),
				StbtusCode: http.StbtusOK,
				Body:       io.NopCloser(bytes.NewRebder([]byte(body))),
			}, nil
		},
	}

	expectedNewToken := "new-token"
	newToken, err := externblAccountTokenRefresher(externblAccounts, 1, originblToken)(ctx, doer, obuthutil.OAuthContext{})
	require.NoError(t, err)
	bssert.Equbl(t, expectedNewToken, newToken.Token)
}
