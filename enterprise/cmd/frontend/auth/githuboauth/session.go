package githuboauth

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/dghubble/gologin/github"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/auth/oauth"
	"github.com/sourcegraph/sourcegraph/pkg/actor"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc"
	githubsvc "github.com/sourcegraph/sourcegraph/pkg/extsvc/github"
	"golang.org/x/oauth2"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

type sessionIssuerHelper struct {
	*githubsvc.CodeHost
	clientID    string
	allowSignup bool
}

func (s *sessionIssuerHelper) GetOrCreateUser(ctx context.Context, token *oauth2.Token) (actr *actor.Actor, safeErrMsg string, err error) {
	ghUser, err := github.UserFromContext(ctx)
	if ghUser == nil {
		if err != nil {
			err = errors.Wrap(err, "could not read user from context")
		} else {
			err = errors.New("could not read user from context")
		}
		return nil, "Could not read GitHub user from callback request.", err
	}

	login, err := auth.NormalizeUsername(deref(ghUser.Login))
	if err != nil {
		return nil, fmt.Sprintf("Error normalizing the username %q. See https://docs.sourcegraph.com/admin/auth/#username-normalization.", login), err
	}

	// 🚨 SECURITY: Ensure that the user email is verified
	verifiedEmail := s.getVerifiedPrimaryEmail(ctx, token)
	if verifiedEmail == "" {
		return nil, "Could not get verified primary email for GitHub user. Check that your GitHub account's primary email is verified.", errors.New("no verified primary email")
	}

	var data extsvc.ExternalAccountData
	data.SetAccountData(ghUser)
	data.SetAuthData(token)
	userID, safeErrMsg, err := auth.GetAndSaveUser(ctx, auth.GetAndSaveUserOp{
		UserProps: db.NewUser{
			Username:        login,
			Email:           verifiedEmail,
			EmailIsVerified: true,
			DisplayName:     deref(ghUser.Name),
			AvatarURL:       deref(ghUser.AvatarURL),
		},
		ExternalAccount: extsvc.ExternalAccountSpec{
			ServiceType: s.ServiceType(),
			ServiceID:   s.ServiceID(),
			ClientID:    s.clientID,
			AccountID:   strconv.FormatInt(derefInt64(ghUser.ID), 10),
		},
		ExternalAccountData: data,
		CreateIfNotExist:    s.allowSignup,
	})
	if err != nil {
		return nil, safeErrMsg, err
	}
	return actor.FromUser(userID), "", nil

}

func (s *sessionIssuerHelper) DeleteStateCookie(w http.ResponseWriter) {
	stateConfig := getStateConfig()
	stateConfig.MaxAge = -1
	http.SetCookie(w, oauth.NewCookie(stateConfig, ""))
}

func (s *sessionIssuerHelper) SessionData(token *oauth2.Token) oauth.SessionData {
	return oauth.SessionData{
		ID: auth.ProviderConfigID{
			ID:   s.ServiceID(),
			Type: s.ServiceType(),
		},
		AccessToken: token.AccessToken,
		TokenType:   token.Type(),
		// TODO(beyang): store and use refresh token to auto-refresh sessions
	}
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func derefInt64(i *int64) int64 {
	if i == nil {
		return 0
	}
	return *i
}

// getVerifiedPrimaryEmail returns the primary email of the user if it is verified. If the user has
// no emails or if the primary email is not verified, it returns the empty string.
func (s *sessionIssuerHelper) getVerifiedPrimaryEmail(ctx context.Context, token *oauth2.Token) string {
	apiURL, _ := githubsvc.APIRoot(s.BaseURL())
	ghClient := githubsvc.NewClient(apiURL, "", nil)
	emails, err := ghClient.GetAuthenticatedUserEmails(ctx, token.AccessToken)
	if err != nil {
		log15.Warn("Could not get GitHub authenticated user emails", "error", err)
		return ""
	}
	for _, email := range emails {
		if email.Primary && email.Verified {
			return email.Email
		}
	}
	return ""
}
