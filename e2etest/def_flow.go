package e2etest

import (
	"errors"
	"regexp"

	"sourcegraph.com/sourcegraph/go-selenium"
)

func init() {
	Register(&Test{
		Name:        "def_flow",
		Description: "start at info page for net/http/Header.Get (gddo entrypoint), click on def, go back to info",
		Func:        testDefFlow,
	})
}

func testDefFlow(t *T) error {
	wd := t.WebDriver

	err := wd.Get(t.Endpoint("/github.com/golang/go/-/def/GoPackage/net/http/-/Header/Get/-/info"))
	if err != nil {
		return err
	}

	t.WaitForElement(selenium.ByLinkText, "View")
	t.WaitForElement(selenium.ByXPATH, "//*[contains(text(), 'Get gets the first value associated with the given key')]")
	t.WaitForElement(selenium.ByXPATH, "//*[contains(text(), 'petarm')]")
	t.WaitForElement(selenium.ByXPATH, "//*[contains(text(), 'Used in')]")
	// TODO(keegancsmith) Find a reliable way to tell if the code view has loaded

	// Check that the def link appears
	defLink := t.WaitForElement(selenium.ByLinkText, "(Header).Get(key string) string")
	href, err := defLink.GetAttribute("href")
	if err != nil {
		return err
	}
	if matched, _ := regexp.MatchString("/github.com/golang/go@[^/]+/-/def/GoPackage/net/http/-/Header/Get", href); !matched {
		return errors.New("unexpected def href: " + href)
	}

	err = defLink.Click()
	if err != nil {
		return err
	}
	defLink = t.WaitForElement(selenium.ByLinkText, "(Header).Get(key string) string")
	href, err = defLink.GetAttribute("href")
	if err != nil {
		return err
	}
	if matched, _ := regexp.MatchString("/github.com/golang/go@[^/]+/-/def/GoPackage/net/http/-/Header/Get/-/info", href); !matched {
		return errors.New("unexpected def info href: " + href)
	}

	err = defLink.Click()
	if err != nil {
		return err
	}
	return nil
}
