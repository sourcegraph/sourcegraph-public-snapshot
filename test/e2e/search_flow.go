package e2e

import (
	"time"

	"sourcegraph.com/sourcegraph/go-selenium"
)

func init() {
	Register(&Test{
		Name:        "QuickOpenFiles",
		Description: "Test QuickOpen finds files.",
		Func:        fileSearch,
		Quarantined: true,
	})
	Register(&Test{
		Name:        "QuickOpenRepos",
		Description: "Test QuickOpen finds repositories.",
		Func:        fileSearch,
		Quarantined: true,
	})
	Register(&Test{
		Name:        "QuickOpenDefs",
		Description: "Test QuickOpen finds symbols.",
		Func:        fileSearch,
		Quarantined: true,
	})
}

func startSearch(query string, t *T) (*T, error) {
	wd := t.WebDriver

	err := wd.Get(t.Endpoint("/github.com/gorilla/mux"))
	if err != nil {
		return nil, err
	}

	time.Sleep(3 * time.Second)
	t.ExecuteScript(`document.getElementsByTagName('body')[0].dispatchEvent(new KeyboardEvent("keydown", {key: "/"}))`, nil)
	quickOpenInput := t.WaitForElement(selenium.ById, "SearchInput-e2e-test")
	quickOpenInput.SendKeys(query)
	return t, nil
}

func fileSearch(t *T) error {
	t, err := startSearch("mux", t)
	if err != nil {
		return err
	}
	e := t.WaitForElement(selenium.ByXPATH, `//div[text()="Files"]`)
	e.FindElement(selenium.ByXPATH, `//div[text()="mux.go"]`)
	return nil
}

func defSearch(t *T) error {
	t, err := startSearch("newroute", t)
	if err != nil {
		return err
	}
	e := t.WaitForElement(selenium.ByXPATH, `//div[text()="Definitions"]`)
	e.FindElement(selenium.ByXPATH, `//div[text()="mux.NewRouter"]`)
	return nil
}

func repoSearch(t *T) error {
	t, err := startSearch("github.com/gorilla/mux", t)
	if err != nil {
		return err
	}
	e := t.WaitForElement(selenium.ByXPATH, `//div[text()="Repositories"]`)
	e.FindElement(selenium.ByXPATH, `//div[text()="github.com/gorilla/mux"]`)
	return nil
}
