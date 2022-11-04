package database

import (
	"context"
	"strconv"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/oauthutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// ExternalAccountTokenRefresher returns an oauthutil.TokenRefresher for the
// given external account.
func ExternalAccountTokenRefresher(db DB, externalAccountID int32, refreshToken string) oauthutil.TokenRefresher {
	return func(ctx context.Context, doer httpcli.Doer, oauthCtx oauthutil.OAuthContext) (token string, err error) {
		defer func() {
			success := err == nil
			gitlab.TokenRefreshCounter.WithLabelValues("external_account", strconv.FormatBool(success)).Inc()
		}()

		refreshedToken, err := oauthutil.RetrieveToken(doer, oauthCtx, refreshToken, oauthutil.AuthStyleInParams)
		if err != nil {
			return "", errors.Wrap(err, "refresh token")
		}

		acct, err := db.UserExternalAccounts().Get(ctx, externalAccountID)
		if err != nil {
			return "", errors.Wrap(err, "get user external account")
		}

		err = acct.AuthData.Set(refreshedToken)
		if err != nil {
			return "", errors.Wrap(err, "set auth data")
		}
		_, err = db.UserExternalAccounts().LookupUserAndSave(ctx, acct.AccountSpec, acct.AccountData)
		if err != nil {
			return "", errors.Wrap(err, "save refreshed token")
		}
		return refreshedToken.AccessToken, nil
	}
}

// ExternalServiceTokenRefresher returns an oauthutil.TokenRefresher for the
// given external service.
func ExternalServiceTokenRefresher(db DB, externalServiceID int64, refreshToken string) oauthutil.TokenRefresher {
	return func(ctx context.Context, doer httpcli.Doer, oauthCtx oauthutil.OAuthContext) (token string, err error) {
		defer func() {
			success := err == nil
			gitlab.TokenRefreshCounter.WithLabelValues("codehost", strconv.FormatBool(success)).Inc()
		}()

		refreshedToken, err := oauthutil.RetrieveToken(doer, oauthCtx, refreshToken, oauthutil.AuthStyleInParams)
		if err != nil {
			return "", errors.Wrap(err, "refresh token")
		}

		extsvc, err := db.ExternalServices().GetByID(ctx, externalServiceID)
		if err != nil {
			return "", errors.Wrap(err, "get external service")
		}

		config, err := extsvc.Config.Decrypt(ctx)
		if err != nil {
			return "", errors.Wrap(err, "decrypt config")
		}

		config, err = jsonc.Edit(config, refreshedToken.AccessToken, "token")
		if err != nil {
			return "", errors.Wrap(err, "update OAuth token")
		}
		config, err = jsonc.Edit(config, refreshedToken.RefreshToken, "token.oauth.refresh")
		if err != nil {
			return "", errors.Wrap(err, "update OAuth refresh token")
		}
		config, err = jsonc.Edit(config, refreshedToken.Expiry.Unix(), "token.oauth.expiry")
		if err != nil {
			return "", errors.Wrap(err, "update OAuth token expiry")
		}
		extsvc.Config.Set(config)

		extsvc.UpdatedAt = time.Now()
		if err := db.ExternalServices().Upsert(ctx, extsvc); err != nil {
			return "", errors.Wrap(err, "upsert external service")
		}
		return refreshedToken.AccessToken, nil
	}
}
