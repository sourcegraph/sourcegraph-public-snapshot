package bitbucketserveroauth

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/hubspot"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/hubspot/hubspotutil"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth/oauth"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth/providers"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth/session"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	esauth "github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/telemetry/telemetryrecorder"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
	"golang.org/x/oauth2"
)

type sessionIssuerHelper struct {
	logger      log.Logger
	baseURL     *url.URL
	clientKey   string
	db          database.DB
	allowSignup bool
	client      *bitbucketserver.Client
}

func (s *sessionIssuerHelper) AuthSucceededEventName() database.SecurityEventName {
	return database.SecurityEventBitbucketServerAuthSucceeded
}

func (s *sessionIssuerHelper) AuthFailedEventName() database.SecurityEventName {
	return database.SecurityEventBitbucketServerAuthFailed
}

func (s *sessionIssuerHelper) DeleteStateCookie(w http.ResponseWriter, r *http.Request) {
	session.SetData(w, r, "oauthState", "")
}

func (s *sessionIssuerHelper) GetServiceID() string {
	return s.baseURL.String()
}

func (s *sessionIssuerHelper) GetOrCreateUser(ctx context.Context, token *oauth2.Token, hubSpotProps *hubspot.ContactProperties) (newUserCreated bool, actr *actor.Actor, safeErrMsg string, err error) {
	var client *bitbucketserver.Client
	if s.client != nil {
		client = s.client
	} else {
		conf := &schema.BitbucketServerConnection{
			Url: s.baseURL.String(),
		}
		client, err = bitbucketserver.NewClient(s.baseURL.String(), conf, nil)
		if err != nil {
			return false, nil, "Could not initialize Bitbucket Server client", err
		}
	}

	// The token used here is fresh from Bitbucket OAuth. It should be valid
	// for 1 hour, so we don't bother with setting up token refreshing yet.
	// If account creation/linking succeeds, the token will be stored in the
	// database with the refresh token, and refreshing can happen from that point.
	auther := &esauth.OAuthBearerToken{Token: token.AccessToken}
	client = client.WithAuthenticator(auther)
	username, err := client.AuthenticatedUsername(ctx)
	if err != nil {
		return false, nil, "Could not read Bitbucket user from callback request.", errors.Wrap(err, "could not read user from bitbucket")
	}

	bbUser := bitbucketserver.User{
		Slug: username,
	}

	err = client.LoadUser(ctx, &bbUser)
	if err != nil {
		return false, nil, "Could not read Bitbucket user from callback request.", errors.Wrap(err, "could not read user from bitbucket")
	}

	var data extsvc.AccountData
	if err := bitbucketserver.SetExternalAccountData(&data, &bbUser, token); err != nil {
		return false, nil, "", err
	}

	recorder := telemetryrecorder.New(s.db)
	newUserCreated, userID, safeErrMsg, err := auth.GetAndSaveUser(ctx, s.logger, s.db, recorder, auth.GetAndSaveUserOp{
		UserProps: database.NewUser{
			Username:        bbUser.Name,
			Email:           bbUser.EmailAddress,
			EmailIsVerified: true,
			DisplayName:     bbUser.DisplayName,
		},
		ExternalAccount: extsvc.AccountSpec{
			ServiceType: extsvc.TypeBitbucketServer,
			ServiceID:   s.baseURL.String(),
			ClientID:    s.clientKey,
			AccountID:   strconv.Itoa(bbUser.ID),
		},
		ExternalAccountData: data,
		CreateIfNotExist:    s.allowSignup,
	})
	if err != nil {
		return false, nil, fmt.Sprintf("No Sourcegraph user exists matching the email: %s.\n\nError was: %s", bbUser.EmailAddress, safeErrMsg), err
	}

	go hubspotutil.SyncUser(bbUser.EmailAddress, hubspotutil.SignupEventID, hubSpotProps)
	return newUserCreated, actor.FromUser(userID), "", nil
}

func (s *sessionIssuerHelper) SessionData(token *oauth2.Token) oauth.SessionData {
	return oauth.SessionData{
		ID: providers.ConfigID{
			ID:   s.baseURL.String(),
			Type: extsvc.TypeBitbucketServer,
		},
		AccessToken: token.AccessToken,
		TokenType:   token.Type(),
	}
}
