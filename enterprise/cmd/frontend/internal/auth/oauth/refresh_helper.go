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

	refreshedToken, err := s.Config.TokenSource(ctx, s.Token).Token()
	if err != nil {
		return "", errors.Wrap(err, "refresh token")
	}

	/// todo - remove/replace hardcoded accts
	if refreshedToken.AccessToken != s.Token.AccessToken {
		defer func() {
			success := err == nil
			gitlab.TokenRefreshCounter.WithLabelValues("external_account", strconv.FormatBool(success)).Inc()
		}()
		accts[0].AccountData.SetAuthData(refreshedToken)                                                         // todo
		_, err := s.DB.UserExternalAccounts().LookupUserAndSave(ctx, accts[0].AccountSpec, accts[0].AccountData) // todo
		if err != nil {
			return "", errors.Wrap(err, "save refreshed token")
		}
	}

	return "", nil
}
