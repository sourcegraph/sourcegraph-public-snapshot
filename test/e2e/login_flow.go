package e2e

import (
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"sourcegraph.com/sourcegraph/go-selenium"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
)

func init() {
	Register(&Test{
		Name:        "login_flow",
		Description: "Logs in to an existing user account via the login page.",
		Func:        testLoginFlow,
	})
}

func testLoginFlow(t *T) error {
	// Create gRPC client connection so we can talk to the server. e2etest uses
	// the server's ID key for authentication, which means it can do ANYTHING with
	// no restrictions. Be careful!
	ctx, c := t.GRPCClient()

	// Create the test user account.
	testPassword := "e2etest"
	_, err := c.Accounts.Create(ctx, &sourcegraph.NewAccount{
		Login:    t.TestLogin,
		Email:    t.TestEmail,
		Password: testPassword,
	})
	if err != nil && grpc.Code(err) != codes.AlreadyExists {
		return err
	}

	// Get login page.
	t.Get(t.Endpoint("/login"))

	// Give the JS time to set element focus, etc.
	time.Sleep(1 * time.Second)

	// Validate username input field.
	username := t.FindElement(selenium.ById, "e2etest-login-field")
	if username.TagName() != "input" {
		t.Fatalf("username TagName should be input, found %s", username.TagName())
	}
	if username.Text() != "" {
		t.Fatalf("username input field should be empty, found %s", username.Text())
	}
	if !username.IsDisplayed() {
		t.Fatalf("username input field should be displayed")
	}
	if !username.IsEnabled() {
		t.Fatalf("username input field should be enabled")
	}

	// Validate password input field.
	password := t.FindElement(selenium.ById, "e2etest-password-field")
	if password.TagName() != "input" {
		t.Fatalf("password TagName should be input, found %s", password.TagName())
	}
	if password.Text() != "" {
		t.Fatalf("password input field should be empty, found %s", password.Text())
	}
	if !password.IsDisplayed() {
		t.Fatalf("password input field should be displayed")
	}
	if !password.IsEnabled() {
		t.Fatalf("password input field should be enabled")
	}
	if password.IsSelected() {
		t.Fatalf("password input field should not be selected")
	}

	// Enter username and password for test account.
	username.Click()
	username.SendKeys(t.TestLogin)
	password.Click()
	password.SendKeys(testPassword)

	// Click the submit button.
	t.Click(selenium.ById, "e2etest-login-button")

	t.WaitForRedirect(t.Endpoint("/dashboard"), "wait for redirect to dashboard after sign-in")
	return nil
}
