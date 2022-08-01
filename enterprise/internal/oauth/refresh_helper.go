package oauth

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"golang.org/x/oauth2"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/oauthutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Next steps
// 0 - Use ids to retrieve the objects
// 1 - Fix compiling errors
// 2 - gilaboauath / githuboauth providers
//

type RefreshTokenHelperForExternalAccount struct {
	DB                database.DB
	ExternalAccountID int32
}

type RefreshTokenHelperForExternalService struct {
	DB                database.DB
	ExternalServiceID int64
}

func (r *RefreshTokenHelperForExternalAccount) RefreshToken(ctx context.Context, doer httpcli.Doer, oauthCtx oauthutil.OauthContext) (string, error) {
	refreshedToken, err := oauthutil.RetrieveToken(ctx, doer, oauthCtx, oauthutil.AuthStyleInParams)

	defer func() {
		success := err == nil
		gitlab.TokenRefreshCounter.WithLabelValues("external_account", strconv.FormatBool(success)).Inc()
	}()

	acct, err := r.DB.UserExternalAccounts().Get(ctx, r.ExternalAccountID)
	if err != nil {
		return "", errors.Wrap(err, "getting user external account")
	}

	acct.SetAuthData(refreshedToken)
	_, err = r.DB.UserExternalAccounts().LookupUserAndSave(ctx, acct.AccountSpec, acct.AccountData)
	if err != nil {
		return "", errors.Wrap(err, "save refreshed token")
	}

	return "", nil
}

func (r *RefreshTokenHelperForExternalService) RefreshToken(ctx context.Context, doer httpcli.Doer, oauthCtx oauthutil.OauthContext) (string, error) {
	fmt.Println(".......RefreshToken original funcion")

	refreshedToken, err := oauthutil.RetrieveToken(ctx, doer, oauthCtx, oauthutil.AuthStyleInParams)

	defer func() {
		success := err == nil
		gitlab.TokenRefreshCounter.WithLabelValues("codehost", strconv.FormatBool(success)).Inc()
	}()

	extsvc, err := r.DB.ExternalServices().GetByID(ctx, r.ExternalServiceID)
	if err != nil {
		return "", errors.Wrap(err, "getting external service")
	}
	extsvc.Config, err = jsonc.Edit(extsvc.Config, refreshedToken.AccessToken, "token")
	if err != nil {
		return "", errors.Wrap(err, "updating OAuth token")
	}
	extsvc.Config, err = jsonc.Edit(extsvc.Config, refreshedToken.RefreshToken, "token.oauth.refresh")
	if err != nil {
		return "", errors.Wrap(err, "updating OAuth refresh token")
	}
	extsvc.Config, err = jsonc.Edit(extsvc.Config, refreshedToken.Expiry.Unix(), "token.oauth.expiry")
	if err != nil {
		return "", errors.Wrap(err, "updating OAuth token expiry")
	}
	extsvc.UpdatedAt = time.Now()
	if err := r.DB.ExternalServices().Upsert(ctx, extsvc); err != nil {
		return "", errors.Wrap(err, "upserting external service")
	}

	return "", nil
}

// todo - we have a similar function on perms_syncer. It would be better to avoid dupes or find out how to use the
// same function in both places...s
func Oauth2ConfigFromGitLabProvider() *oauth2.Config {
	for _, authProvider := range conf.SiteConfig().AuthProviders {
		if authProvider.Gitlab != nil {
			p := authProvider.Gitlab

			url := strings.TrimSuffix(p.Url, "/")
			return &oauth2.Config{
				ClientID:     p.ClientID,
				ClientSecret: p.ClientSecret,
				Endpoint: oauth2.Endpoint{
					AuthURL:  url + "/oauth/authorize",
					TokenURL: url + "/oauth/token",
				},
				Scopes: gitlab.RequestedOAuthScopes(p.ApiScope, nil),
			}

		}
	}

	// todo  - log warning
	return nil
}
