package backend

import (
	"errors"
	"net/url"
	"reflect"
	"testing"

	"github.com/mattbaird/gochimp"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	authpkg "sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/services/notif"

	"golang.org/x/net/context"
)

func TestCreate(t *testing.T) {
	ctx, mock := testContext()
	mock.stores.Password.SetPassword_ = func(ctx context.Context, uid int32, password string) error {
		if want := "secret"; password != want {
			t.Errorf("got %s, want %s", password, want)
		}
		if want := int32(123); uid != want {
			t.Errorf("got %d, want %d", uid, want)
		}
		return nil
	}
	mock.stores.Accounts.Create_ = func(ctx context.Context, u *sourcegraph.User, email *sourcegraph.EmailAddr) (*sourcegraph.User, error) {
		if want := "a user"; want != u.Login {
			t.Errorf("got %s, want %s", u.Login, want)
		}
		if u.Write || u.Admin {
			t.Errorf("got privileged account (write:%v, admin:%v), want regular account", u.Write, u.Admin)
		}
		return &sourcegraph.User{UID: 123}, nil
	}
	Accounts.Create(ctx, &sourcegraph.NewAccount{Login: "a user", Password: "func", Email: "email@email.com"})
}

func TestRequestPasswordReset(t *testing.T) {
	notif.MustBeDisabled()
	ctx, mock := testContext()
	ctx = conf.WithURL(ctx, &url.URL{})
	var sendEmailCalled bool
	sendEmail = func(template string, name string, email string, subject string, templateContent []gochimp.Var, mergeVars []gochimp.Var) ([]gochimp.SendResponse, error) {
		sendEmailCalled = true
		if want := "user@example.com"; want != email {
			t.Errorf("email address was %s, wanted %s", email, want)
		}
		return nil, nil
	}

	// This mocks the accesscontrol.VerifyHasAdminAccess function because that function
	// depends on commandline flags that are not mocked in this test.
	// The VerifyHasAdminAccess function is tested separately in the accesscontrol package
	// so the goal here is only to ensure that RequestPasswordReset behaves as expected
	// when it learns about a user's admin privilege.
	var verifyAdminCalled bool
	verifyAdminUser = func(ctx context.Context, method string) error {
		verifyAdminCalled = true
		a := authpkg.ActorFromContext(ctx)
		if a.HasAdminAccess() {
			return nil
		}
		return errors.New("no admin access")
	}

	mock.stores.Accounts.RequestPasswordReset_ = func(ctx context.Context, us *sourcegraph.User) (*sourcegraph.PasswordResetToken, error) {
		return &sourcegraph.PasswordResetToken{Token: "secrettoken"}, nil
	}
	mock.stores.Users.GetWithEmail_ = func(ctx context.Context, emailAddr sourcegraph.EmailAddr) (*sourcegraph.User, error) {
		return &sourcegraph.User{Name: "some user", Login: "user1"}, nil
	}
	mock.stores.Users.Get_ = func(ctx context.Context, userSpec sourcegraph.UserSpec) (*sourcegraph.User, error) {
		return &sourcegraph.User{Name: "some user", Login: "user1"}, nil
	}
	mock.stores.Users.ListEmails_ = func(ctx context.Context, userSpec sourcegraph.UserSpec) ([]*sourcegraph.EmailAddr, error) {
		return []*sourcegraph.EmailAddr{&sourcegraph.EmailAddr{Email: "user@example.com", Primary: true}}, nil
	}

	s := accounts{}
	p, err := s.RequestPasswordReset(ctx, &sourcegraph.PersonSpec{Email: "user@example.com"})
	if err != nil {
		t.Fatal(err)
	}
	if !sendEmailCalled {
		t.Errorf("sendEmail wasn't called")
	}
	if !verifyAdminCalled {
		t.Errorf("verifyAdminCalled wasn't called")
	}
	if p.Link != "" || p.Token.Token != "" || p.Login != "" {
		t.Errorf("expected no sensitive information in response, got %v", p)
	}

	// Request using login
	sendEmailCalled = false
	verifyAdminCalled = false
	p, err = s.RequestPasswordReset(ctx, &sourcegraph.PersonSpec{Login: "user1"})
	if err != nil {
		t.Fatal(err)
	}
	if !sendEmailCalled {
		t.Errorf("sendEmail wasn't called")
	}
	if !verifyAdminCalled {
		t.Errorf("verifyAdminCalled wasn't called")
	}
	if p.Link != "" || p.Token.Token != "" || p.Login != "" {
		t.Errorf("expected no sensitive information in response, got %v", p)
	}

	// Request as admin, expect reset link in response
	ctx = authpkg.WithActor(ctx, authpkg.Actor{UID: 2, Admin: true})
	sendEmailCalled = false
	verifyAdminCalled = false
	p, err = s.RequestPasswordReset(ctx, &sourcegraph.PersonSpec{Email: "user@example.com"})
	if err != nil {
		t.Fatal(err)
	}
	if !sendEmailCalled {
		t.Errorf("sendEmail wasn't called")
	}
	if !verifyAdminCalled {
		t.Errorf("verifyAdminCalled wasn't called")
	}
	want := &sourcegraph.PendingPasswordReset{
		Link:      "/reset?token=secrettoken",
		Token:     &sourcegraph.PasswordResetToken{Token: "secrettoken"},
		Login:     "user1",
		EmailSent: true,
	}

	if !reflect.DeepEqual(want, p) {
		t.Errorf("got %v, want %v", p, want)
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

	dbUser := &sourcegraph.User{
		UID:         123,
		Name:        "someName",
		HomepageURL: "someHomepageURL",
		Company:     "someCompany",
		Location:    "someLocation",
	}

	ctx, mock := testContext()
	mock.servers.Users.Get_ = func(ctx context.Context, in *sourcegraph.UserSpec) (*sourcegraph.User, error) {
		if in.UID == dbUser.UID {
			return dbUser, nil
		} else {
			// not important for the tests here.
			return &sourcegraph.User{UID: in.UID}, nil
		}
	}
	mock.stores.Accounts.Update_ = func(ctx context.Context, in *sourcegraph.User) error {
		if in.UID != dbUser.UID || in.Name != dbUser.Name || in.HomepageURL != dbUser.HomepageURL || in.Company != dbUser.Company || in.Location != dbUser.Location {
			t.Errorf("got %v, want %v", in, user)
		}
		return nil
	}
	if _, err := Accounts.Update(ctx, user); err != nil {
		t.Error(err)
	}
}
