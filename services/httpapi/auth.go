package httpapi

import (
	"encoding/json"
	"net/http"

	"gopkg.in/inconshreveable/log15.v2"

	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
	"sourcegraph.com/sqs/pbtypes"
)

type authInfo struct {
	sourcegraph.AuthInfo
	IncludedUser   *sourcegraph.User          `json:",omitempty"`
	IncludedEmails []*sourcegraph.EmailAddr   `json:",omitempty"`
	GitHubToken    *sourcegraph.ExternalToken `json:",omitempty"`
}

func serveAuthInfo(w http.ResponseWriter, r *http.Request) error {
	cl := handlerutil.Client(r)

	info, err := cl.Auth.Identify(r.Context(), &pbtypes.Void{})
	if err != nil {
		return err
	}

	res := authInfo{AuthInfo: *info}

	if info.UID != 0 {
		tok, err := cl.Auth.GetExternalToken(r.Context(), &sourcegraph.ExternalTokenSpec{
			UID:      info.UID,
			Host:     "github.com",
			ClientID: "", // defaults to GitHub client ID in environment
		})
		if err == nil {
			// No need to include the actual access token.
			tok.Token = ""

			res.GitHubToken = tok
		} else if grpc.Code(err) != codes.NotFound {
			log15.Warn("Error getting GitHub token in serveAuthInfo", "uid", info.UID, "err", err)
		}
	}

	// As an optimization, optimistically include the user to avoid
	// the client needing to make another roundtrip.
	if info.UID != 0 {
		user, err := cl.Users.Get(r.Context(), &sourcegraph.UserSpec{UID: info.UID})
		if err == nil {
			res.IncludedUser = user
		} else {
			log15.Warn("Error optimistically including user in serveAuthInfo", "uid", info.UID, "err", err)
		}
	}

	// Also optimistically include emails
	if info.UID != 0 {
		emails, err := cl.Users.ListEmails(r.Context(), &sourcegraph.UserSpec{UID: info.UID})
		if err == nil {
			res.IncludedEmails = emails.EmailAddrs
		} else {
			log15.Warn("Error optimistically including emails in serveAuthInfo", "uid", info.UID, "err", err)
		}
	}

	return writeJSON(w, res)
}

type authResponse struct {
	Success bool

	// AccessToken is the Sourcegraph access token. It is only set
	// after signing up or logging in successfully through the HTTP
	// API.
	AccessToken string `json:",omitempty"`
}

func serveLogin(w http.ResponseWriter, r *http.Request) error {
	cl := handlerutil.Client(r)

	loginForm := struct {
		sourcegraph.LoginCredentials
	}{}
	if err := json.NewDecoder(r.Body).Decode(&loginForm); err != nil {
		return err
	}
	defer r.Body.Close()

	return finishLoginOrSignup(r.Context(), cl, w, r, loginForm.Login, loginForm.Password)
}

func serveSignup(w http.ResponseWriter, r *http.Request) error {
	cl := handlerutil.Client(r)

	signupForm := struct {
		sourcegraph.NewAccount
	}{}
	if err := json.NewDecoder(r.Body).Decode(&signupForm); err != nil {
		return err
	}
	defer r.Body.Close()

	_, err := cl.Accounts.Create(r.Context(), &signupForm.NewAccount)
	if err != nil {
		return err
	}

	return finishLoginOrSignup(r.Context(), cl, w, r, signupForm.Login, signupForm.Password)
}

func finishLoginOrSignup(ctx context.Context, cl *sourcegraph.Client, w http.ResponseWriter, r *http.Request, login, password string) error {
	// Get the newly created user's API key to authenticate future requests.
	tok, err := cl.Auth.GetAccessToken(ctx, &sourcegraph.AccessTokenRequest{
		AuthorizationGrant: &sourcegraph.AccessTokenRequest_ResourceOwnerPassword{
			ResourceOwnerPassword: &sourcegraph.LoginCredentials{Login: login, Password: password},
		},
	})
	if err != nil {
		return err
	}

	authInfo, err := cl.Auth.Identify(sourcegraph.WithAccessToken(ctx, tok.AccessToken), &pbtypes.Void{})
	if err != nil {
		return err
	}

	// Authenticate as newly created user.
	if err := auth.StartNewSession(w, r, authInfo.UID); err != nil {
		return err
	}

	return writeJSON(w, &authResponse{Success: true, AccessToken: tok.AccessToken})
}

func serveForgotPassword(w http.ResponseWriter, r *http.Request) error {
	cl := handlerutil.Client(r)

	form := struct {
		Email string
	}{}
	if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
		return err
	}
	defer r.Body.Close()

	_, err := cl.Accounts.RequestPasswordReset(r.Context(), &sourcegraph.RequestPasswordResetOp{Email: form.Email})
	if err != nil {
		return err
	}

	return writeJSON(w, &authResponse{Success: true})
}

func servePasswordReset(w http.ResponseWriter, r *http.Request) error {
	cl := handlerutil.Client(r)

	form := struct {
		Password        string
		ConfirmPassword string
		Token           string
	}{}
	if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
		return err
	}
	defer r.Body.Close()

	if form.ConfirmPassword != form.Password {
		return grpc.Errorf(codes.InvalidArgument, "passwords do not match")
	}

	_, err := cl.Accounts.ResetPassword(r.Context(), &sourcegraph.NewPassword{Password: form.Password, Token: &sourcegraph.PasswordResetToken{Token: form.Token}})
	if err != nil {
		return err
	}

	return writeJSON(w, &authResponse{Success: true})
}

func serveGitHubToken(w http.ResponseWriter, r *http.Request) error {
	cl := handlerutil.Client(r)

	info, err := cl.Auth.Identify(r.Context(), &pbtypes.Void{})
	if err != nil {
		return err
	}

	if info.UID == 0 {
		return grpc.Errorf(codes.Unauthenticated, "not logged in")
	}

	tok, err := cl.Auth.GetExternalToken(r.Context(), &sourcegraph.ExternalTokenSpec{
		UID:      info.UID,
		Host:     "github.com",
		ClientID: "", // defaults to GitHub client ID in environment
	})
	if err != nil {
		return err
	}

	return writeJSON(w, tok)
}
