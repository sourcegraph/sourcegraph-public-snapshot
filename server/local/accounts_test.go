package local

import (
	"net/url"
	"os"
	"testing"

	"github.com/mattbaird/gochimp"

	authpkg "src.sourcegraph.com/sourcegraph/auth"
	"src.sourcegraph.com/sourcegraph/conf"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/notif"

	"golang.org/x/net/context"
)

func TestCreateFirstAccount(t *testing.T) {
	ctx, mock := testContext()
	mock.stores.Users.Count_ = func(ctx context.Context, _ *pbtypes.Void) (*sourcegraph.UserCount, error) {
		return &sourcegraph.UserCount{}, nil
	}
	mock.stores.Password.SetPassword_ = func(ctx context.Context, uid int32, password string) error {
		if want := "secret"; password != want {
			t.Errorf("got %s, want %s", password, want)
		}
		if want := int32(123); uid != want {
			t.Errorf("got %d, want %d", uid, want)
		}
		return nil
	}
	mock.stores.Accounts.Create_ = func(ctx context.Context, u *sourcegraph.User) (*sourcegraph.User, error) {
		if want := "a user"; want != u.Login {
			t.Errorf("got %s, want %s", u.Login, want)
		}
		if !u.Write || !u.Admin {
			t.Errorf("got non-privileged account (write:%v, admin:%v), want admin account", u.Write, u.Admin)
		}
		return &sourcegraph.User{UID: 123}, nil
	}
	Accounts.Create(ctx, &sourcegraph.NewAccount{Login: "a user", Password: "func"})
}

func TestCreate(t *testing.T) {
	ctx, mock := testContext()
	mock.stores.Users.Count_ = func(ctx context.Context, _ *pbtypes.Void) (*sourcegraph.UserCount, error) {
		return &sourcegraph.UserCount{Count: 1}, nil
	}
	mock.stores.Password.SetPassword_ = func(ctx context.Context, uid int32, password string) error {
		if want := "secret"; password != want {
			t.Errorf("got %s, want %s", password, want)
		}
		if want := int32(123); uid != want {
			t.Errorf("got %d, want %d", uid, want)
		}
		return nil
	}
	mock.stores.Accounts.Create_ = func(ctx context.Context, u *sourcegraph.User) (*sourcegraph.User, error) {
		if want := "a user"; want != u.Login {
			t.Errorf("got %s, want %s", u.Login, want)
		}
		if u.Write || u.Admin {
			t.Errorf("got privileged account (write:%v, admin:%v), want regular account", u.Write, u.Admin)
		}
		return &sourcegraph.User{UID: 123}, nil
	}
	Accounts.Create(ctx, &sourcegraph.NewAccount{Login: "a user", Password: "func"})
}

func TestRequestPasswordReset(t *testing.T) {
	notif.MustBeDisabled()
	ctx, mock := testContext()
	ctx = conf.WithAppURL(ctx, &url.URL{})
	var called bool
	sendEmail = func(template string, name string, email string, subject string, templateContent []gochimp.Var, mergeVars []gochimp.Var) ([]gochimp.SendResponse, error) {
		called = true
		if want := "user@example.com"; want != email {
			t.Errorf("email address was %s, wanted %s", email, want)
		}
		return nil, nil
	}

	mock.stores.Accounts.RequestPasswordReset_ = func(ctx context.Context, us *sourcegraph.User) (*sourcegraph.PasswordResetToken, error) {
		return &sourcegraph.PasswordResetToken{Token: "secrettoken"}, nil
	}
	mock.stores.Users.GetWithEmail_ = func(ctx context.Context, emailAddr sourcegraph.EmailAddr) (*sourcegraph.User, error) {
		return &sourcegraph.User{Name: "some user", Login: "user1"}, nil
	}

	s := accounts{}
	u, err := s.RequestPasswordReset(ctx, &sourcegraph.EmailAddr{Email: "user@example.com"})
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Errorf("sendEmail wasn't called")
	}
	if want := "user1"; u.Login != want {
		t.Errorf("Got user login %s, wanted %s.", u.Login, want)
	}
}

func TestResetPassword(t *testing.T) {
	var called bool
	ctx, mock := testContext()
	mock.stores.Accounts.ResetPassword_ = func(ctx context.Context, new *sourcegraph.NewPassword) error {
		called = true
		if new.Password != "hunter2" || new.Token.Token != "secrettoken" {
			t.Errorf("Didn't receive expected new password")
		}
		return nil
	}
	s := accounts{}
	_, err := s.ResetPassword(ctx, &sourcegraph.NewPassword{Password: "hunter2", Token: &sourcegraph.PasswordResetToken{Token: "secrettoken"}})
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Errorf("(store) reset password wasn't called")
	}
}

func TestUpdate(t *testing.T) {
	user := &sourcegraph.User{
		UID:         123,
		Name:        "someName",
		HomepageURL: "someHomepageURL",
		Company:     "someCompany",
		Location:    "someLocation",
	}

	ctx, mock := testContext()
	mock.stores.Accounts.Update_ = func(ctx context.Context, in *sourcegraph.User) error {
		if in.UID != user.UID || in.Name != user.Name || in.HomepageURL != user.HomepageURL || in.Company != user.Company || in.Location != user.Location {
			t.Errorf("got %v, want %v", in, user)
		}
		return nil
	}
	ctx = authpkg.WithActor(ctx, authpkg.Actor{UID: 123})

	if _, err := Accounts.Update(ctx, user); err != nil {
		t.Error(err)
	}

	if _, err := Accounts.Update(ctx, &sourcegraph.User{UID: 124}); err != os.ErrPermission {
		t.Errorf("expected os.ErrPermission, got %v", err)
	}
}
