package bitbucketcloudoauth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/oauth2"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth/providers"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/hubspot"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/hubspot/hubspotutil"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/auth/oauth"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	esauth "github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketcloud"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

type sessionIssuerHelper struct {
	*extsvc.CodeHost
	clientKey   string
	db          database.DB
	allowSignup bool
}

func (s *sessionIssuerHelper) AuthSucceededEventName() database.SecurityEventName {
	return database.SecurityEventBitbucketCloudAuthSucceeded
}

func (s *sessionIssuerHelper) AuthFailedEventName() database.SecurityEventName {
	return database.SecurityEventBitbucketCloudAuthFailed
}

func (s *sessionIssuerHelper) GetOrCreateUser(ctx context.Context, token *oauth2.Token, anonymousUserID, firstSourceURL, lastSourceURL string) (actr *actor.Actor, safeErrMsg string, err error) {
	conf := &schema.BitbucketCloudConnection{
		Url:    s.BaseURL.String(),
		ApiURL: s.BaseURL.String(),
	}
	bbClient, err := bitbucketcloud.NewClient(s.BaseURL.String(), conf, nil)
	if err != nil {
		return nil, "Could not initialize Bitbucket Cloud client", err
	}

	auther := &esauth.OAuthBearerToken{Token: token.AccessToken, RefreshToken: token.RefreshToken, Expiry: token.Expiry}
	bbClient = bbClient.WithAuthenticator(auther)
	bbUser, err := bbClient.CurrentUser(ctx)
	if err != nil {
		return nil, "Could not read Bitbucket user from callback request.", errors.Wrap(err, "could not read user from bitbucket")
	}

	var data extsvc.AccountData
	if err := bitbucketcloud.SetExternalAccountData(&data, &bbUser.Account, token); err != nil {
		return nil, "", err
	}

	emails, err := bbClient.CurrentUserEmails(ctx)
	if err != nil {
		return nil, "", err
	}

	type attemptConfig struct {
		email            string
		createIfNotExist bool
	}
	attempts := []attemptConfig{}
	verifiedEmails := []string{}
	for _, email := range emails {
		if email.IsConfirmed {
			attempts = append(attempts, attemptConfig{
				email:            email.Email,
				createIfNotExist: false,
			})
			verifiedEmails = append(verifiedEmails, email.Email)
		}
	}
	if len(verifiedEmails) == 0 {
		return nil, "Could not find verified email address for Bitbucket user.", errors.New("no verified email")
	}
	// If allowSignup is true, we will create an account using the first verified
	// email address from Bitbucket which we expect to be their primary address. Note
	// that the order of attempts is important. If we manage to connect with an
	// existing account we return early and don't attempt to create a new account.
	if s.allowSignup {
		attempts = append(attempts, attemptConfig{
			email:            emails[0].Email,
			createIfNotExist: true,
		})
	}

	var (
		firstSafeErrMsg string
		firstErr        error
	)

	for i, attempt := range attempts {
		userID, safeErrMsg, err := auth.GetAndSaveUser(ctx, s.db, auth.GetAndSaveUserOp{
			UserProps: database.NewUser{
				Username:        bbUser.Username,
				Email:           attempt.email,
				EmailIsVerified: true,
				DisplayName:     bbUser.Nickname,
				AvatarURL:       "",
			},
			ExternalAccount: extsvc.AccountSpec{
				ServiceType: s.ServiceType,
				ServiceID:   s.ServiceID,
				ClientID:    s.clientKey,
				AccountID:   bbUser.UUID,
			},
			ExternalAccountData: data,
			CreateIfNotExist:    attempt.createIfNotExist,
		})
		if err == nil {
			go hubspotutil.SyncUser(attempt.email, hubspotutil.SignupEventID, &hubspot.ContactProperties{
				AnonymousUserID: anonymousUserID,
				FirstSourceURL:  firstSourceURL,
				LastSourceURL:   lastSourceURL,
			})
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
		ID: providers.ConfigID{
			ID:   s.ServiceID,
			Type: s.ServiceType,
		},
		AccessToken: token.AccessToken,
		TokenType:   token.Type(),
		// TODO(beyang): store and use refresh token to auto-refresh sessions
	}
}
