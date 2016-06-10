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

	t.WaitForElement(selenium.ByLinkText, "View")
	t.WaitForElement(selenium.ByXPATH, "//*[contains(text(), 'Get gets the first value associated with the given key')]")
	t.WaitForElement(selenium.ByXPATH, "//*[contains(text(), 'petarm')]")
	t.WaitForElement(selenium.ByXPATH, "//*[contains(text(), 'Usage examples')]")
	// TODO(keegancsmith) Find a reliable way to tell if the code view has loaded

	// Check that the def link appears
	defLink := t.WaitForElement(selenium.ByLinkText, "(Header).Get(key string) string", MatchAttribute("href", `/github\.com/golang/go/-/def/GoPackage/net/http/-/Header/Get`))
	if err = defLink.Click(); err != nil {
		return err
	}

	defLink = t.WaitForElement(selenium.ByLinkText, "(Header).Get(key string) string", MatchAttribute("href", `/github\.com/golang/go/-/info/GoPackage/net/http/-/Header/Get`))
	if err = defLink.Click(); err != nil {
		return err
	}

	return nil
}
