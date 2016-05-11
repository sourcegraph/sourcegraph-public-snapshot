package e2etest

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
	var muxLink selenium.WebElement
	t.WaitForCondition(
		20*time.Second,
		100*time.Millisecond,
		func() bool {
			var err error
			muxLink, err = wd.FindElement(selenium.ByLinkText, "mux.go")
			return err == nil
		},
		"Wait for mux.go codefile link to appear",
	)

	// If the link is displayed and enabled, click it.
	want := "/github.com/gorilla/mux@master/-/blob/mux.go"

	got, err := muxLink.GetAttribute("href")
	if err != nil {
		return err
	}

	if !strings.Contains(got, want) {
		return fmt.Errorf("got %s, want %s", got, want)
	}

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

	t.WaitForRedirect(t.Endpoint(want), "wait for mux.go code file to load")

	// Wait for the "Router" ref link to appear.
	var routerLink selenium.WebElement
	t.WaitForCondition(
		20*time.Second,
		4*time.Second,
		func() bool {
			var err error
			routerLink, err = wd.FindElement(selenium.ByLinkText, "Router")
			return err == nil
		},
		"Wait for Router link to appear",
	)
	// TODO(poler) test the hover-over

	// Perform a 2s sleep because the ref needs time to be linkified.
	time.Sleep(2 * time.Second)
	routerLink.MoveTo(0, 0) // Hover over element.
	routerLink.Click()      // Click the element.

	t.WaitForRedirect(
		t.Endpoint("/github.com/gorilla/mux@master/-/def/GoPackage/github.com/gorilla/mux/-/Router"),
		"wait for Router def to load",
	)
	return nil

}
