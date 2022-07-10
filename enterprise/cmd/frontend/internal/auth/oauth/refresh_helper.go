package oauth

import (
	"context"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"golang.org/x/oauth2"
	"strconv"
)

type RefreshTokenHelper struct {
	DB          database.DB
	Config      *oauth2.Config
	Token       *oauth2.Token
	ServiceType string
}

func (s *RefreshTokenHelper) RefreshToken(ctx context.Context, doer httpcli.Doer) (string, error) {
	userID := actor.FromContext(ctx).UID
	accts, err := s.DB.UserExternalAccounts().List(ctx,
		database.ExternalAccountsListOptions{
			UserID:         userID,
			ExcludeExpired: true,
			ServiceType:    s.ServiceType,
		},
	)
	if err != nil {
		return "", errors.Wrap(err, "list external accounts")
	}

	// todo: check/test this is fine
	//if accts[0].ServiceType != extsvc.TypeGitLab { // todo
	//	return "", err
	//}

	//var oauthConfig *oauth2.Config
	//expiryWindow := 10 * time.Minute
	//for _, authProvider := range conf.SiteConfig().AuthProviders {
	//	if authProvider.Gitlab == nil ||
	//		strings.TrimSuffix(accts[0].ServiceID, "/") != strings.TrimSuffix(authProvider.Gitlab.Url, "/") {
	//		continue
	//	}
	//	oauthConfig = oauth2ConfigFromGitLabProvider(authProvider.Gitlab)
	//	if authProvider.Gitlab.TokenRefreshWindowMinutes > 0 {
	//		expiryWindow = time.Duration(authProvider.Gitlab.TokenRefreshWindowMinutes) * time.Minute
	//	}
	//	break
	//}
	//if oauthConfig == nil {
	//	//logger.Warn("external account has no auth.provider") // todo
	//	return "", nil
	//}
	//
	//_, tok, err := gitlab.GetExternalAccountData(&accts[0].AccountData) // todo
	//if err != nil {
	//	return "", errors.Wrap(err, "get external account data")
	//} else if tok == nil {
	//	return "", errors.New("no token found in the external account data")
	//}
	//
	//tok.Expiry = tok.Expiry.Add(expiryWindow)

	//refreshedToken, err := oauthConfig.TokenSource(ctx, tok).Token()
	//if err != nil {
	//	return "", errors.Wrap(err, "refresh token")
	//}
	//
	//if refreshedToken.AccessToken != tok.AccessToken {
	//	defer func() {
	//		success := err == nil
	//		gitlab.TokenRefreshCounter.WithLabelValues("external_account", strconv.FormatBool(success)).Inc()
	//	}()
	//	accts[0].AccountData.SetAuthData(refreshedToken)
	//	_, err := s.DB.UserExternalAccounts().LookupUserAndSave(ctx, accts[0].AccountSpec, accts[0].AccountData)
	//	if err != nil {
	//		return "", errors.Wrap(err, "save refreshed token")
	//	}
	//}

	refreshedToken, err := s.Config.TokenSource(ctx, s.Token).Token()
	if err != nil {
		return "", errors.Wrap(err, "refresh token")
	}

	/// todo -
	if refreshedToken.AccessToken != s.Token.AccessToken {
		defer func() {
			success := err == nil
			gitlab.TokenRefreshCounter.WithLabelValues("external_account", strconv.FormatBool(success)).Inc()
		}()
		accts[0].AccountData.SetAuthData(refreshedToken)
		_, err := s.DB.UserExternalAccounts().LookupUserAndSave(ctx, accts[0].AccountSpec, accts[0].AccountData)
		if err != nil {
			return "", errors.Wrap(err, "save refreshed token")
		}
	}

	return "", nil
}
