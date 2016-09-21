package oauth2client

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"gopkg.in/inconshreveable/log15.v2"

	"golang.org/x/oauth2"

	"sourcegraph.com/sourcegraph/sourcegraph/app/internal"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/canonicalurl"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/returnto"
	"sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/errcode"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/githubutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/randstring"
)

var githubNonceCookiePath = router.Rel.URLTo(router.GitHubOAuth2Receive).Path

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

	nonce := randstring.NewLen(32)
	http.SetCookie(w, &http.Cookie{
		Name:    "nonce",
		Value:   nonce,
		Path:    "",
		Expires: time.Now().Add(10 * time.Minute),
	})

	var scopes []string
	if s := r.URL.Query().Get("scopes"); s == "" {
		// if we have no scope, we upgrade the credential to the
		// minimum scope required, read access to email
		scopes = []string{"user:email"}
	} else {
		scopes = strings.Split(s, ",")
	}

	http.Redirect(w, r, githubutil.Default.OAuth.AuthCodeURL(nonce+":"+returnTo.String(),
		oauth2.SetAuthURLParam("scope", strings.Join(scopes, " ")),
		oauth2.SetAuthURLParam("redirect_uri", conf.AppURL.ResolveReference(router.Rel.URLTo(router.GitHubOAuth2Receive)).String()),
	), http.StatusSeeOther)
	return nil
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

	parts := strings.SplitN(r.URL.Query().Get("state"), ":", 2)
	if len(parts) != 2 {
		return &errcode.HTTPErr{Status: http.StatusBadRequest, Err: errors.New("invalid OAuth2 authorize client state")}
	}

	nonceCookie, err := r.Cookie("nonce")
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:   "nonce",
		Path:   "/",
		MaxAge: -1,
	})
	if len(parts) != 2 || nonceCookie.Value == "" || parts[0] != nonceCookie.Value {
		return &errcode.HTTPErr{Status: http.StatusForbidden, Err: errors.New("invalid state")}
	}
	returnTo = parts[1]

	// Exchange the code for a GitHub access token.
	ghToken, err := githubutil.Default.OAuth.Exchange(r.Context(), r.URL.Query().Get("code"))
	if err != nil {
		return err
	}
	if !ghToken.Valid() {
		return grpc.Errorf(codes.PermissionDenied, "exchanging auth code yielded invalid GitHub OAuth2 token")
	}

	// Get the current user.
	ghUser, ghResp, err := githubutil.Default.AuthedClient(ghToken.AccessToken).Users.Get("")
	if err != nil {
		return err
	}

	firstTime := false

	defer func() {
		if err != nil {
			log15.Error("Error during GitHub account linking or login flow (suppressing HTTP 500 and returning redirect to non-GitHub login form).", "err", err, "sourcegraph-uid", *ghUser.ID, "first-time", firstTime)
			http.Redirect(w, r, "/login?github-login-error=unknown&_event=FailedGitHubOAuth2Flow&return-to="+url.QueryEscape(returnTo), http.StatusSeeOther)
			err = nil
		}
	}()

	// Write cookie.
	if err := auth.StartNewSession(w, r, &auth.Actor{
		UID:             strconv.Itoa(*ghUser.ID),
		Login:           *ghUser.Login,
		Email:           *ghUser.Email,
		AvatarURL:       *ghUser.AvatarURL,
		GitHubConnected: true,
		GitHubScopes:    ghResp.Header["X-Oauth-Scopes"],
		GitHubToken:     ghToken.AccessToken,
	}); err != nil {
		return err
	}

	// Add tracking info to return-to URL.
	returnToURL, err := url.Parse(returnTo)
	if err != nil {
		return err
	}
	q := returnToURL.Query()
	if firstTime {
		q.Set("ob", "chrome")
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
