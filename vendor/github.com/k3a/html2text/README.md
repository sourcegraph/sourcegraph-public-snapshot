[![GoDoc](https://godoc.org/github.com/k3a/html2text?status.svg)](https://godoc.org/github.com/k3a/html2text)
[![Build Status](https://travis-ci.org/k3a/html2text.svg?branch=master)](https://travis-ci.org/k3a/html2text)
[![Coverage Status](https://coveralls.io/repos/github/k3a/html2text/badge.svg?branch=master)](https://coveralls.io/github/k3a/html2text?branch=master)
[![Report Card](https://goreportcard.com/badge/github.com/k3a/html2text)](https://goreportcard.com/report/github.com/k3a/html2text)

# html2text

A simple Golang package to convert HTML to plain text (without non-standard dependencies).

It converts HTML tags to text and also parses HTML entities into characters they represent.
A `<head>` section of the HTML document, as well as most other tags are stripped out but 
links are properly converted into their href attribute.

It can be used for converting HTML emails into text.

Some tests are installed as well.
Uses semantic versioning and no breaking changes are planned.

Fell free to publish a pull request if you have suggestions for improvement but please note that the library can now be considered feature-complete and API stable. If you need more than this basic conversion, please use an alternative mentioned at the bottom.

## Install
```bash
go get github.com/k3a/html2text
```

## Usage

```go
package main

import (
	"fmt"
	"github.com/k3a/html2text"
)

func main() {
	html := `<html><head><title>Good</title></head><body><strong>clean</strong> text</body>`
	
	plain := html2text.HTML2Text(html)
			  
	fmt.Println(plain)
}

/*	Outputs:

	clean text
*/

```

To see all features, please look info `html2text_test.go`.

## Alternatives
- https://github.com/jaytaylor/html2text (heavier, with more features)
- https://git.alexwennerberg.com/nanohtml2text (rewrite of this module in Rust)

## License

MIT

