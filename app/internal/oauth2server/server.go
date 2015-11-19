package oauth2server

import (
	"errors"
	"net/http"
	"net/url"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/google/go-querystring/query"

	"strings"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/app/internal"
	appauthutil "src.sourcegraph.com/sourcegraph/app/internal/authutil"
	"src.sourcegraph.com/sourcegraph/app/internal/schemautil"
	"src.sourcegraph.com/sourcegraph/app/internal/tmpl"
	"src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/auth/authutil"
	"src.sourcegraph.com/sourcegraph/errcode"
	"src.sourcegraph.com/sourcegraph/pkg/oauth2util"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
	"src.sourcegraph.com/sourcegraph/util/httputil"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
)

func init() {
	internal.Handlers[router.OAuth2ServerAuthorize] = serveOAuth2ServerAuthorize
	internal.Handlers[router.OAuth2ServerToken] = serveOAuth2ServerToken
}

func serveOAuth2ServerAuthorize(w http.ResponseWriter, r *http.Request) error {
	cl := handlerutil.APIClient(r)
	ctx := httpctx.FromRequest(r)

	if !authutil.ActiveFlags.OAuth2AuthServer {
		return &errcode.HTTPErr{Status: http.StatusNotFound, Err: errors.New("oauth2 auth server disabled")}
	}

	currentUser := handlerutil.UserFromRequest(r)
	if currentUser == nil {
		return appauthutil.RedirectToLogIn(w, r)
	}

	var opt oauth2util.AuthorizeParams
	if err := schemautil.Decode(&opt, r.URL.Query()); err != nil {
		return err
	}

	// Look up registered API client based on client ID. If the client
	// isn't yet registered, redirect to the client registration page.
	if opt.ClientID == "" {
		return &errcode.HTTPErr{Status: http.StatusBadRequest, Err: errors.New("no client ID")}
	}
	regClient, err := cl.RegisteredClients.Get(ctx, &sourcegraph.RegisteredClientSpec{ID: opt.ClientID})
	if grpc.Code(err) == codes.NotFound {
		u := router.Rel.URLTo(router.RegisterClient)
		q, err := query.Values(opt)
		if err != nil {
			return err
		}
		u.RawQuery = q.Encode()
		http.Redirect(w, r, u.String(), http.StatusSeeOther)
		return nil
	}
	if err != nil {
		return &errcode.HTTPErr{Status: http.StatusBadRequest, Err: err}
	}

	// Create code that will enable the OAuth2 client to obtain the
	// user's access token. This also checks the redirect URI.
	code, err := cl.Auth.GetAuthorizationCode(ctx, &sourcegraph.AuthorizationCodeRequest{
		ResponseType: "code",
		ClientID:     opt.ClientID,
		RedirectURI:  opt.RedirectURI,
		Scope:        strings.Fields(opt.Scope),
		UID:          currentUser.UID,
	})
	if err != nil {
		if grpc.Code(err) == codes.PermissionDenied {
			// User is not granted access to server.
			return tmpl.Exec(r, w,
				"oauth-provider/auth_required.html",
				http.StatusForbidden,
				http.Header{"cache-control": []string{"no-cache"}},
				&struct {
					RegClient *sourcegraph.RegisteredClient
					Err       error
					tmpl.Common
				}{
					RegClient: regClient,
					Err:       err,
				},
			)
		}
		return err
	}

	// Generate redirect URL to send the resource owner back to the
	// client.
	redirectURL, err := url.Parse(code.RedirectURI)
	if err != nil {
		return &errcode.HTTPErr{Status: http.StatusBadRequest, Err: err}
	}
	q, err := query.Values(oauth2util.ReceiveParams{
		ClientID: regClient.ID,
		State:    opt.State,
		Code:     code.Code,
		Scope:    opt.Scope,
	})
	if err != nil {
		return err
	}
	redirectURL.RawQuery = q.Encode()

	return tmpl.Exec(r, w, "oauth-provider/authorize.html", http.StatusOK, nil, &struct {
		tmpl.Common
		RegisteredClient *sourcegraph.RegisteredClient
		RedirectURL      *url.URL
	}{
		RegisteredClient: regClient,
		RedirectURL:      redirectURL,
	})
}

func serveOAuth2ServerToken(w http.ResponseWriter, r *http.Request) error {
	cl := handlerutil.APIClient(r)
	ctx := httpctx.FromRequest(r)

	if !authutil.ActiveFlags.OAuth2AuthServer {
		return &errcode.HTTPErr{Status: http.StatusNotFound, Err: errors.New("oauth2 auth server disabled")}
	}

	currentUser := handlerutil.UserFromRequest(r)
	if currentUser != nil {
		return appauthutil.RedirectToLogIn(w, r)
	}

	if err := r.ParseForm(); err != nil {
		return err
	}

	var opt oauth2util.TokenParams
	if err := schemautil.Decode(&opt, r.PostForm); err != nil {
		return err
	}

	var tokReq sourcegraph.AccessTokenRequest

	switch opt.GrantType {
	case "authorization_code":
		tokReq.AuthorizationGrant = &sourcegraph.AccessTokenRequest_AuthorizationCode{
			AuthorizationCode: &sourcegraph.AuthorizationCode{
				Code:        opt.Code,
				RedirectURI: opt.RedirectURI,
			},
		}

	case "urn:ietf:params:oauth:grant-type:jwt-bearer":
		tokReq.AuthorizationGrant = &sourcegraph.AccessTokenRequest_BearerJWT{
			BearerJWT: &sourcegraph.BearerJWT{
				Assertion: opt.Assertion,
			},
		}
	}

	atok, err := cl.Auth.GetAccessToken(ctx, &tokReq)
	if err != nil {
		return err
	}

	type accessTokenResponse struct {
		AccessToken  string `json:"access_token,omitempty"`
		RefreshToken string `json:"refresh_token,omitempty"`
		ExpiresIn    int32  `json:"expires_in,omitempty"`
		TokenType    string `json:"token_type,omitempty"`
	}
	return httputil.WriteJSON(w, &accessTokenResponse{
		AccessToken: atok.AccessToken,
		TokenType:   atok.TokenType,
		ExpiresIn:   atok.ExpiresInSec,
	})
}
