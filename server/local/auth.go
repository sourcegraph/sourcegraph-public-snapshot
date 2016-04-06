package local

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"sourcegraph.com/sourcegraph/sourcegraph/app/router"
	authpkg "sourcegraph.com/sourcegraph/sourcegraph/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/auth/accesstoken"
	"sourcegraph.com/sourcegraph/sourcegraph/auth/idkey"
	"sourcegraph.com/sourcegraph/sourcegraph/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/oauth2util"
	"sourcegraph.com/sourcegraph/sourcegraph/services/ext/github/githubcli"
	"sourcegraph.com/sourcegraph/sourcegraph/services/svc"
	"sourcegraph.com/sourcegraph/sourcegraph/store"
	"sourcegraph.com/sqs/pbtypes"
)

var (
	githubClientID string
)

func init() {
	githubClientID = os.Getenv("GITHUB_CLIENT_ID")
}

var Auth sourcegraph.AuthServer = &auth{}

type auth struct{}

var _ sourcegraph.AuthServer = (*auth)(nil)

func (s *auth) GetAuthorizationCode(ctx context.Context, op *sourcegraph.AuthorizationCodeRequest) (*sourcegraph.AuthorizationCode, error) {
	authStore := store.AuthorizationsFromContext(ctx)

	if op.ResponseType != "code" {
		return nil, grpc.Errorf(codes.InvalidArgument, "invalid response_type")
	}

	client, err := svc.RegisteredClients(ctx).Get(ctx, &sourcegraph.RegisteredClientSpec{ID: op.ClientID})
	if err != nil {
		return nil, err
	}

	// RedirectURI is OPTIONAL
	// (https://tools.ietf.org/html/rfc6749#section-4.1.1) but must be
	// validated if set.
	if op.RedirectURI != "" {
		if err := oauth2util.AllowRedirectURI(client.RedirectURIs, op.RedirectURI); err != nil {
			return nil, err
		}
	}

	code, err := authStore.CreateAuthCode(ctx, op, 5*time.Minute)
	if err != nil {
		return nil, err
	}
	return &sourcegraph.AuthorizationCode{Code: code, RedirectURI: op.RedirectURI}, nil
}

func (s *auth) GetAccessToken(ctx context.Context, op *sourcegraph.AccessTokenRequest) (*sourcegraph.AccessTokenResponse, error) {
	if authCode := op.GetAuthorizationCode(); authCode != nil {
		return s.exchangeCodeForAccessToken(ctx, authCode)
	} else if resOwnerPassword := op.GetResourceOwnerPassword(); resOwnerPassword != nil {
		return s.authenticateLogin(ctx, resOwnerPassword)
	} else if bearerJWT := op.GetBearerJWT(); bearerJWT != nil {
		return s.authenticateBearerJWT(ctx, bearerJWT)
	} else {
		return nil, grpc.Errorf(codes.Unauthenticated, "no supported auth credentials provided")
	}
}

func (s *auth) exchangeCodeForAccessToken(ctx context.Context, code *sourcegraph.AuthorizationCode) (*sourcegraph.AccessTokenResponse, error) {
	authStore := store.AuthorizationsFromContext(ctx)

	usersStore := store.UsersFromContext(ctx)

	clientID := authpkg.ActorFromContext(ctx).ClientID
	client, err := svc.RegisteredClients(ctx).Get(ctx, &sourcegraph.RegisteredClientSpec{ID: clientID})
	if err != nil {
		return nil, err
	}

	// RedirectURI is REQUIRED if one was provided when the code was
	// created (https://tools.ietf.org/html/rfc6749#section-4.1.3).
	if code.RedirectURI != "" {
		if err := oauth2util.AllowRedirectURI(client.RedirectURIs, code.RedirectURI); err != nil {
			return nil, err
		}
	}

	req, err := authStore.MarkExchanged(ctx, code, authpkg.ActorFromContext(ctx).ClientID)
	if err != nil {
		return nil, err
	}

	user, err := usersStore.Get(ctx, sourcegraph.UserSpec{UID: req.UID})
	if err != nil {
		return nil, err
	}

	tok, err := accesstoken.New(idkey.FromContext(ctx), authpkg.Actor{
		UID:      int(user.UID),
		Login:    user.Login,
		ClientID: req.ClientID,
		Scope:    authpkg.UnmarshalScope(req.Scope),
	}, map[string]string{"GrantType": "AuthorizationCode"}, 7*24*time.Hour)
	if err != nil {
		return nil, err
	}

	return accessTokenToTokenResponse(tok), nil
}

func (s *auth) authenticateLogin(ctx context.Context, cred *sourcegraph.LoginCredentials) (*sourcegraph.AccessTokenResponse, error) {
	usersStore := store.UsersFromContext(ctx)

	user, err := usersStore.Get(elevatedActor(ctx), sourcegraph.UserSpec{Login: cred.Login})
	if err != nil {
		return nil, err
	}

	if store.PasswordFromContext(ctx).CheckUIDPassword(elevatedActor(ctx), user.UID, cred.Password) != nil {
		return nil, grpc.Errorf(codes.PermissionDenied, "bad password for user %q", cred.Login)
	}

	a := authpkg.ActorFromContext(ctx)
	if a.IsUser() {
		return nil, grpc.Errorf(codes.PermissionDenied, "refusing to issue access token from resource owner password to already authenticated user %d (only client, not user, must be authenticated)", a.UID)
	}

	a.UID = int(user.UID)
	a.Login = user.Login
	a.ClientID = idkey.FromContext(ctx).ID
	a.Write = user.Write
	a.Admin = user.Admin

	tok, err := accesstoken.New(
		idkey.FromContext(ctx),
		a,
		map[string]string{"GrantType": "ResourceOwnerPassword"},
		7*24*time.Hour,
	)

	if err != nil {
		return nil, err
	}

	return accessTokenToTokenResponse(tok), nil
}

func (s *auth) authenticateBearerJWT(ctx context.Context, rawTok *sourcegraph.BearerJWT) (*sourcegraph.AccessTokenResponse, error) {
	var regClient *sourcegraph.RegisteredClient
	tok, err := jwt.Parse(rawTok.Assertion, func(tok *jwt.Token) (interface{}, error) {
		// The JWT's "iss" is the client's OAuth2 client ID.
		clientID, _ := tok.Claims["iss"].(string)
		if clientID == "" {
			return nil, errors.New("bearer JWT has empty issuer, can't look up key")
		}
		var err error
		regClient, err = svc.RegisteredClients(ctx).Get(elevatedActor(ctx), &sourcegraph.RegisteredClientSpec{ID: clientID})
		if err != nil {
			return nil, err
		}

		// Get the client's registered public key.
		if regClient.JWKS == "" {
			return nil, fmt.Errorf("client ID %s (identified by bearer JWT) has no JWKS", clientID)
		}
		pubKey, err := idkey.UnmarshalJWKSPublicKey([]byte(regClient.JWKS))
		if err != nil {
			return nil, fmt.Errorf("parsing client ID %s JWKS public key: %s", clientID, err)
		}
		return pubKey, nil
	})
	if err != nil {
		return nil, err
	}

	// Validate claims; see
	// https://tools.ietf.org/html/draft-ietf-oauth-jwt-bearer-12#section-3.
	aud, _ := tok.Claims["aud"].(string)
	tokURL := conf.AppURL(ctx).ResolveReference(router.Rel.URLTo(router.OAuth2ServerToken))
	if subtle.ConstantTimeCompare([]byte(aud), []byte(tokURL.String())) != 1 {
		return nil, grpc.Errorf(codes.PermissionDenied, "bearer JWT aud claim mismatch (JWT %q, server %q)", aud, tokURL)
	}

	atok, err := accesstoken.New(
		idkey.FromContext(ctx),
		authpkg.Actor{ClientID: regClient.ID},
		map[string]string{"GrantType": "BearerJWT"},
		time.Hour,
	)
	if err != nil {
		return nil, err
	}

	return accessTokenToTokenResponse(atok), nil
}

func accessTokenToTokenResponse(t *oauth2.Token) *sourcegraph.AccessTokenResponse {
	if t.AccessToken == "" {
		panic("empty AccessToken")
	}
	if t.TokenType == "" {
		panic("empty TokenType")
	}
	r := &sourcegraph.AccessTokenResponse{
		AccessToken: t.AccessToken,
		TokenType:   t.TokenType,
	}
	if !t.Expiry.IsZero() {
		sec := t.Expiry.Sub(time.Now()) / time.Second
		if sec > math.MaxInt32 {
			sec = math.MaxInt32
		}
		r.ExpiresInSec = int32(sec)
	}
	return r
}

func (s *auth) Identify(ctx context.Context, _ *pbtypes.Void) (*sourcegraph.AuthInfo, error) {
	a := authpkg.ActorFromContext(ctx)
	return &sourcegraph.AuthInfo{
		ClientID: a.ClientID,
		UID:      int32(a.UID),
		Login:    a.Login,

		Write: a.HasWriteAccess(),
		Admin: a.HasAdminAccess(),
	}, nil
}

func (s *auth) GetExternalToken(ctx context.Context, request *sourcegraph.ExternalTokenRequest) (*sourcegraph.ExternalToken, error) {
	if request == nil {
		request = &sourcegraph.ExternalTokenRequest{}
	}
	extTokensStore := store.ExternalAuthTokensFromContext(ctx)

	if request.ClientID == "" {
		request.ClientID = githubClientID
	}

	if request.Host == "" {
		request.Host = githubcli.Config.Host()
	}

	uid := int(request.UID)
	if uid == 0 {
		uid = authpkg.ActorFromContext(ctx).UID
	}

	dbToken, err := extTokensStore.GetUserToken(ctx, uid, request.Host, request.ClientID)
	if err == authpkg.ErrNoExternalAuthToken {
		return nil, grpc.Errorf(codes.NotFound, "no external auth token found")
	} else if err != nil {
		return nil, err
	}

	return &sourcegraph.ExternalToken{
		UID:      int32(dbToken.User),
		Host:     dbToken.Host,
		Token:    dbToken.Token,
		Scope:    dbToken.Scope,
		ClientID: dbToken.ClientID,
		ExtUID:   int32(dbToken.ExtUID),
	}, nil
}

func (s *auth) SetExternalToken(ctx context.Context, extToken *sourcegraph.ExternalToken) (*pbtypes.Void, error) {
	if extToken == nil {
		extToken = &sourcegraph.ExternalToken{}
	}
	extTokensStore := store.ExternalAuthTokensFromContext(ctx)

	if extToken.ClientID == "" {
		extToken.ClientID = githubClientID
	}

	if extToken.Host == "" {
		extToken.Host = githubcli.Config.Host()
	}

	uid := int(extToken.UID)
	if uid == 0 {
		uid = authpkg.ActorFromContext(ctx).UID
	}

	dbToken := &authpkg.ExternalAuthToken{
		User:     uid,
		Host:     extToken.Host,
		Token:    extToken.Token,
		Scope:    extToken.Scope,
		ClientID: extToken.ClientID,
		ExtUID:   int(extToken.ExtUID),
	}

	err := extTokensStore.SetUserToken(ctx, dbToken)
	return &pbtypes.Void{}, err
}
