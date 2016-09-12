package e2e

import "sourcegraph.com/sourcegraph/go-selenium"

func init() {
	Register(&Test{
		Name:        "login_flow",
		Description: "Logs in to an existing user account via the login page.",
		Func:        testLoginFlow,
	})
}

func testLoginFlow(t *T) error {
	// Get login page.
	t.Get(t.Endpoint("/"))
	t.Click(selenium.ByLinkText, "Login")

	// Validate username input field.
	t.WaitForElement(selenium.ById, "e2etest-login-field")
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
	password.SendKeys("e2etest")

	// Click the submit button.
	t.Click(selenium.ById, "e2etest-login-button")

	t.WaitForRedirect(t.Endpoint("/"), "wait for redirect to home after sign-in")
	return nil
}
