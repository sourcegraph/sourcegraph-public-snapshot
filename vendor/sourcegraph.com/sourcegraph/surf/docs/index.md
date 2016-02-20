# Surf
[![Build Status](https://img.shields.io/travis/headzoo/surf/master.svg?style=flat-square)](https://travis-ci.org/headzoo/surf)
[![Github](https://img.shields.io/badge/source-github-blue.svg?style=flat-square)](https://github.com/headzoo/surf/)
[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)](https://raw.githubusercontent.com/headzoo/surf/master/LICENSE.md)
[![GitHub Stars](https://img.shields.io/github/stars/headzoo/surf.svg?style=flat-square)](https://github.com/headzoo/surf/stargazers)
[![GitHub Forks](https://img.shields.io/github/forks/headzoo/surf.svg?style=flat-square)](https://github.com/headzoo/surf/network)

Surf is a Go (golang) library that implements a virtual web browser that you control pragmatically.
Surf isn't just another Go solution for downloading content from the web. Surf is designed to behave like web
browser, and includes: cookie management, history, bookmarking, user agent spoofing (with a nifty user agent
builder), submitting forms, DOM selection and traversal via jQuery style CSS selectors, scraping assets like images,
stylesheets, and other features.


### Installation
Download Surf using go.

```sh
$ go get github.com/headzoo/surf
```

Import the library into your project.

```go
import "github.com/headzoo/surf"
```


### Quick Start

```go
package main

import (
	"github.com/headzoo/surf"
	"fmt"
)

func main() {
	// Create a new browser and open reddit.
	bow := surf.NewBrowser()
	err := bow.Open("http://reddit.com")
	if err != nil {
		panic(err)
	}
	
	// Outputs: "reddit: the front page of the internet"
	fmt.Println(bow.Title())
	
	// Click the link for the newest submissions.
	bow.Click("a.new")
	
	// Outputs: "newest submissions: reddit.com"
    fmt.Println(bow.Title())
    
    // Log in to the site.
    fm, _ := bow.Form("form.login-form")
    fm.Input("user", "JoeRedditor")
    fm.Input("passwd", "d234rlkasd")
    if fm.Submit() != nil {
    	panic(err)
    }
    
    // Go back to the "newest submissions" page, bookmark it, and
    // print the title of every link on the page.
    bow.Back()
    bow.Bookmark("reddit-new")
    bow.Find("a.title").Each(func(_ int, s *goquery.Selection) {
        fmt.Println(s.Text())
    })
}
```

Read the [Overview](overview) documentation for more information.
