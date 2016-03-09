package oauth2client

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/sourcegraph/go-github/github"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"gopkg.in/inconshreveable/log15.v2"

	"src.sourcegraph.com/sourcegraph/app/internal"
	"src.sourcegraph.com/sourcegraph/app/internal/returnto"
	"src.sourcegraph.com/sourcegraph/app/internal/schemautil"
	"src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/conf"
	"src.sourcegraph.com/sourcegraph/errcode"
	"src.sourcegraph.com/sourcegraph/ext/github/githubcli"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/notif"
	"src.sourcegraph.com/sourcegraph/pkg/oauth2util"
	"src.sourcegraph.com/sourcegraph/util/eventsutil"
	"src.sourcegraph.com/sourcegraph/util/githubutil"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
)

const (
	githubAuthorizeUrl = "https://github.com/login/oauth/authorize"
	githubTokenUrl     = "https://github.com/login/oauth/access_token"
)

var (
	scopes          = []string{"repo", "read:org", "user"}
	nonceCookiePath = router.Rel.URLTo(router.GitHubOAuth2Receive).Path

	githubClientID     string
	githubClientSecret string
)

func init() {
	internal.Handlers[router.GitHubOAuth2Initiate] = serveGitHubOAuth2Initiate
	internal.Handlers[router.GitHubOAuth2Receive] = serveGitHubOAuth2Receive

	githubClientID = os.Getenv("GITHUB_CLIENT_ID")
	githubClientSecret = os.Getenv("GITHUB_CLIENT_SECRET")
}

// serveGitHubOAuth2Initiate generates the OAuth2 authorize URL
// (including a nonce state value, also stored in a cookie) and
// redirects the client to that URL.
func serveGitHubOAuth2Initiate(w http.ResponseWriter, r *http.Request) error {
	returnToURL, err := url.Parse(r.Referer())
	if err != nil {
		return err
	}
	returnTo := returnToURL.Path
	if err := returnto.CheckSafe(returnTo); err != nil {
		return err
	}

	nonce, err := writeNonceCookie(w, r, nonceCookiePath)
	if err != nil {
		return err
	}

	destURL, err := oauthLoginURL(r, oauthAuthorizeClientState{Nonce: nonce, ReturnTo: returnTo})
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

func oauthLoginURL(r *http.Request, state oauthAuthorizeClientState) (*url.URL, error) {
	ctx := httpctx.FromRequest(r)

	oauthCfg := getOAuth2Conf(ctx)

	stateText, err := state.MarshalText()
	if err != nil {
		return nil, err
	}

	return url.Parse(oauthCfg.AuthCodeURL(string(stateText)))
}

func serveGitHubOAuth2Receive(w http.ResponseWriter, r *http.Request) (err error) {
	ctx, cl := handlerutil.Client(r)
	currentUser := handlerutil.UserFromRequest(r)
	if currentUser == nil {
		return &errcode.HTTPErr{Status: http.StatusForbidden, Err: errors.New("user must be logged in")}
	}

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

	oauthCfg := getOAuth2Conf(ctx)
	token, err := oauthCfg.Exchange(oauth2.NoContext, opt.Code)
	if err != nil {
		return &errcode.HTTPErr{Status: http.StatusBadRequest, Err: err}
	}

	if !token.Valid() {
		return &errcode.HTTPErr{Status: http.StatusBadRequest, Err: errors.New("invalid token from OAuth2 provider")}
	}

	client := githubutil.Default.AuthedClient(token.AccessToken)
	user, _, err := client.Users.Get("")
	if err != nil {
		return &errcode.HTTPErr{Status: http.StatusBadRequest, Err: err}
	}

	_, err = cl.Auth.SetExternalToken(ctx, &sourcegraph.ExternalToken{
		UID:      currentUser.UID,
		Host:     githubcli.Config.Host(),
		Token:    token.AccessToken,
		Scope:    strings.Join(scopes, ","),
		ClientID: githubClientID,
		ExtUID:   int32(*user.ID),
	})
	if err != nil {
		return &errcode.HTTPErr{Status: http.StatusBadRequest, Err: err}
	}

	eventsutil.LogLinkGitHub(ctx, user)
	sendLinkGitHubSlackMsg(ctx, currentUser, user)

	sgUser, err := cl.Users.Get(ctx, &sourcegraph.UserSpec{UID: currentUser.UID})
	if err != nil {
		return &errcode.HTTPErr{Status: http.StatusBadRequest, Err: err}
	}
	if sgUser.Name == "" && user.Name != nil {
		sgUser.Name = *user.Name
	}
	if sgUser.AvatarURL == "" && user.AvatarURL != nil {
		sgUser.AvatarURL = *user.AvatarURL
	}
	if sgUser.Location == "" && user.Location != nil {
		sgUser.Location = *user.Location
	}

	_, err = cl.Accounts.Update(ctx, sgUser)
	if err != nil {
		log15.Info("Could not update profile info", "github_user", *user.Login, "sourcegraph_user", currentUser.Login)
	}

	returnTo := state.ReturnTo
	if err := returnto.CheckSafe(returnTo); err != nil {
		return err
	}
	u, err := url.Parse(returnTo)
	if err != nil {
		return err
	}
	q := u.Query()
	q.Set("github-onboarding", "true")
	u.RawQuery = q.Encode()
	http.Redirect(w, r, u.String(), http.StatusSeeOther)
	return nil
}

func getOAuth2Conf(ctx context.Context) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     githubClientID,
		ClientSecret: githubClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  githubAuthorizeUrl,
			TokenURL: githubTokenUrl,
		},
		RedirectURL: conf.AppURL(ctx).ResolveReference(router.Rel.URLTo(router.GitHubOAuth2Receive)).String(),
		Scopes:      scopes,
	}
}

func sendLinkGitHubSlackMsg(ctx context.Context, sgUser *sourcegraph.UserSpec, ghUser *github.User) {
	var ghLogin, ghName, ghEmail string
	if ghUser.Login != nil {
		ghLogin = *ghUser.Login
	}
	if ghUser.Name != nil {
		ghName = *ghUser.Name
	}
	if ghUser.Email != nil {
		ghEmail = *ghUser.Email
	}
	msg := fmt.Sprintf("User *%s* linked their GitHub account: *%s* (%s <%s>)", sgUser.Login, ghLogin, ghName, ghEmail)
	notif.ActionSlackMessage(notif.ActionContext{SlackMsg: msg})
}
