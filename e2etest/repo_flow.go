package e2etest

import (
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

func TestRepoFlow(t *TestSuite) error {
	wd := t.WebDriverT()
	defer wd.Quit()

	wd.Get(t.Endpoint("/github.com/gorilla/mux"))

	muxLink := wd.FindElement(selenium.ByPartialLinkText, "mux.go")
	if muxLink.Text() == "" {
		t.Fatalf("mux link text is empty, should be mux.go")
	}

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
	timeout := time.After(20 * time.Second)
	for {
		if wd.CurrentURL() == t.Endpoint("/github.com/gorilla/mux@master/.tree/mux.go") {
			break
		}
		select {
		case <-timeout:
			t.Fatalf("expected redirect to homepage after sign-in; CurrentURL=%q\n", wd.CurrentURL())
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}

	time.Sleep(2 * time.Second)

	spans := wd.FindElements(selenium.ByTagName, "span")

	found := false
	var routerSpan selenium.WebElementT
	for _, span := range spans {

		if span.Text() == "Router" {
			routerSpan = span
			found = true
			break
		}
	}

	if !found {
		t.Fatalf("mux.go codefile does not contain a Router ref")
	}

	// TODO(poler) test the hover-over

	routerSpan.Click()

	timeout = time.After(20 * time.Second)
	for {
		if wd.CurrentURL() == t.Endpoint("/github.com/gorilla/mux@master/.GoPackage/github.com/gorilla/mux/.def/Router") {
			break
		}
		select {
		case <-timeout:
			t.Fatalf("Expected to be taken to Router def, currentURL=%q", wd.CurrentURL())
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}

	return nil

}
