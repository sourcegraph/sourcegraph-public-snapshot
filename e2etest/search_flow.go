package e2etest

import (
	"errors"
	"regexp"

	"sourcegraph.com/sourcegraph/go-selenium"
)

func init() {
	Register(&Test{
		Name:        "search_flow",
		Description: "fetch gorilla/mux repository, search for RouteMatch and check the result link",
		Func:        TestSearchFlow,
		Quarantined: true,
	})
}

func TestSearchFlow(t *T) error {
	wd := t.WebDriver

	err := wd.Get(t.Endpoint("/github.com/gorilla/mux"))
	if err != nil {
		return err
	}

	searchInput := t.WaitForElement(selenium.ById, "search-input")
	searchInput.SendKeys("RouteMatch")

	resultLink := t.WaitForElement(selenium.ByLinkText, "type RouteMatch struct")
	href, err := resultLink.GetAttribute("href")
	if err != nil {
		return err
	}
	if matched, _ := regexp.MatchString("/github.com/gorilla/mux@[^/]+/-/def/GoPackage/github.com/gorilla/mux/-/RouteMatch", href); !matched {
		return errors.New("unexpected def href: " + href)
	}

	return nil
}
