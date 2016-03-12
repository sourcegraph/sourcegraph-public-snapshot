package e2etest

import (
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
	t.Get(t.Endpoint("/github.com/gorilla/mux"))

	var muxLink selenium.WebElementT
	getMuxLink := func() bool {
		muxLink = t.FindElement(selenium.ByPartialLinkText, "mux.go")
		return strings.Contains(muxLink.Text(), "mux.go")
	}

	waitForCondition(
		t,
		5*time.Second,
		100*time.Millisecond,
		getMuxLink,
		"Wait for mux.go codefile link to appear",
	)

	want := "/github.com/gorilla/mux@master/.tree/mux.go"
	if have := muxLink.GetAttribute("href"); !strings.Contains(have, want) {
		t.Fatalf("wanted: %s, got %s", want, have)
	}

	if !muxLink.IsDisplayed() {
		t.Fatalf("mux link should be displayed")
	}

	if !muxLink.IsEnabled() {
		t.Fatalf("mux link should be enabled")
	}

	muxLink.Click()

	// Wait for redirect.

	waitForCondition(
		t,
		20*time.Second,
		100*time.Millisecond,
		func() bool { return t.CurrentURL() == t.Endpoint("/github.com/gorilla/mux@master/.tree/mux.go") },
		"wait for mux.go codefile to load",
	)

	var routerSpan selenium.WebElementT

	getSpans := func() bool {
		spans := t.FindElements(selenium.ByTagName, "span")

		for _, span := range spans {
			fmt.Println("BEFORE?????")
			text := span.Text()
			fmt.Println("AFTER?????")
			fmt.Println("text", text)
			if text == "Router" {
				routerSpan = span
				return true
			}
		}

		return false
	}

	waitForCondition(
		t,
		5*time.Second,
		100*time.Millisecond,
		getSpans,
		"Wait for Router span to appear",
	)
	// TODO(poler) test the hover-over

	routerSpan.Click()

	waitForCondition(
		t,
		20*time.Second,
		100*time.Millisecond,
		func() bool {
			return t.CurrentURL() == t.Endpoint("/github.com/gorilla/mux@master/.GoPackage/github.com/gorilla/mux/.def/Router")
		},
		"wait for Router def to load",
	)

	return nil

}
