package localauth

import (
	"fmt"
	"net/http"
	"net/url"

	"google.golang.org/grpc/codes"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/app/internal"
	"src.sourcegraph.com/sourcegraph/app/internal/form"
	"src.sourcegraph.com/sourcegraph/app/internal/schemautil"
	"src.sourcegraph.com/sourcegraph/app/internal/tmpl"
	"src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/errcode"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
)

func init() {
	internal.Handlers[router.ForgotPassword] = serveForgotPassword
	internal.Handlers[router.ResetPassword] = serveResetPassword
}

type userForm struct {
	form.Validation

	Login string
}

func serveForgotPassword(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case "GET":
		return serveForgotPasswordForm(w, r, userForm{})
	case "POST":
		return serveForgotPasswordSubmit(w, r)
	}
	return nil
}

func serveForgotPasswordForm(w http.ResponseWriter, r *http.Request, form userForm) error {
	return tmpl.Exec(r, w, "user/forgot_password.html", http.StatusOK, nil, &struct {
		UserForm userForm
		tmpl.Common
	}{
		UserForm: form,
	})
}

func serveForgotPasswordSubmit(w http.ResponseWriter, r *http.Request) error {
	cl := handlerutil.APIClient(r)
	ctx := httpctx.FromRequest(r)

	var form userForm
	if err := r.ParseForm(); err != nil {
		return err
	}

	if err := schemautil.Decode(&form, r.PostForm); err != nil {
		return err
	}

	_, err := cl.Accounts.RequestPasswordReset(ctx, &sourcegraph.UserSpec{Login: form.Login})
	if err != nil {
		switch errcode.GRPC(err) {
		case codes.NotFound:
			form.AddFieldError("Login", formErrorNoUserExists)
			return serveForgotPasswordForm(w, r, form)
		default:
			return err
		}
	}

	// Notify the user that a password reset email was sent to their email address.
	return tmpl.Exec(r, w, "user/password_reset.html", http.StatusOK, nil, &struct {
		tmpl.Common
	}{})
}

type passwordForm struct {
	form.Validation

	Password        string
	ConfirmPassword string
}

// A user should arrive here after clicking the password reset link in their email.
func serveResetPassword(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case "GET":
		return serveNewPassword(w, r, passwordForm{})
	case "POST":
		return serveNewPasswordSubmit(w, r)
	}
	return nil
}

func serveNewPassword(w http.ResponseWriter, r *http.Request, form passwordForm) error {
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form error: %s", err)
	}
	token := r.Form.Get("token")
	u := router.Rel.URLTo(router.ResetPassword)
	v := url.Values{}
	v.Set("token", token)
	u.RawQuery = v.Encode()
	return tmpl.Exec(r, w, "user/new_password.html", http.StatusOK, nil, &struct {
		tmpl.Common
		PostURL      string
		PasswordForm passwordForm
	}{
		PostURL:      u.String(),
		PasswordForm: form,
	})
}

func serveNewPasswordSubmit(w http.ResponseWriter, r *http.Request) error {
	ctx := httpctx.FromRequest(r)
	cl := handlerutil.APIClient(r)
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form error: %s", err)
	}
	token := r.Form.Get("token")

	var form passwordForm
	if err := schemautil.Decode(&form, r.PostForm); err != nil {
		return fmt.Errorf("error decoding form: %s", err)
	}

	if form.ConfirmPassword != form.Password {
		form.AddFieldError("Password", "Your password must be the same as the confirmation")
		return serveNewPassword(w, r, form)
	}
	_, err := cl.Accounts.ResetPassword(ctx, &sourcegraph.NewPassword{Password: form.Password, Token: &sourcegraph.PasswordResetToken{Token: token}})
	if err != nil {
		return fmt.Errorf("error reseting password: %s", err)
	}
	http.Redirect(w, r, router.Rel.URLTo(router.LogIn).String(), http.StatusSeeOther)
	return nil
}
