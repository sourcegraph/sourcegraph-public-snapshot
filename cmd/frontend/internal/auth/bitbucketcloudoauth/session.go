package bitbucketcloudoauth

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/oauth2"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/hubspot"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/hubspot/hubspotutil"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth/oauth"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/auth/providers"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	esauth "github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketcloud"
	"github.com/sourcegraph/sourcegraph/internal/session"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

type sessionIssuerHelper struct {
	baseURL     *url.URL
	clientKey   string
	db          database.DB
	allowSignup bool
	client      bitbucketcloud.Client
}

func (s *sessionIssuerHelper) AuthSucceededEventName() database.SecurityEventName {
	return database.SecurityEventBitbucketCloudAuthSucceeded
}

func (s *sessionIssuerHelper) AuthFailedEventName() database.SecurityEventName {
	return database.SecurityEventBitbucketCloudAuthFailed
}

func (s *sessionIssuerHelper) GetServiceID() string {
	return s.baseURL.String()
}

func (s *sessionIssuerHelper) GetOrCreateUser(ctx context.Context, token *oauth2.Token, anonymousUserID, firstSourceURL, lastSourceURL string) (newUserCreated bool, actr *actor.Actor, safeErrMsg string, err error) {
	var client bitbucketcloud.Client
	if s.client != nil {
		client = s.client
	} else {
		conf := &schema.BitbucketCloudConnection{
			Url: s.baseURL.String(),
		}
		client, err = bitbucketcloud.NewClient(s.baseURL.String(), conf, nil)
		if err != nil {
			return false, nil, "Could not initialize Bitbucket Cloud client", err
		}
	}

	// The token used here is fresh from Bitbucket OAuth. It should be valid
	// for 1 hour, so we don't bother with setting up token refreshing yet.
	// If account creation/linking succeeds, the token will be stored in the
	// database with the refresh token, and refreshing can happen from that point.
	auther := &esauth.OAuthBearerToken{Token: token.AccessToken}
	client = client.WithAuthenticator(auther)
	bbUser, err := client.CurrentUser(ctx)
	if err != nil {
		return false, nil, "Could not read Bitbucket user from callback request.", errors.Wrap(err, "could not read user from bitbucket")
	}

	var data extsvc.AccountData
	if err := bitbucketcloud.SetExternalAccountData(&data, &bbUser.Account, token); err != nil {
		return false, nil, "", err
	}

	emails, err := client.AllCurrentUserEmails(ctx)
	if err != nil {
		return false, nil, "", err
	}

	attempts, err := buildUserFetchAttempts(emails, s.allowSignup)
	if err != nil {
		return false, nil, "Could not find verified email address for Bitbucket user.", err
	}

	var (
		firstSafeErrMsg string
		firstErr        error
	)

	for i, attempt := range attempts {
		newUserCreated, userID, safeErrMsg, err := auth.GetAndSaveUser(ctx, s.db, auth.GetAndSaveUserOp{
			UserProps: database.NewUser{
				Username:        bbUser.Username,
				Email:           attempt.email,
				EmailIsVerified: true,
				DisplayName:     bbUser.Nickname,
				AvatarURL:       bbUser.Links["avatar"].Href,
			},
			ExternalAccount: extsvc.AccountSpec{
				ServiceType: extsvc.TypeBitbucketCloud,
				ServiceID:   s.baseURL.String(),
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
			return newUserCreated, actor.FromUser(userID), "", nil
		}
		if i == 0 {
			firstSafeErrMsg, firstErr = safeErrMsg, err
		}
	}

	// On failure, return the first error
	verifiedEmails := make([]string, 0, len(attempts))
	for i, attempt := range attempts {
		verifiedEmails[i] = attempt.email
	}
	return false, nil, fmt.Sprintf("No Sourcegraph user exists matching any of the verified emails: %s.\n\nFirst error was: %s", strings.Join(verifiedEmails, ", "), firstSafeErrMsg), firstErr
}

type attempt struct {
	email            string
	createIfNotExist bool
}

func buildUserFetchAttempts(emails []*bitbucketcloud.UserEmail, allowSignup bool) ([]attempt, error) {
	attempts := []attempt{}
	for _, email := range emails {
		if email.IsConfirmed {
			attempts = append(attempts, attempt{
				email:            email.Email,
				createIfNotExist: false,
			})
		}
	}
	if len(attempts) == 0 {
		return nil, errors.New("no verified email")
	}
	// If allowSignup is true, we will create an account using the first verified
	// email address from Bitbucket which we expect to be their primary address. Note
	// that the order of attempts is important. If we manage to connect with an
	// existing account we return early and don't attempt to create a new account.
	if allowSignup {
		attempts = append(attempts, attempt{
			email:            attempts[0].email,
			createIfNotExist: true,
		})
	}

	return attempts, nil
}

func (s *sessionIssuerHelper) DeleteStateCookie(w http.ResponseWriter, r *http.Request) {
	session.SetData(w, r, "oauthState", "")
}

func (s *sessionIssuerHelper) SessionData(token *oauth2.Token) oauth.SessionData {
	return oauth.SessionData{
		ID: providers.ConfigID{
			ID:   s.baseURL.String(),
			Type: extsvc.TypeBitbucketCloud,
		},
		AccessToken: token.AccessToken,
		TokenType:   token.Type(),
		// TODO(pjlast): investigate exactly where and how we use this SessionData
	}
}
