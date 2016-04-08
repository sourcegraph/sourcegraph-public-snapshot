package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"

	appauth "sourcegraph.com/sourcegraph/sourcegraph/app/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/util/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/httputil/httpctx"
)

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
	u := handlerutil.UserFromContext(ctx)

	if u != nil && u.UID != 0 {
		return fmt.Errorf("Cannot signup while logged in.")
	}

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
	u := handlerutil.UserFromContext(ctx)

	if u == nil || u.UID == 0 {
		// SECURITY: If there is no authenticated user in the context, serve an
		// error.
		//
		// This prevents CSRF attacks which allow external sites to log the
		// user out by having them submit a form etc. Not a huge threat in the
		// usual case, but it would still log users out and annoy them. Do not
		// allow it.
		return fmt.Errorf("cannot log out (no logged in user in context)")
	}

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
	u := handlerutil.UserFromContext(ctx)

	if u != nil && u.UID != 0 {
		return fmt.Errorf("Cannot reset password while logged in.")
	}

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
	u := handlerutil.UserFromContext(ctx)

	if u != nil && u.UID != 0 {
		return fmt.Errorf("Cannot reset password while logged in.")
	}

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
		return fmt.Errorf("Your password must be the same as the confirmation")
	}

	_, err := cl.Accounts.ResetPassword(ctx, &sourcegraph.NewPassword{Password: form.Password, Token: &sourcegraph.PasswordResetToken{Token: form.Token}})
	if err != nil {
		return fmt.Errorf("error reseting password: %s", err)
	}

	return writeJSON(w, &authResponse{Success: true})
}
