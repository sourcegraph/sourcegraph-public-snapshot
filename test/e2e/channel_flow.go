package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"

	"sourcegraph.com/sourcegraph/go-selenium"
)

func init() {
	Register(&Test{
		Name:        "channel_flow",
		Description: "Registers a brand new user account via the join page.",
		Func:        testChannelFlow,
	})
}

func testChannelFlow(t *T) error {
	wd := t.WebDriver

	err := wd.Get(t.Endpoint("/-/channel/e2etest-asdfasdfasdfasdfasdfasdfasdfasdf"))
	if err != nil {
		return err
	}

	// establish channel initialization page
	t.WaitForElement(selenium.ByXPATH, "//*[contains(text(), 'Click on a symbol in your editor to get started!')]")
	// check that the "connected" status appears
	t.WaitForElement(selenium.ByXPATH, "//*[contains(text(), 'connected')]")

	type Action struct {
		Repo    string `json:"Repo,omitempty"`
		Package string `json:"Package,omitempty"`
		Def     string `json:"Def,omitempty"`
		Error   string `json:"Error,omitempty"`
	}

	type Request struct {
		Action            Action `json:"Action,omitempty"`
		CheckForListeners bool   `json:"CheckForListeners,omitempty"`
	}

	// Test that the page changes to the gorilla/mux repo tree view after POST request
	u := &Request{Action: Action{
		Repo:    "github.com/gorilla/mux",
		Package: "github.com/gorilla/mux",
	}, CheckForListeners: true}
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(u)

	_, err = http.Post("https://grpc.sourcegraph.com/.api/channel/e2etest-asdfasdfasdfasdfasdfasdfasdfasdf", "application/json; charset=utf-8", body)
	if err != nil {
		return err
	}

	t.WaitForElement(selenium.ByXPATH, "//*[contains(text(), 'connected')]")
	// Check that the "mux.go" codefile link appears.
	t.WaitForElement(selenium.ByLinkText, "mux.go", MatchAttribute("href", `/github\.com/gorilla/mux/-/blob/mux.go`))

	// Test that the page changes to the definfo page of http.Post after POST request
	u = &Request{Action: Action{
		Repo:    "net/http",
		Package: "net/http",
		Def:     "Post",
	}, CheckForListeners: true}
	body = new(bytes.Buffer)
	json.NewEncoder(body).Encode(u)

	_, err = http.Post("https://grpc.sourcegraph.com/.api/channel/e2etest-asdfasdfasdfasdfasdfasdfasdfasdf", "application/json; charset=utf-8", body)
	if err != nil {
		return err
	}

	t.WaitForElement(selenium.ByXPATH, "//*[contains(text(), 'connected')]")
	// check that the definfo page has loaded
	t.WaitForElement(selenium.ByLinkText, "View")
	t.WaitForElement(selenium.ByXPATH, "//*[contains(text(), 'Post issues a POST to the specified URL.')]")
	t.WaitForElement(selenium.ByXPATH, "//*[contains(text(), 'bradfitz')]")
	t.WaitForElement(selenium.ByXPATH, "//*[contains(text(), 'Used in')]")

	return nil
}
