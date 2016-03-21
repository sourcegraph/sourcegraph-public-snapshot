==============================================
go-selenium - Selenium WebDriver client for Go
==============================================

[![status](https://sourcegraph.com/api/repos/sourcegraph.com/sourcegraph/go-selenium/badges/status.png)](https://sourcegraph.com/sourcegraph/go-selenium)

go-selenium is a [Selenium](http://seleniumhq.org) WebDriver client for [Go](http://golang.org).

Note: the public API is experimental and subject to change until further notice.


Usage
=====

Documentation: [go-selenium on Sourcegraph](https://sourcegraph.com/sourcegraph/go-selenium).

Example: see example_test.go:

```go
package selenium_test

import (
	"fmt"
	"sourcegraph.com/sourcegraph/go-selenium"
)

func ExampleFindElement() {
	var webDriver selenium.WebDriver
	var err error
	caps := selenium.Capabilities(map[string]interface{}{"browserName": "firefox"})
	if webDriver, err = selenium.NewRemote(caps, "http://localhost:4444/wd/hub"); err != nil {
		fmt.Printf("Failed to open session: %s\n", err)
		return
	}
	defer webDriver.Quit()

	err = webDriver.Get("https://sourcegraph.com/sourcegraph/go-selenium")
	if err != nil {
		fmt.Printf("Failed to load page: %s\n", err)
		return
	}

	if title, err := webDriver.Title(); err == nil {
		fmt.Printf("Page title: %s\n", title)
	} else {
		fmt.Printf("Failed to get page title: %s", err)
		return
	}

	var elem selenium.WebElement
	elem, err = webDriver.FindElement(selenium.ByCSSSelector, ".repo .name")
	if err != nil {
		fmt.Printf("Failed to find element: %s\n", err)
		return
	}

	if text, err := elem.Text(); err == nil {
		fmt.Printf("Repository: %s\n", text)
	} else {
		fmt.Printf("Failed to get text of element: %s\n", err)
		return
	}

	// output:
	// Page title: go-selenium - Sourcegraph
	// Repository: go-selenium
}
```

The `WebDriverT` and `WebElementT` interfaces make test code cleaner. Each method in
`WebDriver` and `WebElement` has a corresponding method in the `*T` interfaces that omits the error
from the return values and instead calls `t.Fatalf` upon encountering an error. For example:

```go
package mytest

import (
  "sourcegraph.com/sourcegraph/go-selenium"
  "testing"
)

var caps selenium.Capabilities
var executorURL = "http://localhost:4444/wd/hub"

// An example test using the WebDriverT and WebElementT interfaces. If you use the non-*T
// interfaces, you must perform error checking that is tangential to what you are testing,
// and you have to destructure results from method calls.
func TestWithT(t *testing.T) {
  wd, _ := selenium.NewRemote(caps, executor)

  // Call .T(t) to obtain a WebDriverT from a WebDriver (or to obtain a WebElementT from
  // a WebElement).
  wdt := wd.T(t)

  // Calls `t.Fatalf("Get: %s", err)` upon failure.
  wdt.Get("http://example.com")

  // Calls `t.Fatalf("FindElement(by=%q, value=%q): %s", by, value, err)` upon failure.
  elem := wdt.FindElement(selenium.ByCSSSelector, ".foo")

  // Calls `t.Fatalf("Text: %s", err)` if the `.Text()` call fails.
  if elem.Text() != "bar" {
    t.Fatalf("want elem text %q, got %q", "bar", elem.Text())
  }
}
```

See remote_test.go for more usage examples.



Running tests
=============

Start Selenium WebDriver and run `go test`. To see all available options, run `go test -test.h`.


TODO
====

* Support Firefox profiles


Contributors
============

* Quinn Slack <sqs@sourcegraph.com>
* Miki Tebeka <miki.tebeka@gmail.com> (go-selenium is based on Miki's
  [bitbucket.org/tebekas/selenium](https://bitbucket.org/tebeka/selenium) library)


License
=======

go-selenium is distributed under the Eclipse Public License.
