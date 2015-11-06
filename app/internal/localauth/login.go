package localauth

import (
	"errors"
	"net/http"

	"golang.org/x/oauth2"
	"google.golang.org/grpc/codes"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sqs/pbtypes"
	appauth "src.sourcegraph.com/sourcegraph/app/auth"
	"src.sourcegraph.com/sourcegraph/app/internal"
	appauthutil "src.sourcegraph.com/sourcegraph/app/internal/authutil"
	"src.sourcegraph.com/sourcegraph/app/internal/form"
	"src.sourcegraph.com/sourcegraph/app/internal/returnto"
	"src.sourcegraph.com/sourcegraph/app/internal/schemautil"
	"src.sourcegraph.com/sourcegraph/app/internal/tmpl"
	"src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/auth/authutil"
	"src.sourcegraph.com/sourcegraph/errcode"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
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
	if !(authutil.ActiveFlags.IsLocal() || authutil.ActiveFlags.IsLDAP()) {
		return appauthutil.RedirectToLogIn(w, r)
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
	return tmpl.Exec(r, w, "user/login.html", http.StatusOK, nil, &struct {
		LoginForm loginForm
		tmpl.Common
	}{
		LoginForm: form,
	})
}

func serveLoginSubmit(w http.ResponseWriter, r *http.Request) error {
	cl := handlerutil.APIClient(r)
	ctx := httpctx.FromRequest(r)

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

	// Fetch username.
	authInfo, err := cl.Auth.Identify(ctx, &pbtypes.Void{})
	if err != nil {
		return err
	}
	user, err := cl.Users.Get(ctx, authInfo.UserSpec())
	if err != nil {
		return err
	}

	// Authenticate as newly created user.
	if err := appauth.WriteSessionCookie(w, appauth.Session{AccessToken: tok.AccessToken}); err != nil {
		return err
	}

	returnTo, err := returnto.ExactURLFromQuery(r)
	if err != nil {
		return err
	}
	if returnTo == "" {
		returnTo = router.Rel.URLToUser(user.Login).String()
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
		return &handlerutil.HTTPErr{Status: http.StatusNotFound, Err: errors.New("login not enabled")}
	}
	return nil
}
