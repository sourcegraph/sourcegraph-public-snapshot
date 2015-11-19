package oauth2client

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"gopkg.in/inconshreveable/log15.v2"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sqs/pbtypes"
	appauth "src.sourcegraph.com/sourcegraph/app/auth"
	"src.sourcegraph.com/sourcegraph/app/internal"
	"src.sourcegraph.com/sourcegraph/app/internal/authutil"
	"src.sourcegraph.com/sourcegraph/app/internal/returnto"
	"src.sourcegraph.com/sourcegraph/app/internal/schemautil"
	"src.sourcegraph.com/sourcegraph/app/internal/tmpl"
	"src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/auth/idkey"
	"src.sourcegraph.com/sourcegraph/client/pkg/oauth2client"
	"src.sourcegraph.com/sourcegraph/conf"
	"src.sourcegraph.com/sourcegraph/errcode"
	"src.sourcegraph.com/sourcegraph/fed"
	"src.sourcegraph.com/sourcegraph/pkg/oauth2util"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
)

func init() {
	internal.Handlers[router.OAuth2ClientInitiate] = redirectToOAuth2Authorize
	internal.Handlers[router.OAuth2ClientReceive] = serveOAuth2ClientReceive
}

var renderedErrorPage = errors.New("checkOAuth2Config already rendered an error page") // sentinel error value

// checkOAuth2Config shecks for some common problems that will cause
// OAuth2 authorization to fail.
func checkOAuth2Config(w http.ResponseWriter, r *http.Request) error {
	cl := handlerutil.APIClient(r)
	ctx := httpctx.FromRequest(r)

	conf, err := cl.Meta.Config(ctx, &pbtypes.Void{})
	if err != nil {
		return err
	}
	rootURL, err := url.Parse(conf.FederationRootURL)
	if err != nil {
		return err
	}

	// Check that the server's AppURL scheme/host/port matches
	// those of URL in the person's browser.
	expected, err := url.Parse(conf.AppURL)
	if err != nil {
		return err
	}
	expected.Path = ""
	actual := &url.URL{Host: r.Host}
	if isHTTPS := r.TLS != nil || r.Header.Get("x-forwarded-proto") == "https"; isHTTPS {
		actual.Scheme = "https"
	} else {
		actual.Scheme = "http"
	}
	if *expected != *actual {
		log15.Warn("App URL & request URL scheme/host mismatch", "expected", expected.String(), "actual", actual.String(), "referrer", r.Referer())
		err := tmpl.Exec(r, w, "oauth-client/app_url_mismatch.error.html", http.StatusConflict, nil, &struct {
			RootHostname     string
			Expected, Actual string
			tmpl.Common
		}{
			RootHostname: rootURL.Host,
			Expected:     expected.String(),
			Actual:       actual.String(),
		})
		if err != nil {
			return err
		}
		return renderedErrorPage
	}

	// Check that server's AppURL conforms to a registered OAuth2
	// redirect URI.
	regClient, err := getRegisteredClientForServer(ctx)
	if err != nil && grpc.Code(err) != codes.NotFound {
		return err
	} else if err == nil {
		oauth2Conf, err := oauth2client.Config(ctx)
		if err != nil {
			return err
		}
		if err := oauth2util.AllowRedirectURI(regClient.RedirectURIs, oauth2Conf.RedirectURL); err != nil {
			log15.Warn("App URL & OAuth2 redirect URI mismatch", "registered", regClient.RedirectURIs, "actual", oauth2Conf.RedirectURL)
			err := tmpl.Exec(r, w, "oauth-client/redirect_uri_mismatch.error.html", http.StatusInternalServerError, nil, &struct {
				RootHostname string
				RedirectURL  string
				tmpl.Common
			}{
				RootHostname: rootURL.Host,
				RedirectURL:  oauth2Conf.RedirectURL,
			})
			if err != nil {
				return err
			}
			return renderedErrorPage
		}
	}

	return nil
}

// getRegisteredClientForServer gets the mothership's registered
// client entry for this server.
var getRegisteredClientForServer = func(ctx context.Context) (*sourcegraph.RegisteredClient, error) {
	rctx := fed.Config.NewRemoteContext(ctx)
	rcl := sourcegraph.NewClientFromContext(rctx)
	c, err := rcl.RegisteredClients.GetCurrent(rctx, &pbtypes.Void{})
	return c, err
}

// ServeOAuth2Initiate serves the welcome screen for OAuth2, which
// will open a popup to complete OAuth2 on the root server.
func ServeOAuth2Initiate(w http.ResponseWriter, r *http.Request) error {
	cl := handlerutil.APIClient(r)
	ctx := httpctx.FromRequest(r)

	u := handlerutil.UserFromContext(ctx)
	if u != nil && u.UID != 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return nil
	}

	if err := checkOAuth2Config(w, r); err == renderedErrorPage {
		return nil
	} else if err != nil {
		return err
	}

	conf, err := cl.Meta.Config(ctx, &pbtypes.Void{})
	if err != nil {
		return err
	}
	rootURL, err := url.Parse(conf.FederationRootURL)
	if err != nil {
		return err
	}

	return tmpl.Exec(r, w, "oauth-client/initiate.html", http.StatusOK, nil, &struct {
		RootHostname string
		tmpl.Common
	}{
		RootHostname: rootURL.Host,
	})
}

// redirectToOAuth2Authorize generates the OAuth2 authorize URL
// (including a nonce state value, also stored in a cookie) and
// redirects the client to that URL.
func redirectToOAuth2Authorize(w http.ResponseWriter, r *http.Request) error {
	if err := checkOAuth2Config(w, r); err == renderedErrorPage {
		return nil
	} else if err != nil {
		return err
	}

	returnTo, err := returnto.BestGuess(r)
	if err != nil {
		return err
	}

	nonce, err := writeNonceCookie(w, r)
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

func init() {
	authutil.RedirectToOAuth2Initiate = ServeOAuth2Initiate
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

	oauth2Conf, err := oauth2client.Config(ctx)
	if err != nil {
		return nil, err
	}

	stateText, err := state.MarshalText()
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(oauth2Conf.AuthCodeURL(string(stateText)))
	if err != nil {
		return nil, err
	}
	return withJWKS(r, u)
}

func serveOAuth2ClientReceive(w http.ResponseWriter, r *http.Request) (err error) {
	// Authentication-required errors in this handler should NOT be
	// handled by redirecting to the login page of the AS, since the
	// user most likely just came from the AS's login page (and it
	// would be confusing to send them back there, since the problem
	// is due to some other issue, not their not having logged
	// in). So, catch these errors and show a special error page.
	defer func() {
		if grpc.Code(err) == codes.Unauthenticated {
			log.Printf("Error in OAuth2 client receive handler: %s.", err)
			internal.HandleError(w, r, http.StatusInternalServerError,
				fmt.Errorf("unexpected internal error during OAuth2 authentication: %s", err),
			)
			err = nil
		}
	}()

	cl := handlerutil.APIClient(r)
	ctx := httpctx.FromRequest(r)

	currentUser := handlerutil.UserFromRequest(r)
	if currentUser != nil {
		return &errcode.HTTPErr{Status: http.StatusForbidden, Err: errors.New("already logged in")}
	}

	var opt oauth2util.ReceiveParams
	if err := schemautil.Decode(&opt, r.URL.Query()); err != nil {
		return err
	}

	// Check client ID.
	clientID := oauth2client.ClientIDFromContext(ctx)
	if clientID == "" {
		return oauth2client.ErrClientNotRegistered
	}
	if opt.ClientID == "" {
		return &errcode.HTTPErr{
			Status: http.StatusBadRequest,
			Err:    errors.New("no OAuth2 client ID"),
		}
	}
	if clientID != opt.ClientID {
		return &errcode.HTTPErr{
			Status: http.StatusForbidden,
			Err:    errors.New("OAuth2 client ID mismatch"),
		}
	}

	// Check the state nonce against what's stored in the cookie (to
	// prevent CSRF).
	var state oauthAuthorizeClientState
	if err := state.UnmarshalText([]byte(opt.State)); err != nil {
		return &errcode.HTTPErr{Status: http.StatusBadRequest, Err: err}
	}
	nonce, present := nonceFromCookie(r)
	deleteNonceCookie(w) // prevent reuse of nonce
	if !present || nonce != state.Nonce || nonce == "" {
		return &errcode.HTTPErr{Status: http.StatusForbidden, Err: errors.New("invalid state nonce from OAuth2 provider")}
	}

	// Configure OAuth2.
	origRedirectURI := conf.AppURL(ctx).ResolveReference(r.URL)
	origRedirectURI.RawQuery = ""
	conf, err := oauth2client.Config(ctx)
	if err != nil {
		return err
	}
	conf.RedirectURL = origRedirectURI.String()

	atok, err := cl.Auth.GetAccessToken(ctx, &sourcegraph.AccessTokenRequest{
		TokenURL: conf.Endpoint.TokenURL,
		AuthorizationGrant: &sourcegraph.AccessTokenRequest_AuthorizationCode{
			AuthorizationCode: &sourcegraph.AuthorizationCode{
				Code:        opt.Code,
				RedirectURI: origRedirectURI.String(),
			},
		},
	})
	if err != nil {
		return err
	}
	tok := &oauth2.Token{
		AccessToken: atok.AccessToken,
		TokenType:   atok.TokenType,
		Expiry:      time.Now().Add(time.Duration(atok.ExpiresInSec) * time.Second),
	}

	// Authenticate future requests.
	ctx = sourcegraph.WithCredentials(ctx, oauth2.StaticTokenSource(tok))

	// Get user.
	authInfo, err := cl.Auth.Identify(ctx, &pbtypes.Void{})
	if err != nil {
		return err
	}
	if authInfo.UID == 0 {
		return fmt.Errorf("UID is zero in auth info: %v (token: %v)", authInfo, tok)
	}
	if authInfo.Login == "" {
		return fmt.Errorf("Login is empty in auth info: %v (token: %v)", authInfo, tok)
	}

	// Set user cookies.
	if err := appauth.WriteSessionCookie(w, appauth.Session{AccessToken: tok.AccessToken}); err != nil {
		return err
	}

	returnTo := state.ReturnTo
	if err := returnto.CheckSafe(returnTo); err != nil {
		return err
	}
	if returnTo == "" {
		returnTo = router.Rel.URLToUser(authInfo.Login).String()
	}
	return tmpl.Exec(r, w, "oauth-client/success.html", http.StatusOK, nil, &struct {
		tmpl.Common
		ReturnTo string
	}{
		ReturnTo: returnTo,
	})
}

func withJWKS(r *http.Request, u *url.URL) (*url.URL, error) {
	ctx := httpctx.FromRequest(r)
	// TODO(sqs!): make it so that the app only has the public key in its ctx, not the private key.
	jwks, err := idkey.FromContext(ctx).MarshalJWKSPublicKey()
	if err != nil {
		return nil, err
	}
	v := u.Query()
	v.Set("jwks", string(jwks))
	u.RawQuery = v.Encode()
	return u, nil
}
