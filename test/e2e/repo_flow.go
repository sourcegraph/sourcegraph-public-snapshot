package e2e

import (
	"errors"
	"time"

	"sourcegraph.com/sourcegraph/go-selenium"
)

func init() {
	Register(&Test{
		Name:        "repo_flow",
		Description: "fetch gorilla/mux repository, visit mux.go, and look at hover-over, def pop-up, and jump-to-def",
		Func:        testRepoFlow,
	})
}

func testRepoFlow(t *T) error {
	wd := t.WebDriver

	err := loginUser(t)
	if err != nil {
		return err
	}

	err = wd.Get(t.Endpoint("/github.com/gorilla/mux"))
	if err != nil {
		return err
	}

	// Check that the "mux.go" codefile link appears.
	muxLink := t.WaitForElement(selenium.ByLinkText, "mux.go", MatchAttribute("href", `/github\.com/gorilla/mux/-/blob/mux.go`))

	isDisplayed, err := muxLink.IsDisplayed()
	if err != nil {
		return err
	}

	if !isDisplayed {
		return errors.New("mux link should be displayed")
	}

	isEnabled, err := muxLink.IsEnabled()
	if err != nil {
		return err
	}

	if !isEnabled {
		return errors.New("mux link should be enabled")
	}

	muxLink.Click()

	t.WaitForRedirect(t.Endpoint("/github.com/gorilla/mux/-/blob/mux.go"), "wait for mux.go code file to load")

	// Wait for the "Router" ref link to appear.
	routerLink := t.WaitForElement(selenium.ByLinkText, "Router")
	// TODO(poler) test the hover-over

	// Perform a 2s sleep because the ref needs time to be linkified.
	time.Sleep(2 * time.Second)
	routerLink.MoveTo(0, 0) // Hover over element.
	routerLink.Click()      // Click the element.

	t.WaitForRedirectSuffix(
		"/-/def/GoPackage/github.com/gorilla/mux/-/Router",
		"wait for Router def to load",
	)
	return nil

}
