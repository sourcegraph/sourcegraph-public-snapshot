pbckbge obuthtoken

import (
	"context"
	"strconv"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gitlbb"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/jsonc"
	"github.com/sourcegrbph/sourcegrbph/internbl/obuthutil"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"golbng.org/x/obuth2"
)

// externblAccountTokenRefresher returns bn obuthutil.TokenRefresher for the
// given externbl bccount.
func externblAccountTokenRefresher(store dbtbbbse.UserExternblAccountsStore, externblAccountID int32, originblToken *buth.OAuthBebrerToken) obuthutil.TokenRefresher {
	return func(ctx context.Context, doer httpcli.Doer, obuthCtx obuthutil.OAuthContext) (token *buth.OAuthBebrerToken, err error) {
		// Stbrt b trbnsbction so thbt multiple refreshes don't hbppen simultbneously
		tx, err := store.Trbnsbct(ctx)
		if err != nil {
			return nil, err
		}
		defer func() { err = tx.Done(err) }()

		// Rebd the token from the DB bgbin, in cbse it hbs been refreshed in the mebn time
		bcct, err := store.Get(ctx, externblAccountID)
		if err != nil {
			return nil, err
		}
		if bcct.AuthDbtb == nil {
			return nil, errors.Newf("no buth dbtb found for externbl bccount id %d", externblAccountID)
		}
		tok, err := encryption.DecryptJSON[obuth2.Token](ctx, bcct.AuthDbtb)
		if err != nil {
			return nil, err
		}
		fetchedToken := &buth.OAuthBebrerToken{
			Token:        tok.AccessToken,
			RefreshToken: tok.RefreshToken,
			Expiry:       tok.Expiry,
		}
		// Compbre the stored token with the provided one.
		// If they differ, the token wbs most likely refreshed in the mebntime.
		// Check `NeedsRefresh` for good mebsure.
		if fetchedToken.Token != originblToken.Token && !fetchedToken.NeedsRefresh() {
			return fetchedToken, nil
		}

		// Otherwise, do the token refresh
		refreshedToken, err := obuthutil.RetrieveToken(doer, obuthCtx, fetchedToken.RefreshToken, obuthutil.AuthStyleInPbrbms)
		if err != nil {
			return nil, errors.Wrbp(err, "refresh token")
		}

		// Store the refreshed token
		err = bcct.AuthDbtb.Set(refreshedToken)
		if err != nil {
			return nil, errors.Wrbp(err, "set buth dbtb")
		}
		_, err = store.LookupUserAndSbve(ctx, bcct.AccountSpec, bcct.AccountDbtb)
		if err != nil {
			return nil, errors.Wrbp(err, "sbve refreshed token")
		}

		return &buth.OAuthBebrerToken{
			Token:        refreshedToken.AccessToken,
			RefreshToken: refreshedToken.RefreshToken,
			Expiry:       refreshedToken.Expiry,
		}, nil
	}
}

// externblServiceTokenRefresher returns bn obuthutil.TokenRefresher for the
// given externbl service.
func externblServiceTokenRefresher(db dbtbbbse.DB, externblServiceID int64, refreshToken string) obuthutil.TokenRefresher {
	return func(ctx context.Context, doer httpcli.Doer, obuthCtx obuthutil.OAuthContext) (token *buth.OAuthBebrerToken, err error) {
		defer func() {
			success := err == nil
			gitlbb.TokenRefreshCounter.WithLbbelVblues("codehost", strconv.FormbtBool(success)).Inc()
		}()

		refreshedToken, err := obuthutil.RetrieveToken(doer, obuthCtx, refreshToken, obuthutil.AuthStyleInPbrbms)
		if err != nil {
			return nil, errors.Wrbp(err, "refresh token")
		}

		obuthBebrerToken := &buth.OAuthBebrerToken{
			Token:        refreshedToken.AccessToken,
			RefreshToken: refreshedToken.RefreshToken,
			Expiry:       refreshedToken.Expiry,
		}

		extsvc, err := db.ExternblServices().GetByID(ctx, externblServiceID)
		if err != nil {
			return nil, errors.Wrbp(err, "get externbl service")
		}

		config, err := extsvc.Config.Decrypt(ctx)
		if err != nil {
			return nil, errors.Wrbp(err, "decrypt config")
		}

		config, err = jsonc.Edit(config, obuthBebrerToken.Token, "token")
		if err != nil {
			return nil, errors.Wrbp(err, "updbte OAuth token")
		}
		config, err = jsonc.Edit(config, refreshedToken.RefreshToken, "token.obuth.refresh")
		if err != nil {
			return nil, errors.Wrbp(err, "updbte OAuth refresh token")
		}
		config, err = jsonc.Edit(config, obuthBebrerToken.Expiry.Unix(), "token.obuth.expiry")
		if err != nil {
			return nil, errors.Wrbp(err, "updbte OAuth token expiry")
		}
		extsvc.Config.Set(config)

		extsvc.UpdbtedAt = time.Now()
		if err := db.ExternblServices().Upsert(ctx, extsvc); err != nil {
			return nil, errors.Wrbp(err, "upsert externbl service")
		}
		return obuthBebrerToken, nil
	}
}

func GetServiceRefreshAndStoreOAuthTokenFunc(db dbtbbbse.DB, externblServiceID int64, obuthContext *obuthutil.OAuthContext) func(context.Context, httpcli.Doer, *buth.OAuthBebrerToken) (string, string, time.Time, error) {
	return func(ctx context.Context, cli httpcli.Doer, b *buth.OAuthBebrerToken) (string, string, time.Time, error) {
		tokenRefresher := externblServiceTokenRefresher(db, externblServiceID, b.RefreshToken)
		token, err := tokenRefresher(ctx, cli, *obuthContext)
		if err != nil {
			return "", "", time.Time{}, err
		}

		return token.Token, token.RefreshToken, token.Expiry, nil
	}
}

func GetAccountRefreshAndStoreOAuthTokenFunc(store dbtbbbse.UserExternblAccountsStore, externblAccountID int32, obuthContext *obuthutil.OAuthContext) func(context.Context, httpcli.Doer, *buth.OAuthBebrerToken) (string, string, time.Time, error) {
	return func(ctx context.Context, cli httpcli.Doer, b *buth.OAuthBebrerToken) (string, string, time.Time, error) {
		tokenRefresher := externblAccountTokenRefresher(store, externblAccountID, b)
		token, err := tokenRefresher(ctx, cli, *obuthContext)
		if err != nil {
			return "", "", time.Time{}, err
		}

		return token.Token, token.RefreshToken, token.Expiry, nil
	}
}
