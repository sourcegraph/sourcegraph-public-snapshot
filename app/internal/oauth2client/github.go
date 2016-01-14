package oauth2client

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/google/go-github/github"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"

	"src.sourcegraph.com/sourcegraph/app/internal"
	"src.sourcegraph.com/sourcegraph/app/internal/returnto"
	"src.sourcegraph.com/sourcegraph/app/internal/schemautil"
	"src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/conf"
	"src.sourcegraph.com/sourcegraph/errcode"
	"src.sourcegraph.com/sourcegraph/pkg/oauth2util"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
)

const (
	githubAuthorizeUrl = "https://github.com/login/oauth/authorize"
	githubTokenUrl     = "https://github.com/login/oauth/access_token"
)

var (
	scopes = []string{"repo", "read:org"}

	nonceCookiePath = router.Rel.URLTo(router.GitHubOAuth2Receive).Path
)

func init() {
	internal.Handlers[router.GitHubOAuth2Initiate] = serveGitHubOAuth2Initiate
	internal.Handlers[router.GitHubOAuth2Receive] = serveGitHubOAuth2Receive
}

// serveGitHubOAuth2Initiate generates the OAuth2 authorize URL
// (including a nonce state value, also stored in a cookie) and
// redirects the client to that URL.
func serveGitHubOAuth2Initiate(w http.ResponseWriter, r *http.Request) error {
	returnTo := router.Rel.URLTo(router.Home).String()

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
	ctx := httpctx.FromRequest(r)
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

	client := github.NewClient(oauthCfg.Client(oauth2.NoContext, token))

	user, _, err := client.Users.Get("")
	if err != nil {
		return &errcode.HTTPErr{Status: http.StatusBadRequest, Err: err}
	}

	log.Printf("github user %s linked to sourcegraph user %s", *user.Login, currentUser.Login)

	returnTo := state.ReturnTo
	if err := returnto.CheckSafe(returnTo); err != nil {
		return err
	}
	http.Redirect(w, r, returnTo, http.StatusSeeOther)
	return nil
}

func getOAuth2Conf(ctx context.Context) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		Endpoint: oauth2.Endpoint{
			AuthURL:  githubAuthorizeUrl,
			TokenURL: githubTokenUrl,
		},
		RedirectURL: conf.AppURL(ctx).ResolveReference(router.Rel.URLTo(router.GitHubOAuth2Receive)).String(),
		Scopes:      scopes,
	}
}
