package httpapi

import (
	"encoding/json"
	"net/http"

	"gopkg.in/inconshreveable/log15.v2"

	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	appauth "sourcegraph.com/sourcegraph/sourcegraph/app/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/util/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/httputil/httpctx"
	"sourcegraph.com/sqs/pbtypes"
)

type authInfo struct {
	sourcegraph.AuthInfo
	IncludedUser *sourcegraph.User `json:",omitempty"`
}

func serveAuthInfo(w http.ResponseWriter, r *http.Request) error {
	ctx, cl := handlerutil.Client(r)

	info, err := cl.Auth.Identify(ctx, &pbtypes.Void{})
	if err != nil {
		return err
	}

	res := authInfo{AuthInfo: *info}

	// As an optimization, optimistically include the user to avoid
	// the client needing to make another roundtrip.
	if info.UID != 0 {
		user, err := cl.Users.Get(ctx, &sourcegraph.UserSpec{UID: info.UID})
		if err == nil {
			res.IncludedUser = user
		} else {
			log15.Warn("Error optimistically including user in serveAuthInfo", "uid", info.UID, "err", err)
		}
	}

	return writeJSON(w, res)
}

type authResponse struct {
	Success bool
}

func serveLogin(w http.ResponseWriter, r *http.Request) error {
	ctx, cl := handlerutil.Client(r)

	loginForm := struct {
		sourcegraph.LoginCredentials
	}{}
	if err := json.NewDecoder(r.Body).Decode(&loginForm); err != nil {
		return err
	}
	defer r.Body.Close()

	tok, err := cl.Auth.GetAccessToken(ctx, &sourcegraph.AccessTokenRequest{
		AuthorizationGrant: &sourcegraph.AccessTokenRequest_ResourceOwnerPassword{
			ResourceOwnerPassword: &sourcegraph.LoginCredentials{
				Login:    loginForm.Login,
				Password: loginForm.Password,
			},
		},
	})
	if err != nil {
		return err
	}

	// Authenticate future requests.
	ctx = sourcegraph.WithCredentials(ctx, oauth2.StaticTokenSource(&oauth2.Token{TokenType: "Bearer", AccessToken: tok.AccessToken}))

	// Authenticate as newly created user.
	if err := appauth.WriteSessionCookie(w, appauth.Session{AccessToken: tok.AccessToken}); err != nil {
		return err
	}

	return writeJSON(w, &authResponse{Success: true})
}

func serveSignup(w http.ResponseWriter, r *http.Request) error {
	ctx, cl := handlerutil.Client(r)

	signupForm := struct {
		sourcegraph.NewAccount
	}{}
	if err := json.NewDecoder(r.Body).Decode(&signupForm); err != nil {
		return err
	}
	defer r.Body.Close()

	_, err := cl.Accounts.Create(ctx, &signupForm.NewAccount)
	if err != nil {
		return err
	}

	// Get the newly created user's API key to authenticate future requests.
	tok, err := cl.Auth.GetAccessToken(ctx, &sourcegraph.AccessTokenRequest{
		AuthorizationGrant: &sourcegraph.AccessTokenRequest_ResourceOwnerPassword{
			ResourceOwnerPassword: &sourcegraph.LoginCredentials{Login: signupForm.Login, Password: signupForm.Password},
		},
	})
	if err != nil {
		return err
	}

	// Authenticate future requests.
	ctx = sourcegraph.WithCredentials(ctx, oauth2.StaticTokenSource(&oauth2.Token{TokenType: "Bearer", AccessToken: tok.AccessToken}))

	// Authenticate as newly created user.
	if err := appauth.WriteSessionCookie(w, appauth.Session{AccessToken: tok.AccessToken}); err != nil {
		return err
	}

	return writeJSON(w, &authResponse{Success: true})
}

func serveLogout(w http.ResponseWriter, r *http.Request) error {
	ctx := httpctx.FromRequest(r)

	// Delete their session.
	appauth.DeleteSessionCookie(w)

	// Clear the user in the request context so that we don't show the logout
	// page with the user's info.
	ctx = handlerutil.ClearUser(ctx)
	httpctx.SetForRequest(r, ctx)

	return writeJSON(w, &authResponse{Success: true})
}

func serveForgotPassword(w http.ResponseWriter, r *http.Request) error {
	ctx, cl := handlerutil.Client(r)

	form := struct {
		Email string
	}{}
	if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
		return err
	}
	defer r.Body.Close()

	_, err := cl.Accounts.RequestPasswordReset(ctx, &sourcegraph.PersonSpec{Email: form.Email})
	if err != nil {
		return err
	}

	return writeJSON(w, &authResponse{Success: true})
}

func servePasswordReset(w http.ResponseWriter, r *http.Request) error {
	ctx, cl := handlerutil.Client(r)

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

	_, err := cl.Accounts.ResetPassword(ctx, &sourcegraph.NewPassword{Password: form.Password, Token: &sourcegraph.PasswordResetToken{Token: form.Token}})
	if err != nil {
		return err
	}

	return writeJSON(w, &authResponse{Success: true})
}
