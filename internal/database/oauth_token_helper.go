package database

import (
	"context"
	"strconv"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/oauthutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"golang.org/x/oauth2"
)

// externalAccountTokenRefresher returns an oauthutil.TokenRefresher for the
// given external account.
func externalAccountTokenRefresher(store UserExternalAccountsStore, externalAccountID int32, originalToken *auth.OAuthBearerToken) oauthutil.TokenRefresher {
	return func(ctx context.Context, doer httpcli.Doer, oauthCtx oauthutil.OAuthContext) (token *auth.OAuthBearerToken, err error) {
		// Start a transaction so that multiple refreshes don't happen simultaneously
		tx, err := store.Transact(ctx)
		if err != nil {
			return nil, err
		}
		defer func() { err = tx.Done(err) }()

		// Read the token from the DB again, in case it has been refreshed in the mean time
		acct, err := store.Get(ctx, externalAccountID)
		if err != nil {
			return nil, err
		}
		if acct.AuthData == nil {
			return nil, errors.Newf("no auth data found for external account id %d", externalAccountID)
		}
		tok, err := encryption.DecryptJSON[oauth2.Token](ctx, acct.AuthData)
		if err != nil {
			return nil, err
		}
		fetchedToken := &auth.OAuthBearerToken{
			Token:        tok.AccessToken,
			RefreshToken: tok.RefreshToken,
			Expiry:       tok.Expiry,
		}
		// Compare the stored token with the provided one.
		// If they differ, the token was most likely refreshed in the meantime.
		// Check `NeedsRefresh` for good measure.
		if fetchedToken.Token != originalToken.Token && !fetchedToken.NeedsRefresh() {
			return fetchedToken, nil
		}

		// Otherwise, do the token refresh
		refreshedToken, err := oauthutil.RetrieveToken(doer, oauthCtx, fetchedToken.RefreshToken, oauthutil.AuthStyleInParams)
		if err != nil {
			return nil, errors.Wrap(err, "refresh token")
		}

		// Store the refreshed token
		err = acct.AuthData.Set(refreshedToken)
		if err != nil {
			return nil, errors.Wrap(err, "set auth data")
		}
		_, err = store.LookupUserAndSave(ctx, acct.AccountSpec, acct.AccountData)
		if err != nil {
			return nil, errors.Wrap(err, "save refreshed token")
		}

		return &auth.OAuthBearerToken{
			Token:        refreshedToken.AccessToken,
			RefreshToken: refreshedToken.RefreshToken,
			Expiry:       refreshedToken.Expiry,
		}, nil
	}
}

// externalServiceTokenRefresher returns an oauthutil.TokenRefresher for the
// given external service.
func externalServiceTokenRefresher(db DB, externalServiceID int64, refreshToken string) oauthutil.TokenRefresher {
	return func(ctx context.Context, doer httpcli.Doer, oauthCtx oauthutil.OAuthContext) (token *auth.OAuthBearerToken, err error) {
		defer func() {
			success := err == nil
			gitlab.TokenRefreshCounter.WithLabelValues("codehost", strconv.FormatBool(success)).Inc()
		}()

		refreshedToken, err := oauthutil.RetrieveToken(doer, oauthCtx, refreshToken, oauthutil.AuthStyleInParams)
		if err != nil {
			return nil, errors.Wrap(err, "refresh token")
		}

		oauthBearerToken := &auth.OAuthBearerToken{
			Token:        refreshedToken.AccessToken,
			RefreshToken: refreshedToken.RefreshToken,
			Expiry:       refreshedToken.Expiry,
		}

		extsvc, err := db.ExternalServices().GetByID(ctx, externalServiceID)
		if err != nil {
			return nil, errors.Wrap(err, "get external service")
		}

		config, err := extsvc.Config.Decrypt(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "decrypt config")
		}

		config, err = jsonc.Edit(config, oauthBearerToken.Token, "token")
		if err != nil {
			return nil, errors.Wrap(err, "update OAuth token")
		}
		config, err = jsonc.Edit(config, refreshedToken.RefreshToken, "token.oauth.refresh")
		if err != nil {
			return nil, errors.Wrap(err, "update OAuth refresh token")
		}
		config, err = jsonc.Edit(config, oauthBearerToken.Expiry.Unix(), "token.oauth.expiry")
		if err != nil {
			return nil, errors.Wrap(err, "update OAuth token expiry")
		}
		extsvc.Config.Set(config)

		extsvc.UpdatedAt = time.Now()
		if err := db.ExternalServices().Upsert(ctx, extsvc); err != nil {
			return nil, errors.Wrap(err, "upsert external service")
		}
		return oauthBearerToken, nil
	}
}

func GetServiceRefreshAndStoreOAuthTokenFunc(db DB, externalServiceID int64, oauthContext *oauthutil.OAuthContext) func(context.Context, httpcli.Doer, *auth.OAuthBearerToken) (string, string, time.Time, error) {
	return func(ctx context.Context, cli httpcli.Doer, a *auth.OAuthBearerToken) (string, string, time.Time, error) {
		tokenRefresher := externalServiceTokenRefresher(db, externalServiceID, a.RefreshToken)
		token, err := tokenRefresher(ctx, cli, *oauthContext)
		if err != nil {
			return "", "", time.Time{}, err
		}

		return token.Token, token.RefreshToken, token.Expiry, nil
	}
}

func GetAccountRefreshAndStoreOAuthTokenFunc(store UserExternalAccountsStore, externalAccountID int32, oauthContext *oauthutil.OAuthContext) func(context.Context, httpcli.Doer, *auth.OAuthBearerToken) (string, string, time.Time, error) {
	return func(ctx context.Context, cli httpcli.Doer, a *auth.OAuthBearerToken) (string, string, time.Time, error) {
		tokenRefresher := externalAccountTokenRefresher(store, externalAccountID, a)
		token, err := tokenRefresher(ctx, cli, *oauthContext)
		if err != nil {
			return "", "", time.Time{}, err
		}

		return token.Token, token.RefreshToken, token.Expiry, nil
	}
}
