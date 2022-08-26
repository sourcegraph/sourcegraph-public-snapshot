package githuboauth

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dghubble/gologin/github"
	"github.com/inconshreveable/log15"
	"golang.org/x/oauth2"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth/providers"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/hubspot"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/hubspot/hubspotutil"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/auth/oauth"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	esauth "github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	githubsvc "github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type sessionIssuerHelper struct {
	*extsvc.CodeHost
	db           database.DB
	clientID     string
	allowSignup  bool
	allowOrgs    []string
	allowOrgsMap map[string][]string
}

func (s *sessionIssuerHelper) GetOrCreateUser(ctx context.Context, token *oauth2.Token, anonymousUserID, firstSourceURL, lastSourceURL string) (actr *actor.Actor, safeErrMsg string, err error) {
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

	ghClient := s.newClient(token.AccessToken)

	// 🚨 SECURITY: Ensure that the user email is verified
	verifiedEmails := getVerifiedEmails(ctx, ghClient)
	if len(verifiedEmails) == 0 {
		return nil, "Could not get verified email for GitHub user. Check that your GitHub account has a verified email that matches one of your Sourcegraph verified emails.", errors.New("no verified email")
	}

	// 🚨 SECURITY: Ensure that the user is part of one of the allow listed orgs or teams, if any.
	userBelongsToAllowedOrgsOrTeams := s.verifyUserOrgsAndTeams(ctx, ghClient)
	if !userBelongsToAllowedOrgsOrTeams {
		message := "user does not belong to allowed GitHub organizations or teams."
		return nil, message, errors.New(message)
	}

	// Try every verified email in succession until the first that succeeds
	var data extsvc.AccountData
	if err := githubsvc.SetExternalAccountData(&data, ghUser, token); err != nil {
		return nil, "", err
	}
	var (
		firstSafeErrMsg string
		firstErr        error
	)

	// We will first attempt to connect one of the verified emails with an existing
	// account in Sourcegraph
	type attemptConfig struct {
		email            string
		createIfNotExist bool
	}
	var attempts []attemptConfig
	for i := range verifiedEmails {
		attempts = append(attempts, attemptConfig{
			email:            verifiedEmails[i],
			createIfNotExist: false,
		})
	}
	// If allowSignup is true, we will create an account using the first verified
	// email address from GitHub which we expect to be their primary address. Note
	// that the order of attempts is important. If we manage to connect with an
	// existing account we return early and don't attempt to create a new account.
	if s.allowSignup {
		attempts = append(attempts, attemptConfig{
			email:            verifiedEmails[0],
			createIfNotExist: true,
		})
	}

	for i, attempt := range attempts {
		userID, safeErrMsg, err := auth.GetAndSaveUser(ctx, s.db, auth.GetAndSaveUserOp{
			UserProps: database.NewUser{
				Username:        login,
				Email:           attempt.email,
				EmailIsVerified: true,
				DisplayName:     deref(ghUser.Name),
				AvatarURL:       deref(ghUser.AvatarURL),
			},
			ExternalAccount: extsvc.AccountSpec{
				ServiceType: s.ServiceType,
				ServiceID:   s.ServiceID,
				ClientID:    s.clientID,
				AccountID:   strconv.FormatInt(derefInt64(ghUser.ID), 10),
			},
			ExternalAccountData: data,
			CreateIfNotExist:    attempt.createIfNotExist,
		})
		if err == nil {
			// Check if GitHub App access token
			if githubsvc.IsGitHubAppAccessToken(token.AccessToken) {
				installations, err := ghClient.GetUserInstallations(ctx)
				if err != nil {
					// Only log a warning, since we still want to create the user account
					// even if we fail to get installations.
					log15.Warn("Could not get GitHub App installations", "error", err)
				}
				for _, installation := range installations {
					accountID := strconv.FormatInt(*installation.ID, 10) + "/" + strconv.FormatInt(derefInt64(ghUser.ID), 10)
					_, _, err := auth.GetAndSaveUser(ctx, s.db, auth.GetAndSaveUserOp{
						UserProps: database.NewUser{
							Username:        login,
							Email:           attempt.email,
							EmailIsVerified: true,
							DisplayName:     deref(ghUser.Name),
							AvatarURL:       deref(ghUser.AvatarURL),
						},
						ExternalAccount: extsvc.AccountSpec{
							ServiceType: extsvc.TypeGitHubApp,
							ServiceID:   s.ServiceID,
							ClientID:    s.clientID,
							AccountID:   accountID,
						},
						CreateIfNotExist: attempt.createIfNotExist,
					})

					if err != nil {
						log15.Warn("Error while saving associated user installation", "error", err)
					}
				}
			}

			go hubspotutil.SyncUser(attempt.email, hubspotutil.SignupEventID, &hubspot.ContactProperties{
				AnonymousUserID: anonymousUserID,
				FirstSourceURL:  firstSourceURL,
				LastSourceURL:   lastSourceURL,
			})
			return actor.FromUser(userID), "", nil // success
		}
		if i == 0 {
			firstSafeErrMsg, firstErr = safeErrMsg, err
		}
	}

	// On failure, return the first error
	return nil, fmt.Sprintf("No user exists matching any of the verified emails: %s.\n\nFirst error was: %s", strings.Join(verifiedEmails, ", "), firstSafeErrMsg), firstErr
}

func (s *sessionIssuerHelper) CreateCodeHostConnection(ctx context.Context, token *oauth2.Token, providerID string) (svc *types.ExternalService, safeErrMsg string, err error) {
	actor := actor.FromContext(ctx)
	if !actor.IsAuthenticated() {
		return nil, "Must be authenticated to create code host connection from OAuth flow.", errors.New("unauthenticated request")
	}

	p := oauth.GetProvider(extsvc.TypeGitHub, providerID)
	if p == nil {
		return nil, "Could not find OAuth provider for the state.", errors.Errorf("provider not found for %q", providerID)
	}

	ghUser, err := github.UserFromContext(ctx)
	if ghUser == nil {
		if err != nil {
			err = errors.Wrap(err, "could not read user from context")
		} else {
			err = errors.New("could not read user from context")
		}
		return nil, "Could not read GitHub user from callback request.", err
	}

	// We have a special flow enabled when a user added code host has been created
	// without `repo` scope and we then enable private code on the instance. In this
	// case we allow the user to request the additional scope. This means that at
	// this point we may already have a code host and we just need to update the
	// token with the new one.

	tx, err := s.db.ExternalServices().Transact(ctx)
	if err != nil {
		return
	}
	defer func() {
		err = tx.Done(err)
		safeErrMsg = "Error committing transaction"
	}()

	services, err := tx.List(ctx, database.ExternalServicesListOptions{
		NamespaceUserID: actor.UID,
		Kinds:           []string{extsvc.KindGitHub},
	})
	if err != nil {
		return nil, "Error checking for existing external service", err
	}
	now := time.Now()
	if len(services) == 0 {
		// Nothing found, create new one
		svc = &types.ExternalService{
			Kind:        extsvc.KindGitHub,
			DisplayName: fmt.Sprintf("GitHub (%s)", deref(ghUser.Login)),
			Config: extsvc.NewUnencryptedConfig(fmt.Sprintf(`
{
  "url": "%s",
  "token": "%s",
  "orgs": []
}
`, p.ServiceID, token.AccessToken)),
			NamespaceUserID: actor.UID,
			CreatedAt:       now,
			UpdatedAt:       now,
		}
	} else if len(services) > 1 {
		return nil, "Multiple services of same kind found for user", errors.New("multiple services of same kind found for user")
	} else {
		// We have an existing service, update it
		svc = services[0]

		rawConfig, err := svc.Config.Decrypt(ctx)
		if err != nil {
			return nil, "", err
		}

		rawConfig, err = jsonc.Edit(rawConfig, token.AccessToken, "token")
		if err != nil {
			return nil, "Error updating OAuth token", err
		}
		svc.Config.Set(rawConfig)
		svc.UpdatedAt = now
	}

	err = tx.Upsert(ctx, svc)
	if err != nil {
		return nil, "Could not create code host connection.", err
	}
	return svc, "", nil // success
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

func (s *sessionIssuerHelper) newClient(token string) *githubsvc.V3Client {
	apiURL, _ := githubsvc.APIRoot(s.BaseURL)
	return githubsvc.NewV3Client(log.Scoped("session.github.v3", "github v3 client for session issuer"),
		extsvc.URNGitHubOAuth, apiURL, &esauth.OAuthBearerToken{Token: token}, nil)
}

// getVerifiedEmails returns the list of user emails that are verified. If the primary email is verified,
// it will be the first email in the returned list. It only checks the first 100 user emails.
func getVerifiedEmails(ctx context.Context, ghClient *githubsvc.V3Client) (verifiedEmails []string) {
	emails, err := ghClient.GetAuthenticatedUserEmails(ctx)
	if err != nil {
		log15.Warn("Could not get GitHub authenticated user emails", "error", err)
		return nil
	}

	for _, email := range emails {
		if !email.Verified {
			continue
		}
		if email.Primary {
			verifiedEmails = append([]string{email.Email}, verifiedEmails...)
			continue
		}
		verifiedEmails = append(verifiedEmails, email.Email)
	}
	return verifiedEmails
}

// verifyUserOrgs checks whether the authenticated user belongs to one of the GitHub orgs
// listed in auth.provider > allowOrgs configuration
func (s *sessionIssuerHelper) verifyUserOrgs(ctx context.Context, ghClient *githubsvc.V3Client) bool {
	allowed := make(map[string]bool, len(s.allowOrgs))
	for _, org := range s.allowOrgs {
		allowed[org] = true
	}

	hasNextPage := true
	var userOrgs []*githubsvc.Org
	var err error
	page := 1
	for hasNextPage {
		userOrgs, hasNextPage, _, err = ghClient.GetAuthenticatedUserOrgsForPage(ctx, page)

		if err != nil {
			log15.Warn("Could not get GitHub authenticated user organizations", "error", err)
			return false
		}

		for _, org := range userOrgs {
			if allowed[org.Login] {
				return true
			}
		}
		page++
	}

	return false
}

// verifyUserTeams checks whether the authenticated user belongs to one of the GitHub teams listed in the auth.provider > allowOrgsMap configuration
func (s *sessionIssuerHelper) verifyUserTeams(ctx context.Context, ghClient *githubsvc.V3Client) bool {
	var err error
	hasNextPage := true
	allowedTeams := make(map[string]map[string]bool, len(s.allowOrgsMap))

	for org, teams := range s.allowOrgsMap {
		teamsMap := make(map[string]bool)
		for _, team := range teams {
			teamsMap[team] = true
		}

		allowedTeams[org] = teamsMap
	}

	for page := 1; hasNextPage; page++ {
		var githubTeams []*githubsvc.Team

		githubTeams, hasNextPage, _, err = ghClient.GetAuthenticatedUserTeams(ctx, page)
		if err != nil {
			log15.Warn("Could not get GitHub authenticated user teams", "error", err)
			return false
		}

		for _, ghTeam := range githubTeams {
			_, ok := allowedTeams[ghTeam.Organization.Login][ghTeam.Name]
			if ok {
				return true
			}
		}
	}

	return false
}

// verifyUserOrgsAndTeams checks if the user belongs to one of the allowed listed orgs or teams provided in the auth.provider configuration.
func (s *sessionIssuerHelper) verifyUserOrgsAndTeams(ctx context.Context, ghClient *githubsvc.V3Client) bool {
	if len(s.allowOrgs) == 0 && len(s.allowOrgsMap) == 0 {
		return true
	}

	if len(s.allowOrgs) > 0 && s.verifyUserOrgs(ctx, ghClient) {
		return true
	}

	if len(s.allowOrgsMap) > 0 && s.verifyUserTeams(ctx, ghClient) {
		return true
	}

	return false
}
