package e2e

import (
	"errors"
	"fmt"
	"strings"
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

	err := wd.Get(t.Endpoint("/github.com/gorilla/mux"))
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

	// Perform a sleep because the editor needs to load.
	time.Sleep(3 * time.Second)
	editor := t.FindElement(selenium.ByCSSSelector, "div[data-mode-id=\"go\"]")
	time.Sleep(2 * time.Second)
	if got, want := editor.Text(), "func NewRouter()"; !strings.Contains(got, want) {
		return fmt.Errorf("editor does not contain %q", want)
	}

	// TODO(monaco): add tests for go-to-def

	return nil
}
