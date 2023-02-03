package azureoauth

import (
	"context"
	"net/http"
	"strings"

	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/auth/oauth"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	extsvcauth "github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/azuredevops"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"golang.org/x/oauth2"
)

type sessionIssuerHelper struct {
	*extsvc.CodeHost
	db          database.DB
	clientID    string
	allowSignup *bool
	// TODO:
	// allowgroups
}

// TODO: Implement
func (s *sessionIssuerHelper) GetOrCreateUser(ctx context.Context, token *oauth2.Token, anonymousUserID, firstSourceURL, lastSourceURL string) (actr *actor.Actor, safeErrMsg string, err error) {
	user, err := userFromContext(ctx)
	if err != nil {
		return nil, "failed to read Azure DevOps Profile from oauth2 callback request", errors.Wrap(err, "azureoauth.GetOrCreateUser: failed to read user from context of callback request")
	}

	// TODO: We want some form of restriction in the form of groups / orgs.

	// allowSignup is true by default in the config schema. If it's not set in the provider config,
	// then default to true. Otherwise defer to the value that's set in the config.
	signupAllowed := s.allowSignup == nil || *s.allowSignup

	var data extsvc.AccountData
	if err := azuredevops.SetExternalAccountData(&data, user, token); err != nil {
		return nil, "", errors.Wrapf(err, "failed to set external account data for azure devops user with email %q", user.EmailAddress)
	}

	// The API returned an email address with the first character capitalized during development.
	// Not taking any chances.
	email := strings.ToLower(user.EmailAddress)
	username, err := auth.NormalizeUsername(email)

	userID, safeErrMsg, err := auth.GetAndSaveUser(ctx, s.db, auth.GetAndSaveUserOp{
		UserProps: database.NewUser{
			Username: username,
			Email:    email,
			// TODO: Verify if we can assume this.
			EmailIsVerified: email != "",
			DisplayName:     user.DisplayName,
		},
		ExternalAccount: extsvc.AccountSpec{
			ServiceType: s.ServiceType,
			ServiceID:   s.ServiceID,
			ClientID:    s.clientID,
			AccountID:   user.ID,
		},
		ExternalAccountData: data,
		CreateIfNotExist:    signupAllowed,
	})
	if err != nil {
		return nil, safeErrMsg, err
	}

	return actor.FromUser(userID), "", nil
}

func (s *sessionIssuerHelper) DeleteStateCookie(w http.ResponseWriter) {}

func (s *sessionIssuerHelper) SessionData(token *oauth2.Token) oauth.SessionData {
	return oauth.SessionData{}
}

func (s *sessionIssuerHelper) AuthSucceededEventName() database.SecurityEventName {
	return database.SecurityEventAzureDevOpsAuthSucceeded
}

func (s *sessionIssuerHelper) AuthFailedEventName() database.SecurityEventName {
	return database.SecurityEventAzureDevOpsAuthFailed
}

func (s *sessionIssuerHelper) newOauth2Client() (*azuredevops.Client, error) {
	httpCli, err := httpcli.ExternalClientFactory.Doer()
	if err != nil {
		return nil, errors.Wrap(err, "azuredevops: failed to create Oauth2 client")
	}

	// FIXME: Empty token
	auth := extsvcauth.OAuthBearerToken{}
	return azuredevops.NewClient("azuredevopsoauth", s.CodeHost.BaseURL, &auth, httpCli)
}

type key int

const userKey key = iota

func withUser(ctx context.Context, user azuredevops.Profile) context.Context {
	return context.WithValue(ctx, userKey, user)
}

func userFromContext(ctx context.Context) (*azuredevops.Profile, error) {
	user, ok := ctx.Value(userKey).(azuredevops.Profile)
	if !ok {
		return nil, errors.Errorf("azuredevops: Context missing Azure DevOps user")
	}
	return &user, nil
}
