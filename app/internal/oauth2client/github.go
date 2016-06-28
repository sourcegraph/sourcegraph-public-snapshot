package oauth2client

import (
	"bytes"
	"errors"
	"net/http"
	"net/url"
	"os"
	"strings"

	"gopkg.in/inconshreveable/log15.v2"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	appauth "sourcegraph.com/sourcegraph/sourcegraph/app/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/canonicalurl"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/returnto"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/schemautil"
	"sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/errcode"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/httputil/httpctx"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/oauth2util"
	"sourcegraph.com/sourcegraph/sourcegraph/services/ext/github/githubcli"
)

const (
	githubAuthorizeURL = "https://github.com/login/oauth/authorize"
	githubTokenURL     = "https://github.com/login/oauth/access_token"
)

var (
	nonceCookiePath = router.Rel.URLTo(router.GitHubOAuth2Receive).Path

	githubClientID     string
	githubClientSecret string
)

func init() {
	internal.Handlers[router.GitHubOAuth2Initiate] = internal.Handler(serveGitHubOAuth2Initiate)
	internal.Handlers[router.GitHubOAuth2Receive] = internal.Handler(serveGitHubOAuth2Receive)

	githubClientID = os.Getenv("GITHUB_CLIENT_ID")
	githubClientSecret = os.Getenv("GITHUB_CLIENT_SECRET")
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

	nonce, err := writeNonceCookie(w, r, nonceCookiePath)
	if err != nil {
		return err
	}

	var scopes []string
	if s := r.URL.Query().Get("scopes"); s == "" {
		// scopes remains nil, and the GitHub OAuth2 flow authorizes
		// either no scopes or the previously authorized scopes to
		// this application.
	} else {
		scopes = strings.Split(s, ",")
	}

	destURL, err := oauthLoginURL(r, oauthAuthorizeClientState{Nonce: nonce, ReturnTo: returnTo.String()}, scopes)
	if err != nil {
		return err
	}

	http.Redirect(w, r, destURL.String(), http.StatusSeeOther)
	return nil
}

// oauthAuthorizeClientState holds the state that the OAuth2 client
// passes to the provider and expects to receive back, during the
// OAuth2 authorization flow.
//
// No authentication of these values is performed; callers should
// check that, e.g., the State field matches the cookie nonce.
type oauthAuthorizeClientState struct {
	// Nonce is the state nonce that the OAuth2 client expects to
	// receive from the provider. It is generated and stored in a
	// cookie by writeNonceCookie.
	Nonce string

	// ReturnTo is the request URI on the client app that the resource owner
	// should be redirected to, after successful completion of OAuth2
	// authorization.
	ReturnTo string
}

func (s oauthAuthorizeClientState) MarshalText() ([]byte, error) {
	return []byte(s.Nonce + ":" + s.ReturnTo), nil
}

func (s *oauthAuthorizeClientState) UnmarshalText(text []byte) error {
	parts := bytes.SplitN(text, []byte(":"), 2)
	if len(parts) != 2 {
		return errors.New("invalid OAuth2 authorize client state: no ':' delimiter")
	}
	*s = oauthAuthorizeClientState{Nonce: string(parts[0]), ReturnTo: string(parts[1])}
	return nil
}

func oauthLoginURL(r *http.Request, state oauthAuthorizeClientState, scopes []string) (*url.URL, error) {
	ctx := httpctx.FromRequest(r)

	stateText, err := state.MarshalText()
	if err != nil {
		return nil, err
	}

	return url.Parse(oauth2Config(ctx, scopes).AuthCodeURL(string(stateText)))
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

	ctx, cl := handlerutil.Client(r)

	actor := auth.ActorFromContext(ctx)

	var opt oauth2util.ReceiveParams
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
	deleteNonceCookie(w, nonceCookiePath) // prevent reuse of nonce
	if !present || nonce != state.Nonce || nonce == "" {
		return &errcode.HTTPErr{Status: http.StatusForbidden, Err: errors.New("invalid state nonce from OAuth2 provider")}
	}

	// Don't allow usage of the state's ReturnTo field until now that
	// we've checked the state against the nonce (which we do right
	// above).
	returnTo = state.ReturnTo

	tok, err := cl.Auth.GetAccessToken(ctx, &sourcegraph.AccessTokenRequest{
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
		if actor.IsAuthenticated() {
			// Logged in as a Sourcegraph user, has not yet linked GitHub.
			return linkAccountWithGitHub(w, r, ctx, cl, int32(actor.UID), ghUser, tok, true, state.ReturnTo)
		}

		// Not logged in as a Sourcegraph user, has not ever linked
		// this GitHub account to Sourcegraph.
		return createAccountFromGitHub(w, r, ctx, cl, ghUser, tok, state.ReturnTo)
	}

	// Logged in as a Sourcegraph user, has already linked GitHub.
	//
	// Elevate the actor to the Sourcegraph user identified by the
	// just-authenticated linked GitHub account. The user must have
	// previously linked the accounts for the Auth.GetAccessToken call to
	// return this Sourcegraph UID, so we can do this safely.
	ctx = sourcegraph.WithCredentials(ctx, oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: tok.AccessToken,
		TokenType:   "Bearer",
	}))
	httpctx.SetForRequest(r, ctx)
	return linkAccountWithGitHub(w, r, ctx, cl, tok.UID, ghUser, tok, false, state.ReturnTo)
}

func linkAccountWithGitHub(w http.ResponseWriter, r *http.Request, ctx context.Context, cl *sourcegraph.Client, sgUID int32, ghUser *sourcegraph.GitHubUser, tok *sourcegraph.AccessTokenResponse, firstTime bool, returnTo string) (err error) {
	defer func() {
		if err != nil {
			log15.Error("Error during GitHub account linking or login flow (suppressing HTTP 500 and returning redirect to non-GitHub login form).", "err", err, "sourcegraph-uid", sgUID, "github-login", ghUser.Login, "first-time", firstTime)
			http.Redirect(w, r, "/login?github-login-error=unknown&_event=FailedGitHubOAuth2Flow&return-to="+url.QueryEscape(returnTo), http.StatusSeeOther)
			err = nil
		}
	}()

	sgUser, err := cl.Users.Get(ctx, &sourcegraph.UserSpec{UID: sgUID})
	if err != nil {
		return err
	}

	_, err = cl.Auth.SetExternalToken(ctx, &sourcegraph.ExternalToken{
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
	if _, err := cl.Accounts.Update(ctx, sgUser); err != nil {
		return err
	}

	// Write cookie.
	cred := sourcegraph.CredentialsFromContext(ctx)
	sgTok, err := cred.Token()
	if err != nil {
		return err
	}
	if err := appauth.WriteSessionCookie(w, appauth.Session{AccessToken: sgTok.AccessToken}, appauth.OnlySecureCookies(ctx)); err != nil {
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
		q.Set("_event", "CompletedGitHubOAuth2Flow")
		q.Set("_githubAuthed", "true")
	}
	returnToURL.RawQuery = q.Encode()

	http.Redirect(w, r, returnToURL.String(), http.StatusSeeOther)
	return nil
}

func createAccountFromGitHub(w http.ResponseWriter, r *http.Request, ctx context.Context, cl *sourcegraph.Client, ghUser *sourcegraph.GitHubUser, tok *sourcegraph.AccessTokenResponse, returnTo string) (err error) {
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

	createdAcct, err := cl.Accounts.Create(ctx, &newAcct)
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

	ctx = sourcegraph.WithCredentials(ctx, oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: createdAcct.TemporaryAccessToken,
		TokenType:   "Bearer",
	}))
	httpctx.SetForRequest(r, ctx)

	return linkAccountWithGitHub(w, r, ctx, cl, createdAcct.UID, ghUser, tok, true, returnTo)
}

func oauth2Config(ctx context.Context, scopes []string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     githubClientID,
		ClientSecret: githubClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  githubAuthorizeURL,
			TokenURL: githubTokenURL,
		},
		RedirectURL: conf.AppURL(ctx).ResolveReference(router.Rel.URLTo(router.GitHubOAuth2Receive)).String(),
		Scopes:      scopes,
	}
}
