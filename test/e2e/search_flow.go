package e2e

import (
	"time"

	"sourcegraph.com/sourcegraph/go-selenium"
)

func init() {
	if true {
		// Disabled https://github.com/sourcegraph/sourcegraph/issues/1292
		return
	}

	registerTest := func(name, q string) {
		Register(&Test{
			Name:        name,
			Description: "fetch every search item on sourcegraph.com for Go, ensure each first listing has usage examples",
			Func: func(t *T) error {
				return runSearchFlow(t, q)
			},
			Quarantined: true,
		})
	}

	registerTest("search_flow_0", "new http request")
	registerTest("search_flow_1", "read file")
	registerTest("search_flow_2", "json encoder")
	registerTest("search_flow_3", "sql query")
	registerTest("search_flow_4", "indent json")
}

func runSearchFlow(t *T, query string) error {
	wd := t.WebDriver

	err := wd.Get(t.Endpoint("/search"))
	if err != nil {
		return err
	}

	searchInput := t.WaitForElement(selenium.ById, "e2etest-search-input")
	searchInput.SendKeys(query)
	time.Sleep(3 * time.Second)

	// Since the search results are listed in `code` tags,
	// this will find the first search result (so it can be clicked)
	t.Click(selenium.ByTagName, "code")

	// The usage examples are in these elements.
	t.WaitForElement(selenium.ByCSSSelector, "pre.snippet")
	return nil
}
