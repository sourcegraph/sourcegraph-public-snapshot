package localauth

import (
	"errors"
	"net/http"

	"golang.org/x/oauth2"
	"google.golang.org/grpc/codes"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	appauth "sourcegraph.com/sourcegraph/sourcegraph/app/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/form"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/returnto"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/schemautil"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/tmpl"
	"sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/auth/authutil"
	"sourcegraph.com/sourcegraph/sourcegraph/errcode"
	"sourcegraph.com/sourcegraph/sourcegraph/util/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/httputil/httpctx"
)

func init() {
	internal.Handlers[router.SignUp] = serveSignUp
}

type signupForm struct {
	sourcegraph.NewAccount
	form.Validation
}

func (f *signupForm) Validate() {
	if f.Login == "" {
		f.AddFieldError("Login", "")
	}
	if f.Email == "" {
		f.AddFieldError("Email", "")
	}
	if f.Password == "" {
		f.AddFieldError("Password", "")
	}
}

func serveSignUp(w http.ResponseWriter, r *http.Request) error {
	if err := checkSignupEnabled(); err != nil {
		return err
	}

	ctx := httpctx.FromRequest(r)
	if auth.IsAuthenticated(ctx) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return nil
	}

	switch r.Method {
	case "GET":
		return serveSignupForm(w, r, signupForm{})
	case "POST":
		return serveSignupSubmit(w, r)
	}
	http.Error(w, "", http.StatusMethodNotAllowed)
	return nil
}

func serveSignupForm(w http.ResponseWriter, r *http.Request, form signupForm) error {
	if err := checkSignupEnabled(); err != nil {
		return err
	}

	return tmpl.Exec(r, w, "user/signup.html", http.StatusOK, nil, &struct {
		SignupForm signupForm
		tmpl.Common
	}{
		SignupForm: form,
	})
}

func serveSignupSubmit(w http.ResponseWriter, r *http.Request) error {
	if err := checkSignupEnabled(); err != nil {
		return err
	}

	cl := handlerutil.APIClient(r)
	ctx := httpctx.FromRequest(r)

	var form signupForm
	if err := r.ParseForm(); err != nil {
		return err
	}
	if err := schemautil.Decode(&form, r.PostForm); err != nil {
		return err
	}

	form.Validate()
	if form.HasErrors() {
		return serveSignupForm(w, r, form)
	}

	userSpec, err := cl.Accounts.Create(ctx, &form.NewAccount)
	if err != nil {
		switch errcode.GRPC(err) {
		case codes.InvalidArgument:
			form.AddFieldError("Login", formErrorInvalidUsername)
		case codes.AlreadyExists:
			form.AddFieldError("Login", formErrorUsernameAlreadyTaken)

		default:
			return err
		}

		// Re-render form.
		return serveSignupForm(w, r, form)
	}

	// Get the newly created user's API key to authenticate future requests.
	tok, err := cl.Auth.GetAccessToken(ctx, &sourcegraph.AccessTokenRequest{
		ResourceOwnerPassword: &sourcegraph.LoginCredentials{Login: form.Login, Password: form.Password},
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

	returnTo, err := returnto.ExactURLFromQuery(r)
	if err != nil {
		return err
	}
	if returnTo == "" {
		returnTo = router.Rel.URLToUser(userSpec.Login).String()
	}

	http.Redirect(w, r, returnTo, http.StatusSeeOther)
	return nil
}

const (
	formErrorUsernameAlreadyTaken = "This username is already taken. Try another."
)

func checkSignupEnabled() error {
	if !authutil.ActiveFlags.HasSignup() {
		return &handlerutil.HTTPErr{Status: http.StatusNotFound, Err: errors.New("signup not enabled")}
	}
	return nil
}
