package e2e

import "sourcegraph.com/sourcegraph/go-selenium"

func init() {
	Register(&Test{
		Name:        "search_flow",
		Description: "fetch gorilla/mux repository, search for RouteMatch and check the result link",
		Func:        testSearchFlow,
		Quarantined: true,
	})
}

func testSearchFlow(t *T) error {
	wd := t.WebDriver

	err := wd.Get(t.Endpoint("/github.com/gorilla/mux"))
	if err != nil {
		return err
	}

	searchInput := t.WaitForElement(selenium.ById, "search-input")
	searchInput.SendKeys("RouteMatch")

	t.WaitForElement(selenium.ByLinkText, "type RouteMatch struct", MatchAttribute("href", `/github\.com/gorilla/mux(@[^/]+)?/-/def/GoPackage/github.com/gorilla/mux/-/RouteMatch`))

	return nil
}
