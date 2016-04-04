package localauth

import (
	"fmt"
	"net/http"
	"net/url"

	"google.golang.org/grpc/codes"

	"sourcegraph.com/sourcegraph/sourcegraph/app/internal"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/form"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/schemautil"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/tmpl"
	"sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/services/notif"
	"sourcegraph.com/sourcegraph/sourcegraph/util/errcode"
	"sourcegraph.com/sourcegraph/sourcegraph/util/handlerutil"
)

func init() {
	internal.Handlers[router.ForgotPassword] = serveForgotPassword
	internal.Handlers[router.ResetPassword] = serveResetPassword
}

type userForm struct {
	form.Validation

	Email string
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
	userSpec := handlerutil.UserFromRequest(r)
	if userSpec != nil {
		http.Redirect(w, r, router.Rel.URLTo(router.Home).String(), http.StatusSeeOther)
		return nil
	}

	return tmpl.Exec(r, w, "user/forgot_password.html", http.StatusOK, nil, &struct {
		UserForm          userForm
		IsEmailConfigured bool
		tmpl.Common
	}{
		IsEmailConfigured: notif.EmailIsConfigured(),
		UserForm:          form,
	})
}

func serveForgotPasswordSubmit(w http.ResponseWriter, r *http.Request) error {
	ctx, cl := handlerutil.Client(r)

	var form userForm
	if err := r.ParseForm(); err != nil {
		return err
	}

	if err := schemautil.Decode(&form, r.PostForm); err != nil {
		return err
	}

	_, err := cl.Accounts.RequestPasswordReset(ctx, &sourcegraph.PersonSpec{Email: form.Email})
	if err != nil {
		switch errcode.GRPC(err) {
		case codes.NotFound, codes.InvalidArgument:
			form.AddFieldError("Email", formErrorNoEmailExists)
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
	ctx, cl := handlerutil.Client(r)
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
