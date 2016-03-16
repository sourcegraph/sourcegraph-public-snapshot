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

	var muxLink selenium.WebElement
	getMuxLink := func() bool {
		var err error
		muxLink, err = wd.FindElement(selenium.ByPartialLinkText, "mux.go")
		if err != nil {
			return false
		}
		text, err := muxLink.Text()
		if err != nil {
			return false
		}

		return strings.Contains(text, "mux.go")
	}

	t.WaitForCondition(
		5*time.Second,
		100*time.Millisecond,
		getMuxLink,
		"Wait for mux.go codefile link to appear",
	)

	want := "/github.com/gorilla/mux@master/.tree/mux.go"

	have, err := muxLink.GetAttribute("href")
	if err != nil {
		return err
	}

	if !strings.Contains(have, want) {
		return fmt.Errorf("wanted: %s, got %s", want, have)
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

	// Wait for redirect.

	t.WaitForCondition(
		20*time.Second,
		100*time.Millisecond,
		wantURL(t.Endpoint("/github.com/gorilla/mux@master/.tree/mux.go"), wd),
		"wait for mux.go codefile to load",
	)

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
		5*time.Second,
		4*time.Second,
		getSpans,
		"Wait for Router span to appear",
	)
	// TODO(poler) test the hover-over

	routerSpan.Click()

	t.WaitForCondition(
		20*time.Second,
		100*time.Millisecond,
		wantURL(t.Endpoint("/github.com/gorilla/mux@master/.GoPackage/github.com/gorilla/mux/.def/Router"), wd),
		"wait for Router def to load",
	)

	return nil

}

func wantURL(wantedURL string, wd selenium.WebDriver) func() bool {
	return func() bool {
		currentURL, err := wd.CurrentURL()
		if err != nil {
			return false
		}

		return currentURL == wantedURL
	}
}
