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
		Func:        TestRepoFlow,
	})
}

func TestRepoFlow(t *T) error {
	wd := t.WebDriver

	err := wd.Get(t.Endpoint("/github.com/gorilla/mux"))
	if err != nil {
		return err
	}

	// Check that the "mux.go" codefile link appears.
	var muxLink selenium.WebElement
	getMuxLink := func() bool {
		spans, err := wd.FindElements(selenium.ByTagName, "span")
		if err != nil {
			return false
		}

		for _, span := range spans {
			text, err := span.Text()
			if err != nil {
				return false
			}
			if strings.Contains(text, "mux.go") {
				muxLink = span
				return true
			}
		}
		return false
	}

	t.WaitForCondition(
		20*time.Second,
		100*time.Millisecond,
		getMuxLink,
		"Wait for mux.go codefile link to appear",
	)

	// If the link is displayed and enabled, click it.
	want := "/github.com/gorilla/mux@master/-/blob/mux.go"

	got, err := muxLink.GetAttribute("href")
	if err != nil {
		return err
	}

	if !strings.Contains(got, want) {
		return fmt.Errorf("wanted: %s, got %s", want, got)
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

	t.WaitForRedirect(want, "wait for mux.go codefile to load")

	// Wait for the "Router" ref span to appear.
	var routerSpan selenium.WebElement
	getSpans := func() bool {
		spans, err := wd.FindElements(selenium.ByTagName, "span")
		if err != nil {
			return false
		}

		for _, span := range spans {
			text, err := span.Text()
			if err != nil {
				return false
			}
			if text == "Router" {
				routerSpan = span
				return true
			}
		}

		return false
	}

	t.WaitForCondition(
		20*time.Second,
		4*time.Second,
		getSpans,
		"Wait for Router span to appear",
	)
	// TODO(poler) test the hover-over

	// Perform a 2s sleep because the span needs time to be linkified.
	time.Sleep(2 * time.Second)
	routerSpan.Click()

	t.WaitForRedirect(
		"/github.com/gorilla/mux@master/-/def/GoPackage/github.com/gorilla/mux/-/Router",
		"wait for Router def to load",
	)
	return nil

}
