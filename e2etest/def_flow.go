package e2etest

import (
	"errors"
	"regexp"
	"time"

	"sourcegraph.com/sourcegraph/go-selenium"
)

func init() {
	Register(&Test{
		Name:        "def_flow",
		Description: "start at info page for net/http/Header.Get (gddo entrypoint), click on def, go back to info",
		Func:        TestDefFlow,
	})
}

func TestDefFlow(t *T) error {
	wd := t.WebDriver

	err := wd.Get(t.Endpoint("/github.com/golang/go/-/def/GoPackage/net/http/-/Header/Get/-/info"))
	if err != nil {
		return err
	}

	timeout := 20 * time.Second
	canFindElement := func(by, value string) func() bool {
		return func() bool {
			_, err := wd.FindElement(by, value)
			return err == nil
		}
	}

	t.WaitForCondition(
		timeout,
		100*time.Millisecond,
		canFindElement(selenium.ByLinkText, "View"),
		"Wait for View link(s) to appear",
	)
	t.WaitForCondition(
		timeout,
		100*time.Millisecond,
		canFindElement(selenium.ByXPATH, "//*[contains(text(), 'Get gets the first value associated with the given key')]"),
		"Wait for doc string to appear",
	)
	t.WaitForCondition(
		timeout,
		100*time.Millisecond,
		canFindElement(selenium.ByXPATH, "//*[contains(text(), 'petarm')]"),
		"Wait for author to appear",
	)
	t.WaitForCondition(
		timeout,
		100*time.Millisecond,
		canFindElement(selenium.ByXPATH, "//*[contains(text(), 'Used in')]"),
		"Wait for DefInfo tracked count",
	)
	// TODO(keegancsmith) Find a reliable way to tell if the code view has loaded

	// Check that the def link appears
	var defLink selenium.WebElement
	getDefLink := func() bool {
		links, err := wd.FindElements(selenium.ByTagName, "a")
		if err != nil {
			return false
		}

		for _, link := range links {
			text, err := link.Text()
			if err != nil {
				return false
			}
			if text == "(Header).Get(key string) string" {
				defLink = link
				return true
			}
		}
		return false
	}
	t.WaitForCondition(
		timeout,
		100*time.Millisecond,
		getDefLink,
		"Wait for Def link",
	)
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
	t.WaitForCondition(
		timeout,
		100*time.Millisecond,
		getDefLink,
		"Wait for Def Info link",
	)
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
