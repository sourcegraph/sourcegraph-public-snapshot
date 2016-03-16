package localauth

import (
	"errors"
	"net/http"

	"golang.org/x/oauth2"
	"google.golang.org/grpc/codes"

	appauth "sourcegraph.com/sourcegraph/sourcegraph/app/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/form"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/returnto"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/schemautil"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/tmpl"
	"sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/auth/authutil"
	"sourcegraph.com/sourcegraph/sourcegraph/errcode"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/util/eventsutil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/httputil/httpctx"
	"sourcegraph.com/sqs/pbtypes"
)

func init() {
	internal.Handlers[router.LogIn] = serveLogIn
}

type loginForm struct {
	sourcegraph.LoginCredentials
	form.Validation
}

func (f *loginForm) Validate() {
	if f.Login == "" {
		f.AddFieldError("Login", "Empty username.")
	}
	if f.Password == "" {
		f.AddFieldError("Password", "Empty password.")
	}
}

func serveLogIn(w http.ResponseWriter, r *http.Request) error {
	if err := checkLoginEnabled(); err != nil {
		return err
	}

	ctx := httpctx.FromRequest(r)
	u := handlerutil.UserFromContext(ctx)
	if u != nil && u.UID != 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return nil
	}

	switch r.Method {
	case "GET":
		return serveLoginForm(w, r, loginForm{})
	case "POST":
		return serveLoginSubmit(w, r)
	}
	http.Error(w, "", http.StatusMethodNotAllowed)
	return nil
}

func serveLoginForm(w http.ResponseWriter, r *http.Request, form loginForm) error {
	ctx, cl := handlerutil.Client(r)

	numUsers, err := cl.Users.Count(ctx, &pbtypes.Void{})
	if err != nil {
		return err
	}

	if numUsers.Count == 0 && authutil.ActiveFlags.IsLocal() {
		http.Redirect(w, r, "/join", http.StatusSeeOther)
		return nil
	}

	return tmpl.Exec(r, w, "user/login.html", http.StatusOK, nil, &struct {
		LoginForm loginForm
		FirstUser bool
		tmpl.Common
	}{
		LoginForm: form,
		FirstUser: (numUsers.Count == 0),
	})
}

func serveLoginSubmit(w http.ResponseWriter, r *http.Request) error {
	ctx, cl := handlerutil.Client(r)

	var form loginForm
	if err := r.ParseForm(); err != nil {
		return err
	}
	if err := schemautil.Decode(&form, r.PostForm); err != nil {
		return err
	}

	form.Validate()
	if form.HasErrors() {
		return serveLoginForm(w, r, form)
	}

	tok, err := cl.Auth.GetAccessToken(ctx, &sourcegraph.AccessTokenRequest{
		AuthorizationGrant: &sourcegraph.AccessTokenRequest_ResourceOwnerPassword{
			ResourceOwnerPassword: &form.LoginCredentials,
		},
	})
	if err != nil {
		switch errcode.GRPC(err) {
		case codes.InvalidArgument:
			form.AddFieldError("Login", formErrorInvalidUsername)
		case codes.NotFound:
			form.AddFieldError("Login", formErrorNoUserExists)
		case codes.PermissionDenied:
			form.AddFieldError("Password", formErrorWrongPassword)

		default:
			return err
		}

		// Re-render form.
		return serveLoginForm(w, r, form)
	}

	// Authenticate future requests.
	ctx = sourcegraph.WithCredentials(ctx, oauth2.StaticTokenSource(&oauth2.Token{TokenType: "Bearer", AccessToken: tok.AccessToken}))

	// Authenticate as newly created user.
	if err := appauth.WriteSessionCookie(w, appauth.Session{AccessToken: tok.AccessToken}); err != nil {
		return err
	}

	eventsutil.LogSignIn(ctx)

	returnTo, err := returnto.ExactURLFromQuery(r)
	if err != nil {
		return err
	}
	if returnTo == "" {
		returnTo = "/"
	}

	http.Redirect(w, r, returnTo, http.StatusSeeOther)
	return nil
}

const (
	formErrorInvalidUsername = "Invalid username (bad format or not whitelisted on this server)."
	formErrorNoUserExists    = "No user exists with this username."
	formErrorNoEmailExists   = "No user exists with this email address."
	formErrorWrongPassword   = "Wrong password."
)

func checkLoginEnabled() error {
	if !authutil.ActiveFlags.HasLogin() {
		return &errcode.HTTPErr{Status: http.StatusNotFound, Err: errors.New("login not enabled")}
	}
	return nil
}
