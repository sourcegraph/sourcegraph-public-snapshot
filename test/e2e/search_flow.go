package e2e

import "sourcegraph.com/sourcegraph/go-selenium"

func init() {
	Register(&Test{
		Name:        "search_flow",
		Description: "fetch every search item on sourcegraph.com for Go, ensure each first listing has usage examples",
		Func:        testSearchFlow,
	})
}

func runSearchFlow(t *T, query string) {
	wd := t.WebDriver

	err := wd.Get(t.Endpoint("/search"))
	if err != nil {
		t.Fatalf("TestSearchFlow: %s ", err)
	}

	searchInput := t.WaitForElement(selenium.ById, "e2etest-search-input")
	searchInput.SendKeys(query)

	// Since the search results are listed in `code` tags,
	// this will find the first search result (so it can be clicked)
	t.Click(selenium.ByTagName, "code")

	// The usage examples are in `table` elements
	t.WaitForElement(selenium.ByTagName, "table")
}

func testSearchFlow(t *T) error {
	wd := t.WebDriver

	err := wd.Get(t.Endpoint("/search"))
	if err != nil {
		t.Fatalf("TestSearchFlow: %s", err)
	}

	queries := [5]string{"new http request", "read file", "json encoder", "sql query", "indent json"}
	for _, q := range queries {
		runSearchFlow(t, q)
	}

	return nil
}
