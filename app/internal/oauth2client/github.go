package oauth2client

import (
	"errors"
	"net/http"
	"net/url"
	"os"
	"strings"

	"gopkg.in/inconshreveable/log15.v2"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"context"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	appauth "sourcegraph.com/sourcegraph/sourcegraph/app/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/canonicalurl"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/returnto"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/schemautil"
	"sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/errcode"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/services/ext/github/githubcli"
	"sourcegraph.com/sqs/pbtypes"
)

var (
	githubNonceCookiePath = router.Rel.URLTo(router.GitHubOAuth2Receive).Path

	githubClientID     = os.Getenv("GITHUB_CLIENT_ID")
	githubClientSecret = os.Getenv("GITHUB_CLIENT_SECRET")
)

func init() {
	internal.Handlers[router.GitHubOAuth2Initiate] = internal.Handler(serveGitHubOAuth2Initiate)
	internal.Handlers[router.GitHubOAuth2Receive] = internal.Handler(serveGitHubOAuth2Receive)
}

// serveGitHubOAuth2Initiate generates the OAuth2 authorize URL
// (including a nonce state value, also stored in a cookie) and
// redirects the client to that URL.
func serveGitHubOAuth2Initiate(w http.ResponseWriter, r *http.Request) error {
	returnTo, err := returnto.URLFromRequest(r)
	if err != nil {
		log15.Warn("Invalid return-to URL provided to OAuth2 flow initiation; ignoring.", "err", err)
	}

	// Remove UTM campaign params to avoid double
	// attribution. TODO(sqs): consider doing this on the frontend in
	// JS so we centralize usage analytics there.
	returnTo = canonicalurl.FromURL(returnTo)

	nonce, err := writeNonceCookie(w, r, githubNonceCookiePath)
	if err != nil {
		return err
	}

	var scopes []string
	if s := r.URL.Query().Get("scopes"); s == "" {
		// if we have no scope, we upgrade the credential to the
		// minimum scope required, read access to email
		scopes = []string{"user:email"}
	} else {
		scopes = strings.Split(s, ",")
	}

	destURL, err := githubOAuthLoginURL(r, oauthAuthorizeClientState{Nonce: nonce, ReturnTo: returnTo.String()}, scopes)
	if err != nil {
		return err
	}

	http.Redirect(w, r, destURL.String(), http.StatusSeeOther)
	return nil
}

func githubOAuthLoginURL(r *http.Request, state oauthAuthorizeClientState, scopes []string) (*url.URL, error) {
	stateText, err := state.MarshalText()
	if err != nil {
		return nil, err
	}

	return url.Parse(githubOAuth2Config(r.Context(), scopes).AuthCodeURL(string(stateText)))
}

func githubOAuth2Config(ctx context.Context, scopes []string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     githubClientID,
		ClientSecret: githubClientSecret,
		Endpoint:     github.Endpoint,
		RedirectURL:  conf.AppURL.ResolveReference(router.Rel.URLTo(router.GitHubOAuth2Receive)).String(),
		Scopes:       scopes,
	}
}

func serveGitHubOAuth2Receive(w http.ResponseWriter, r *http.Request) (err error) {
	returnTo := "/"

	defer func() {
		if err != nil {
			log15.Error("Error in receive handler in GitHub OAuth2 auth flow (suppressing HTTP 500 and returning redirect to non-GitHub login form).", "err", err)
			http.Redirect(w, r, "/login?github-login-error=unknown&_event=FailedGitHubOAuth2Flow&return-to="+url.QueryEscape(returnTo), http.StatusSeeOther)
			err = nil
		}
	}()

	cl := handlerutil.Client(r)
	authInfo, err := cl.Auth.Identify(r.Context(), &pbtypes.Void{})
	if err != nil {
		return err
	}

	var opt oauthReceiveParams
	if err := schemautil.Decode(&opt, r.URL.Query()); err != nil {
		return err
	}

	// Check the state nonce against what's stored in the cookie (to
	// prevent CSRF).
	var state oauthAuthorizeClientState
	if err := state.UnmarshalText([]byte(opt.State)); err != nil {
		return &errcode.HTTPErr{Status: http.StatusBadRequest, Err: err}
	}
	nonce, present := nonceFromCookie(r)
	deleteNonceCookie(w, githubNonceCookiePath) // prevent reuse of nonce
	if !present || nonce != state.Nonce || nonce == "" {
		return &errcode.HTTPErr{Status: http.StatusForbidden, Err: errors.New("invalid state nonce from OAuth2 provider")}
	}

	// Don't allow usage of the state's ReturnTo field until now that
	// we've checked the state against the nonce (which we do right
	// above).
	returnTo = state.ReturnTo

	tok, err := cl.Auth.GetAccessToken(r.Context(), &sourcegraph.AccessTokenRequest{
		AuthorizationGrant: &sourcegraph.AccessTokenRequest_GitHubAuthCode{
			GitHubAuthCode: &sourcegraph.GitHubAuthCode{
				Code: opt.Code,
				Host: "github.com",
			},
		},
	})
	if err != nil {
		return err
	}

	ghUser := tok.GitHubUser

	// If this GitHub user is already authed with us, then continue
	// logging in. Otherwise continue to create an account.
	if tok.UID == 0 {
		if authInfo.UID != 0 {
			// Logged in as a Sourcegraph user, has not yet linked GitHub.
			return linkAccountWithGitHub(w, r, cl, authInfo.UID, ghUser, tok, true, state.ReturnTo)
		}

		// Not logged in as a Sourcegraph user, has not ever linked
		// this GitHub account to Sourcegraph.
		return createAccountFromGitHub(w, r, cl, ghUser, tok, state.ReturnTo)
	}

	// Logged in as a Sourcegraph user, has already linked GitHub.
	//
	// Elevate the credentials to the Sourcegraph user identified by the
	// just-authenticated linked GitHub account. The user must have previously
	// linked the accounts for the Auth.GetAccessToken call to return this
	// Sourcegraph UID, so we can do this safely.
	r = r.WithContext(sourcegraph.WithAccessToken(r.Context(), tok.AccessToken))

	return linkAccountWithGitHub(w, r, cl, tok.UID, ghUser, tok, false, state.ReturnTo)
}

func linkAccountWithGitHub(w http.ResponseWriter, r *http.Request, cl *sourcegraph.Client, sgUID int32, ghUser *sourcegraph.GitHubUser, tok *sourcegraph.AccessTokenResponse, firstTime bool, returnTo string) (err error) {
	defer func() {
		if err != nil {
			log15.Error("Error during GitHub account linking or login flow (suppressing HTTP 500 and returning redirect to non-GitHub login form).", "err", err, "sourcegraph-uid", sgUID, "github-login", ghUser.Login, "first-time", firstTime)
			http.Redirect(w, r, "/login?github-login-error=unknown&_event=FailedGitHubOAuth2Flow&return-to="+url.QueryEscape(returnTo), http.StatusSeeOther)
			err = nil
		}
	}()

	sgUser, err := cl.Users.Get(r.Context(), &sourcegraph.UserSpec{UID: sgUID})
	if err != nil {
		return err
	}

	_, err = cl.Auth.SetExternalToken(r.Context(), &sourcegraph.ExternalToken{
		UID:      sgUID,
		Host:     githubcli.Config.Host(),
		Token:    tok.GitHubAccessToken,
		Scope:    strings.Join(tok.Scope, ","),
		ClientID: githubClientID,
		ExtUID:   ghUser.ID,
	})
	if err != nil {
		return &errcode.HTTPErr{Status: http.StatusBadRequest, Err: err}
	}

	sgUser.Name = ghUser.Name
	sgUser.Location = ghUser.Location
	sgUser.Company = ghUser.Company
	sgUser.AvatarURL = ghUser.AvatarURL
	if _, err := cl.Accounts.Update(r.Context(), sgUser); err != nil {
		return err
	}

	// Write cookie.
	token := sourcegraph.AccessTokenFromContext(r.Context())
	if err := appauth.WriteSessionCookie(w, appauth.Session{AccessToken: token}, conf.AppURL.Scheme == "https"); err != nil {
		return err
	}

	// Add tracking info to return-to URL.
	returnToURL, err := url.Parse(returnTo)
	if err != nil {
		return err
	}
	q := returnToURL.Query()
	if firstTime {
		q.Set("_event", "SignupCompleted")
		q.Set("_signupChannel", "GitHubOAuth")
		q.Set("_githubAuthed", "true")
	} else {
		// Do not redirect a user while inside the onboarding flow.
		// This is accomplished by not removing the onboarding query params.
		if q.Get("ob") != "github" {
			q.Del("ob")
		}
		q.Set("_event", "CompletedGitHubOAuth2Flow")
		q.Set("_githubAuthed", "true")
	}
	returnToURL.RawQuery = q.Encode()

	http.Redirect(w, r, returnToURL.String(), http.StatusSeeOther)
	return nil
}

func createAccountFromGitHub(w http.ResponseWriter, r *http.Request, cl *sourcegraph.Client, ghUser *sourcegraph.GitHubUser, tok *sourcegraph.AccessTokenResponse, returnTo string) (err error) {
	defer func() {
		if err != nil {
			log15.Error("Error during GitHub account creation (suppressing HTTP 500 and returning redirect to non-GitHub signup form).", "err", err, "github-login", ghUser.Login)
			http.Redirect(w, r, "/join?github-signup-error=unknown&_event=FailedGitHubOAuth2Flow&return-to="+url.QueryEscape(returnTo), http.StatusSeeOther)
			err = nil
		}
	}()

	var newAcct sourcegraph.NewAccount
	newAcct.Login = ghUser.Login
	if !strings.HasSuffix(ghUser.Email, "@users.noreply.github.com") {
		newAcct.Email = ghUser.Email
	}

	createdAcct, err := cl.Accounts.Create(r.Context(), &newAcct)
	if grpc.Code(err) == codes.AlreadyExists {
		// There is already a Sourcegraph user whose username is this
		// user's GitHub username. Redirect to the app and tell the
		// user they need to create a unique Sourcegraph account
		// first, and then they can *link* their GitHub account to
		// their newly created Sourcegraph account.
		http.Redirect(w, r, "/join?github-signup-error=username-or-email-taken&login="+url.QueryEscape(newAcct.Login)+"&email="+url.QueryEscape(newAcct.Email)+"&_event=FailedGitHubOAuth2Flow&return-to="+url.QueryEscape(returnTo), http.StatusSeeOther)
		return nil
	} else if err != nil {
		return err
	}
	log15.Info("Created Sourcegraph account from GitHub account", "uid", createdAcct.UID, "login", newAcct.Login, "email", newAcct.Email)

	r = r.WithContext(sourcegraph.WithAccessToken(r.Context(), createdAcct.TemporaryAccessToken))

	return linkAccountWithGitHub(w, r, cl, createdAcct.UID, ghUser, tok, true, returnTo)
}
