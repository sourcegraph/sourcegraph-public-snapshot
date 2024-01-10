package oauthtoken

import (
	"context"
	"time"

	"golang.org/x/oauth2"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/oauthutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// externalAccountTokenRefresher returns an oauthutil.TokenRefresher for the
// given external account.
func externalAccountTokenRefresher(store database.UserExternalAccountsStore, externalAccountID int32, originalToken *auth.OAuthBearerToken) oauthutil.TokenRefresher {
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
		_, err = store.Update(ctx, acct)
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

func GetAccountRefreshAndStoreOAuthTokenFunc(store database.UserExternalAccountsStore, externalAccountID int32, oauthContext *oauthutil.OAuthContext) func(context.Context, httpcli.Doer, *auth.OAuthBearerToken) (string, string, time.Time, error) {
	return func(ctx context.Context, cli httpcli.Doer, a *auth.OAuthBearerToken) (string, string, time.Time, error) {
		tokenRefresher := externalAccountTokenRefresher(store, externalAccountID, a)
		token, err := tokenRefresher(ctx, cli, *oauthContext)
		if err != nil {
			return "", "", time.Time{}, err
		}

		return token.Token, token.RefreshToken, token.Expiry, nil
	}
}
