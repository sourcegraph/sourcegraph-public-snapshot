package e2e

import "sourcegraph.com/sourcegraph/go-selenium"

func init() {
	Register(&Test{
		Name:        "search_flow",
		Description: "fetch gorilla/mux repository, search for RouteMatch and check the result link",
		Func:        testSearchFlow,
	})
}

func testSearchFlow(t *T) error {
	wd := t.WebDriver

	err := wd.Get(t.Endpoint("/search"))
	if err != nil {
		return err
	}

	selectLang := t.FindElement(selenium.ById, "e2etest-search-lang-select-golang")
	selectLang.Click()
	searchInput := t.WaitForElement(selenium.ById, "e2etest-search-input")
	searchInput.SendKeys("RouteMatch")

	t.WaitForElement(selenium.ByTagName, "a", MatchAttribute("href", `/github\.com/gorilla/mux(@[^/]+)?/-/info/GoPackage/github.com/gorilla/mux/-/RouteMatch`))

	return nil
}
