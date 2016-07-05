package e2e

import "sourcegraph.com/sourcegraph/go-selenium"

func init() {
	Register(&Test{
		Name:        "def_flow",
		Description: "start at info page for net/http/Header.Get (gddo entrypoint), click on def, go back to info",
		Func:        testDefFlow,
	})
}

func testDefFlow(t *T) error {
	wd := t.WebDriver

	err := wd.Get(t.Endpoint("/github.com/golang/go/-/info/GoPackage/net/http/-/Header/Get"))
	if err != nil {
		return err
	}

	// Check that the def link appears
	defLink := t.WaitForElement(selenium.ByLinkText, "View definition", MatchAttribute("href", `/github\.com/golang/go/-/def/GoPackage/net/http/-/Header/Get`))
	if err = defLink.Click(); err != nil {
		return err
	}

	defLink = t.WaitForElement(selenium.ByLinkText, "View all references", MatchAttribute("href", `/github\.com/golang/go/-/info/GoPackage/net/http/-/Header/Get`))
	if err = defLink.Click(); err != nil {
		return err
	}

	return nil
}
