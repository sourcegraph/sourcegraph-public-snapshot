package backend

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	authpkg "sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth/accesstoken"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth/idkey"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/githubutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
	"sourcegraph.com/sourcegraph/sourcegraph/services/ext/github"
	"sourcegraph.com/sourcegraph/sourcegraph/services/ext/github/githubcli"
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
	} else if githubAuthCode := op.GetGitHubAuthCode(); githubAuthCode != nil {
		return s.authenticateGitHubAuthCode(ctx, githubAuthCode)
	} else {
		return nil, grpc.Errorf(codes.Unauthenticated, "no supported auth credentials provided")
	}
}

func (s *auth) authenticateLogin(ctx context.Context, cred *sourcegraph.LoginCredentials) (*sourcegraph.AccessTokenResponse, error) {
	usersStore := store.UsersFromContext(ctx)

	if cred.Login == "" {
		return nil, grpc.Errorf(codes.InvalidArgument, "login cannot be empty")
	}

	if cred.Password == "" {
		return nil, grpc.Errorf(codes.InvalidArgument, "password cannot be empty")
	}

	user, err := usersStore.Get(elevatedActor(ctx), sourcegraph.UserSpec{Login: cred.Login})
	if err != nil {
		return nil, grpc.Errorf(codes.PermissionDenied, "user not found")
	}

	if store.PasswordFromContext(ctx).CheckUIDPassword(elevatedActor(ctx), user.UID, cred.Password) != nil {
		return nil, grpc.Errorf(codes.PermissionDenied, "incorrect password for %q", cred.Login)
	}

	a := authpkg.ActorFromContext(ctx)
	if a.UID != 0 {
		return nil, grpc.Errorf(codes.PermissionDenied, "refusing to issue access token from resource owner password to already authenticated user %d (only client, not user, must be authenticated)", a.UID)
	}

	a.UID = int(user.UID)
	a.Login = user.Login
	a.Write = user.Write
	a.Admin = user.Admin

	tok, err := accesstoken.New(idkey.FromContext(ctx), &a, nil, 7*24*time.Hour, true)

	if err != nil {
		return nil, err
	}

	return accessTokenToTokenResponse(tok), nil
}

func (s *auth) authenticateGitHubAuthCode(ctx context.Context, authCode *sourcegraph.GitHubAuthCode) (*sourcegraph.AccessTokenResponse, error) {
	if authCode.Host != "github.com" {
		return nil, grpc.Errorf(codes.Unimplemented, "non-github.com hosts not yet supported")
	}
	if authCode.Code == "" {
		return nil, grpc.Errorf(codes.InvalidArgument, "code cannot be empty")
	}

	// Exchange the code for a GitHub access token.
	ghToken, err := githubutil.Default.OAuth.Exchange(oauth2.NoContext, authCode.Code)
	if err != nil {
		return nil, err
	}
	if !ghToken.Valid() {
		return nil, grpc.Errorf(codes.PermissionDenied, "exchanging auth code yielded invalid GitHub OAuth2 token")
	}

	// Get the current user.
	gh := githubutil.Default.AuthedClient(ghToken.AccessToken)
	ghUser, _, err := gh.Users.Get("")
	if err != nil {
		return nil, err
	}

	// Check if this corresponds to an existing Sourcegraph user; if
	// so, return that info in the response to allow the app to know
	// when it needs to create a new Sourcegraph user.
	var resp *sourcegraph.AccessTokenResponse
	uid, err := store.UsersFromContext(ctx).GetUIDByGitHubID(ctx, *ghUser.ID)
	if err == nil {
		user, err := (&users{}).Get(ctx, &sourcegraph.UserSpec{UID: uid})
		if err != nil {
			return nil, err
		}

		// Generate the Sourcegraph access token. We only do this if
		// there's an existing Sourcegraph user for this GitHub access
		// token; otherwise, the Sourcegraph access token would be
		// anonymous, which'd be pointless.
		a := authpkg.ActorFromContext(ctx)
		a.UID = int(user.UID)
		a.Login = user.Login
		a.Write = user.Write
		a.Admin = user.Admin
		tok, err := accesstoken.New(idkey.FromContext(ctx), &a, nil, 7*24*time.Hour, true)
		if err != nil {
			return nil, err
		}
		resp = accessTokenToTokenResponse(tok)
		resp.UID = uid
	} else if grpc.Code(err) == codes.NotFound {
		// Do nothing.
	} else if err != nil {
		return nil, err
	}

	if resp == nil {
		resp = &sourcegraph.AccessTokenResponse{}
	}

	if scope, ok := ghToken.Extra("scope").(string); ok && scope != "" {
		resp.Scope = strings.Split(scope, ",")
	}

	resp.GitHubAccessToken = ghToken.AccessToken
	resp.GitHubUser = &sourcegraph.GitHubUser{
		ID:    int32(*ghUser.ID),
		Login: *ghUser.Login,
	}
	if ghUser.Name != nil {
		resp.GitHubUser.Name = *ghUser.Name
	}
	if ghUser.Email != nil {
		resp.GitHubUser.Email = *ghUser.Email
	}
	if ghUser.Location != nil {
		resp.GitHubUser.Location = *ghUser.Location
	}
	if ghUser.Company != nil {
		resp.GitHubUser.Company = *ghUser.Company
	}
	if ghUser.AvatarURL != nil {
		resp.GitHubUser.AvatarURL = *ghUser.AvatarURL
	}

	return resp, nil
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

		IntercomHash: intercomHMAC(a.UID),
	}, nil
}

func (s *auth) GetExternalToken(ctx context.Context, tok *sourcegraph.ExternalTokenSpec) (*sourcegraph.ExternalToken, error) {
	if tok == nil {
		tok = &sourcegraph.ExternalTokenSpec{}
	}
	extTokensStore := store.ExternalAuthTokensFromContext(ctx)

	setExternalTokenSpecDefaults(ctx, tok, nil)

	dbToken, err := extTokensStore.GetUserToken(ctx, int(tok.UID), tok.Host, tok.ClientID)
	if err == store.ErrNoExternalAuthToken {
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

	setExternalTokenSpecDefaults(ctx, nil, extToken)

	dbToken := &store.ExternalAuthToken{
		User:     int(extToken.UID),
		Host:     extToken.Host,
		Token:    extToken.Token,
		Scope:    extToken.Scope,
		ClientID: extToken.ClientID,
		ExtUID:   int(extToken.ExtUID),
	}

	err := extTokensStore.SetUserToken(ctx, dbToken)
	return &pbtypes.Void{}, err
}

func (s *auth) DeleteAndRevokeExternalToken(ctx context.Context, tokSpec *sourcegraph.ExternalTokenSpec) (*pbtypes.Void, error) {
	store := store.ExternalAuthTokensFromContext(ctx)

	setExternalTokenSpecDefaults(ctx, tokSpec, nil)

	tok, err := store.GetUserToken(ctx, int(tokSpec.UID), tokSpec.Host, tokSpec.ClientID)
	if err != nil {
		return nil, err
	}

	if err := store.DeleteToken(ctx, tokSpec); err != nil {
		return nil, err
	}

	// Revoke token if it's from GitHub.
	if tok.Host == githubcli.Config.Host() && tok.ClientID == githubClientID {
		if err := (&github.Authorizations{}).Revoke(ctx, tok.ClientID, tok.Token); err != nil {
			return nil, err
		}
	}

	return &pbtypes.Void{}, err
}

func setExternalTokenSpecDefaults(ctx context.Context, tokSpec *sourcegraph.ExternalTokenSpec, tok *sourcegraph.ExternalToken) {
	// Exact same logic for both ExternalToken and ExternalTokenSpec, so
	// handle both here. Usually you'll only pass in one or the other.

	if tok != nil && tok.ClientID == "" {
		tok.ClientID = githubClientID
	}
	if tokSpec != nil && tokSpec.ClientID == "" {
		tokSpec.ClientID = githubClientID
	}

	if tok != nil && tok.Host == "" {
		tok.Host = githubcli.Config.Host()
	}
	if tokSpec != nil && tokSpec.Host == "" {
		tokSpec.Host = githubcli.Config.Host()
	}

	if tok != nil && tok.UID == 0 {
		tok.UID = int32(authpkg.ActorFromContext(ctx).UID)
	}
	if tokSpec != nil && tokSpec.UID == 0 {
		tokSpec.UID = int32(authpkg.ActorFromContext(ctx).UID)
	}
}

var intercomSecretKey = os.Getenv("SG_INTERCOM_SECRET_KEY")

func intercomHMAC(uid int) string {
	if uid == 0 || intercomSecretKey == "" {
		return ""
	}
	mac := hmac.New(sha256.New, []byte(intercomSecretKey))
	mac.Write([]byte(strconv.Itoa(uid)))
	return hex.EncodeToString(mac.Sum(nil))
}
