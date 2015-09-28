package oauth2client

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	appauth "sourcegraph.com/sourcegraph/sourcegraph/app/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/authutil"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/returnto"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/schemautil"
	"sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/auth/idkey"
	"sourcegraph.com/sourcegraph/sourcegraph/client/pkg/oauth2client"
	"sourcegraph.com/sourcegraph/sourcegraph/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/oauth2util"
	"sourcegraph.com/sourcegraph/sourcegraph/util/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/httputil/httpctx"
	"sourcegraph.com/sqs/pbtypes"
)

// TODO(public-release): This is a very WIP OAuth client
// implementation. It doesn't do any verification, only redirection!

func init() {
	internal.Handlers[router.OAuth2ClientReceive] = serveOAuth2ClientReceive
}

// RedirectToOAuth2Authorize generates the OAuth2 authorize URL
// (including a nonce state value, also stored in a cookie) and
// redirects the client to that URL.
func RedirectToOAuth2Authorize(w http.ResponseWriter, r *http.Request) error {
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
	authutil.RedirectToOAuth2Authorize = RedirectToOAuth2Authorize
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
		return &handlerutil.HTTPErr{Status: http.StatusForbidden, Err: errors.New("already logged in")}
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
		return &handlerutil.HTTPErr{
			Status: http.StatusBadRequest,
			Err:    errors.New("no OAuth2 client ID"),
		}
	}
	if clientID != opt.ClientID {
		return &handlerutil.HTTPErr{
			Status: http.StatusForbidden,
			Err:    errors.New("OAuth2 client ID mismatch"),
		}
	}

	// Check the state nonce against what's stored in the cookie (to
	// prevent CSRF).
	var state oauthAuthorizeClientState
	if err := state.UnmarshalText([]byte(opt.State)); err != nil {
		return &handlerutil.HTTPErr{Status: http.StatusBadRequest, Err: err}
	}
	nonce, present := nonceFromCookie(r)
	deleteNonceCookie(w) // prevent reuse of nonce
	if !present || nonce != state.Nonce || nonce == "" {
		return &handlerutil.HTTPErr{Status: http.StatusForbidden, Err: errors.New("invalid state nonce from OAuth2 provider")}
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
		AuthorizationCode: &sourcegraph.AuthorizationCode{
			Code:        opt.Code,
			RedirectURI: origRedirectURI.String(),
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
	user, err := cl.Users.Get(ctx, authInfo.UserSpec())
	if err != nil {
		return err
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
		returnTo = router.Rel.URLToUser(user.Login).String()
	}
	http.Redirect(w, r, returnTo, http.StatusSeeOther)
	return nil
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
