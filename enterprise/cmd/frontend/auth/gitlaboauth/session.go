package gitlaboauth

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/auth/oauth"
	"github.com/sourcegraph/sourcegraph/pkg/actor"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/gitlab"
	"golang.org/x/oauth2"
)

type sessionIssuerHelper struct {
	*gitlab.CodeHost
	clientID string
}

func (s *sessionIssuerHelper) GetOrCreateUser(ctx context.Context, token *oauth2.Token) (actr *actor.Actor, safeErrMsg string, err error) {
	gUser, err := UserFromContext(ctx)
	if err != nil {
		return nil, "Could not read GitLab user from callback request.", errors.Wrap(err, "could not read user from context")
	}

	login, err := auth.NormalizeUsername(gUser.Username)
	if err != nil {
		return nil, fmt.Sprintf("Error normalizing the username %q. See https://docs.sourcegraph.com/admin/auth/#username-normalization.", login), err
	}

	var data extsvc.ExternalAccountData
	gitlab.SetExternalAccountData(&data, gUser, token)

	verifiedEmails, err := s.getVerifiedEmails(ctx, token)
	if err != nil || len(verifiedEmails) == 0 {
		if err == nil {
			err = errors.New("no verified email")
		}
		return nil, "Could not get verified email for GitLab user. Check that your GitLab account has a verified email that matches one of your Sourcegraph verified emails.", err
	}

	var (
		firstSafeErrMsg string
		firstErr        error
	)
	// for i, verifiedEmail := range verifiedEmails {
	for i, verifiedEmail := range []string{gUser.Email} {
		log.Printf("# considering verified email: %q", verifiedEmail)
		userID, safeErrMsg, err := auth.GetAndSaveUser(ctx, auth.GetAndSaveUserOp{
			UserProps: db.NewUser{
				Username:        login,
				Email:           verifiedEmail,
				EmailIsVerified: true,
				DisplayName:     gUser.Name,
				AvatarURL:       gUser.AvatarURL,
			},
			ExternalAccount: extsvc.ExternalAccountSpec{
				ServiceType: s.ServiceType(),
				ServiceID:   s.ServiceID(),
				ClientID:    s.clientID,
				AccountID:   strconv.FormatInt(int64(gUser.ID), 10),
			},
			ExternalAccountData: data,
			CreateIfNotExist:    true,
		})
		if err == nil {
			return actor.FromUser(userID), "", nil
		}
		if i == 0 {
			firstSafeErrMsg, firstErr = safeErrMsg, err
		}
	}
	// On failure, return the first error
	return nil, fmt.Sprintf("No user exists matching any of the verified emails: %s.\n\nFirst error was: %s", strings.Join(verifiedEmails, ", "), firstSafeErrMsg), firstErr
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

func SignOutURL(gitlabURL string) (string, error) {
	if gitlabURL == "" {
		gitlabURL = "https://gitlab.com"
	}
	ghURL, err := url.Parse(gitlabURL)
	if err != nil {
		return "", err
	}
	ghURL.Path = path.Join(ghURL.Path, "users/sign_out")
	return ghURL.String(), nil
}

func (s *sessionIssuerHelper) getVerifiedEmails(ctx context.Context, token *oauth2.Token) (verifiedEmails []string, err error) {
	c := gitlab.NewClientProvider(s.BaseURL(), nil).GetOAuthClient(token.AccessToken)
	emails, err := c.ListEmails(ctx)
	if err != nil {
		return nil, err
	}
	emailStrs := make([]string, 0, len(emails))
	for _, e := range emails {
		emailStrs = append(emailStrs, e.Email)
	}
	return emailStrs, nil
}
