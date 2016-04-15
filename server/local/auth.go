package local

import (
	"math"
	"os"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	authpkg "sourcegraph.com/sourcegraph/sourcegraph/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/auth/accesstoken"
	"sourcegraph.com/sourcegraph/sourcegraph/auth/idkey"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/services/ext/github/githubcli"
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

func (s *auth) GetAccessToken(ctx context.Context, op *sourcegraph.AccessTokenRequest) (*sourcegraph.AccessTokenResponse, error) {
	if resOwnerPassword := op.GetResourceOwnerPassword(); resOwnerPassword != nil {
		return s.authenticateLogin(ctx, resOwnerPassword)
	} else {
		return nil, grpc.Errorf(codes.Unauthenticated, "no supported auth credentials provided")
	}
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
	if a.UID != 0 {
		return nil, grpc.Errorf(codes.PermissionDenied, "refusing to issue access token from resource owner password to already authenticated user %d (only client, not user, must be authenticated)", a.UID)
	}

	a.UID = int(user.UID)
	a.Login = user.Login
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
		UID:   int32(a.UID),
		Login: a.Login,

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
