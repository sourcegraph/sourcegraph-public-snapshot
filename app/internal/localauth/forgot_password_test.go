package localauth

import (
	"net/http"
	"net/url"
	"testing"

	"golang.org/x/net/context"

	"github.com/google/go-querystring/query"

	"sourcegraph.com/sqs/pbtypes"
	"src.sourcegraph.com/sourcegraph/app/internal/apptest"
	"src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

func TestGetForgotPassword(t *testing.T) {
	c, _ := apptest.New()

	resp, err := c.GetOK(router.Rel.URLTo(router.ForgotPassword).String())
	if err != nil {
		t.Errorf("wanted nil, got %s", err)
		t.Fatal(err)
	}
	if want := http.StatusOK; resp.StatusCode != want {
		t.Errorf("wanted %d, got %d", want, resp.StatusCode)
	}
}

func TestPostForgotPassword(t *testing.T) {
	c, mock := apptest.New()

	data, err := query.Values(userForm{Email: "who@me.com"})
	if err != nil {
		t.Fatal(err)
	}

	var called bool
	mock.Accounts.RequestPasswordReset_ = func(ctx context.Context, email *sourcegraph.EmailAddr) (*sourcegraph.User, error) {
		called = true
		if want := "who@me.com"; email.Email != want {
			t.Errorf("wanted %s, got %s", want, email.Email)
		}
		return nil, nil
	}

	resp, err := c.PostFormNoFollowRedirects(router.Rel.URLTo(router.ForgotPassword).String(), data)
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Errorf("mock function was not called")
	}
	if want := http.StatusOK; resp.StatusCode != want {
		t.Errorf("wanted %d, got %d", want, resp.StatusCode)
	}
}

func TestServeNewPassword(t *testing.T) {
	c, _ := apptest.New()

	u := router.Rel.URLTo(router.ResetPassword)
	p := url.Values{}
	p.Add("token", "supersecrettoken")
	u.RawQuery = p.Encode()
	resp, err := c.GetOK(u.String())
	if err != nil {
		t.Fatal(err)
	}
	if want := http.StatusOK; resp.StatusCode != want {
		t.Errorf("wanted %d, got %d", want, resp.StatusCode)
	}
}

func TestServeResetPassword(t *testing.T) {
	c, mock := apptest.New()

	var called bool
	mock.Accounts.ResetPassword_ = func(ctx context.Context, new *sourcegraph.NewPassword) (*pbtypes.Void, error) {
		called = true
		if want := "hunter2"; new.Password != want {
			t.Errorf("wanted password %s, got %s", want, new.Password)
		}

		if want := "supersecrettoken"; new.Token.Token != want {
			t.Errorf("wanted token %s, got %s", want, new.Token.Token)
		}
		return nil, nil
	}

	u := router.Rel.URLTo(router.ResetPassword)
	v := url.Values{}
	v.Add("token", "supersecrettoken")
	u.RawQuery = v.Encode()

	data, err := query.Values(passwordForm{Password: "hunter2", ConfirmPassword: "hunter2"})
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.PostFormNoFollowRedirects(u.String(), data)
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Errorf("mock function was not called")
	}
	if resp.StatusCode != http.StatusSeeOther {
		t.Errorf("got %d, wanted %d", resp.StatusCode, http.StatusSeeOther)
	}
	lv := resp.Header.Get("Location")
	if "/login" != lv {
		t.Errorf("redirected to %s, wanted %s", lv, "/login")
	}
}
