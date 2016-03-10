package e2etest

import (
	"time"

	"sourcegraph.com/sourcegraph/go-selenium"
)

func init() {
	Register(&Test{
		Name:        "login_flow",
		Description: "Logs in to an existing user account via the login page.",
		Func:        TestLoginFlow,
	})
}

func TestLoginFlow(t *TestSuite) error {
	wd := t.WebDriverT()
	defer wd.Quit()

	// Get login page.
	wd.Get(t.Endpoint("/login"))

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

	// Enter username and password for test account.
	username.SendKeys("test123")
	password.SendKeys("test123")

	// Click the submit button.
	//
	// TODO(slimsag): we should give a proper ID to this button.
	submit := wd.FindElement(selenium.ByCSSSelector, ".log-in > button.btn")
	submit.Click()

	// Wait for redirect.
	timeout := time.After(1 * time.Second)
	for {
		if wd.CurrentURL() == t.Endpoint("/") {
			break
		}
		select {
		case <-timeout:
			t.Fatalf("expected redirect to homepage after sign-in; CurrentURL=%q\n", wd.CurrentURL())
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
	return nil
}
