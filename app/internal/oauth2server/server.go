package oauth2server

import (
	"net/http"

	"sourcegraph.com/sourcegraph/sourcegraph/app/internal"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/schemautil"
	"sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/oauth2util"
	"sourcegraph.com/sourcegraph/sourcegraph/util/errcode"
	"sourcegraph.com/sourcegraph/sourcegraph/util/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/httputil"
)

func init() {
	internal.Handlers[router.OAuth2ServerToken] = serveOAuth2ServerToken
}

func serveOAuth2ServerToken(w http.ResponseWriter, r *http.Request) error {
	ctx, cl := handlerutil.Client(r)

	currentUser := handlerutil.UserFromRequest(r)
	if currentUser != nil {
		return &errcode.HTTPErr{Status: http.StatusUnauthorized}
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
