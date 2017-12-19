package localstore

import (
	"testing"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

func TestUsers_NativeAuth(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := testContext()

	if _, err := Users.Create(ctx, "native:foo@bar.com", "foo@bar.com", "foo", "foo", sourcegraph.UserProviderNative, nil, "", ""); err == nil {
		t.Fatal("native user created without password")
	}
	if _, err := Users.Create(ctx, "native:foo@bar.com", "foo@bar.com", "foo", "foo", sourcegraph.UserProviderNative, nil, "asdfasdf", ""); err == nil {
		t.Fatal("native user created without email verification code")
	}
	if _, err := Users.Create(ctx, "foo@bar.com", "foo@bar.com", "foo", "foo", "", nil, "qwer", ""); err == nil {
		t.Fatal("non-native user created with password")
	}
	if _, err := Users.Create(ctx, "foo@bar.com", "foo@bar.com", "foo", "foo", "", nil, "", "qwer"); err == nil {
		t.Fatal("non-native user created with email verification code")
	}
	if _, err := Users.Create(ctx, "sso:foo@bar.com", "foo@bar.com", "foo", "foo", "", nil, "qwer", ""); err == nil {
		t.Fatal("sso user created with password")
	}
	if _, err := Users.Create(ctx, "sso:foo@bar.com", "foo@bar.com", "foo", "foo", "", nil, "", "qwer"); err == nil {
		t.Fatal("sso user created with email verification code")
	}

	usr, err := Users.Create(ctx, "native:foo@bar.com", "foo@bar.com", "foo", "foo", sourcegraph.UserProviderNative, nil, "right-password", "email-code")
	if err != nil {
		t.Fatal(err)
	}
	if usr.Verified {
		t.Fatal("new user should not be verified")
	}
	if isValid, err := Users.ValidateEmail(ctx, usr.ID, "wrong_email-code"); err == nil && isValid {
		t.Fatal("should validate email with wrong code")
	}
	if isValid, err := Users.ValidateEmail(ctx, usr.ID, "email-code"); err != nil || !isValid {
		t.Fatal("couldn't vaidate email")
	}
	usr, err = Users.GetByID(ctx, usr.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !usr.Verified {
		t.Fatal("user should not be verified")
	}
	if isPassword, err := Users.IsPassword(ctx, usr.ID, "right-password"); err != nil || !isPassword {
		t.Fatal("didn't accept correct password")
	}
	if isPassword, err := Users.IsPassword(ctx, usr.ID, "wrong-password"); err == nil && isPassword {
		t.Fatal("accepted wrong password")
	}
	if _, err := Users.RenewPasswordResetCode(ctx, 193092309); err == nil {
		t.Fatal("no error renewing password reset for non-existent users")
	}
	resetCode, err := Users.RenewPasswordResetCode(ctx, usr.ID)
	if err != nil {
		t.Fatal(err)
	}
	if success, err := Users.SetPassword(ctx, usr.ID, "wrong-code", "new-password"); err == nil && success {
		t.Fatal("password updated without right reset code")
	}
	if success, err := Users.SetPassword(ctx, usr.ID, "", "new-password"); err == nil && success {
		t.Fatal("password updated without reset code")
	}
	if isPassword, err := Users.IsPassword(ctx, usr.ID, "right-password"); err != nil || !isPassword {
		t.Fatal("password changed")
	}
	if success, err := Users.SetPassword(ctx, usr.ID, resetCode, "new-password"); err != nil || !success {
		t.Fatalf("failed to update user password with code: %s", err)
	}
	if isPassword, err := Users.IsPassword(ctx, usr.ID, "new-password"); err != nil || !isPassword {
		t.Fatalf("new password doesn't work: %s", err)
	}
	if isPassword, err := Users.IsPassword(ctx, usr.ID, "right-password"); err == nil && isPassword {
		t.Fatal("old password still works")
	}
}
