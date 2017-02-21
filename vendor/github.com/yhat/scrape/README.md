# scrape

A simple, higher level interface for Go web scraping.

When scraping with Go, I find myself redefining tree traversal and other
utility functions.

This package is a place to put some simple tools which build on top of the
[Go HTML parsing library](https://godoc.org/golang.org/x/net/html).

For the full interface check out the godoc
[![GoDoc](https://godoc.org/github.com/yhat/scrape?status.svg)](https://godoc.org/github.com/yhat/scrape)

## Sample

Scrape defines traversal functions like `Find` and `FindAll` while attempting
to be generic. It also defines convenience functions such as `Attr` and `Text`.

```go
// Parse the page
root, err := html.Parse(resp.Body)
if err != nil {
    // handle error
}
// Search for the title
title, ok := scrape.Find(root, scrape.ByTag(atom.Title))
if ok {
    // Print the title
    fmt.Println(scrape.Text(title))
}
```

## A full example: Scraping Hacker News

```go
package main

import (
	"fmt"
	"net/http"

	"github.com/yhat/scrape"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func main() {
	// request and parse the front page
	resp, err := http.Get("https://news.ycombinator.com/")
	if err != nil {
		panic(err)
	}
	root, err := html.Parse(resp.Body)
	if err != nil {
		panic(err)
	}

	// define a matcher
	matcher := func(n *html.Node) bool {
		// must check for nil values
		if n.DataAtom == atom.A && n.Parent != nil && n.Parent.Parent != nil {
			return scrape.Attr(n.Parent.Parent, "class") == "athing"
		}
		return false
	}
	// grab all articles and print them
	articles := scrape.FindAll(root, matcher)
	for i, article := range articles {
		fmt.Printf("%2d %s (%s)\n", i, scrape.Text(article), scrape.Attr(article, "href"))
	}
}
```
