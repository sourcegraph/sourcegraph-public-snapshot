package e2etest

import (
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"

	"sourcegraph.com/sourcegraph/go-selenium"
)

func init() {
	Register(&Test{
		Name:        "register_flow",
		Description: "Registers a brand new user account via the join page.",
		Func:        TestRegisterFlow,
	})
}

func TestRegisterFlow(t *TestSuite) error {
	wd := t.WebDriverT()
	defer wd.Quit()

	// Create gRPC client connection so we can talk to the server. e2etest uses
	// the server's ID key for authentication, which means it can do ANYTHING with
	// no restrictions. Be careful!
	ctx, c := t.GRPCClient()

	// Delete the test user account.
	_, err := c.Accounts.Delete(ctx, &sourcegraph.PersonSpec{
		Login: "test123",
	})
	if err != nil && grpc.Code(err) != codes.NotFound {
		return err
	}

	// Get join page.
	wd.Get(t.Endpoint("/join"))

	// Validate username input field.
	username := wd.FindElement(selenium.ById, "login")
	if username.TagName() != "input" {
		t.Fatalf("username TagName should be input, found", username.TagName())
	}
	if username.Text() != "" {
		t.Fatalf("username input field should be empty, found", username.Text())
	}
	if !username.IsDisplayed() {
		t.Fatalf("username input field should be displayed")
	}
	if !username.IsEnabled() {
		t.Fatalf("username input field should be enabled")
	}
	// TODO(slimsag): Why don't we select this by default?!
	//if !username.IsSelected() {
	//	t.Fatalf("username input field should be selected")
	//}

	// Validate password input field.
	password := wd.FindElement(selenium.ById, "password")
	if password.TagName() != "input" {
		t.Fatalf("password TagName should be input, found", password.TagName())
	}
	if password.Text() != "" {
		t.Fatalf("password input field should be empty, found", password.Text())
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

	// Validate email input field.
	email := wd.FindElement(selenium.ById, "email")
	if email.TagName() != "input" {
		t.Fatalf("email TagName should be input, found", email.TagName())
	}
	if email.Text() != "" {
		t.Fatalf("email input field should be empty, found", email.Text())
	}
	if !email.IsDisplayed() {
		t.Fatalf("email input field should be displayed")
	}
	if !email.IsEnabled() {
		t.Fatalf("email input field should be enabled")
	}
	if email.IsSelected() {
		t.Fatalf("email input field should not be selected")
	}

	// Enter username and password for test account.
	username.SendKeys("test123")
	password.SendKeys("test123")
	email.SendKeys("test123@sourcegraph.com")

	// Click the submit button.
	//
	// TODO(slimsag): we should give a proper ID to this button.
	submit := wd.FindElement(selenium.ByCSSSelector, ".sign-up > button.btn")
	submit.Click()

	// Wait for redirect.
	timeout := time.After(1 * time.Second)
	for {
		if wd.CurrentURL() == t.Endpoint("/") {
			break
		}
		select {
		case <-timeout:
			t.Fatalf("expected redirect to homepage after register; CurrentURL=%q\n", wd.CurrentURL())
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
	return nil
}
