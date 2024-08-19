# link

[![Build Status](https://travis-ci.org/peterhellberg/link.svg?branch=master)](https://travis-ci.org/peterhellberg/link)
[![Go Report Card](https://goreportcard.com/badge/github.com/peterhellberg/link)](https://goreportcard.com/report/github.com/peterhellberg/link)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/peterhellberg/link)
[![License MIT](https://img.shields.io/badge/license-MIT-lightgrey.svg?style=flat)](https://github.com/peterhellberg/link#license-mit)

Parses **Link** headers used for pagination, as defined in [RFC 5988](https://tools.ietf.org/html/rfc5988).

This package was originally based on <https://github.com/swhite24/link>, but **Parse** takes a `string` instead of `*http.Request` in this version.
It also has the convenience functions **ParseHeader**, **ParseRequest** and **ParseResponse**.

## Installation

    go get -u github.com/peterhellberg/link

## Exported functions

 - [Parse(s string) Group](https://godoc.org/github.com/peterhellberg/link#Parse)
 - [ParseHeader(h http.Header) Group](https://godoc.org/github.com/peterhellberg/link#ParseHeader)
 - [ParseRequest(req \*http.Request) Group](https://godoc.org/github.com/peterhellberg/link#ParseRequest)
 - [ParseResponse(resp \*http.Response) Group](https://godoc.org/github.com/peterhellberg/link#ParseResponse)

## Usage

```go
package main

import (
	"fmt"
	"net/http"

	"github.com/peterhellberg/link"
)

func main() {
	for _, l := range link.Parse(`<https://example.com/?page=2>; rel="next"; foo="bar"`) {
		fmt.Printf("URI: %q, Rel: %q, Extra: %+v\n", l.URI, l.Rel, l.Extra)
		// URI: "https://example.com/?page=2", Rel: "next", Extra: map[foo:bar]
	}

	if resp, err := http.Get("https://api.github.com/search/code?q=Println+user:golang"); err == nil {
		for _, l := range link.ParseResponse(resp) {
			fmt.Printf("URI: %q, Rel: %q, Extra: %+v\n", l.URI, l.Rel, l.Extra)
			// URI: "https://api.github.com/search/code?q=Println+user%3Agolang&page=2", Rel: "next", Extra: map[]
			// URI: "https://api.github.com/search/code?q=Println+user%3Agolang&page=34", Rel: "last", Extra: map[]
		}
	}
}
```

## Not supported

 - Extended notation ([RFC 5987](https://tools.ietf.org/html/rfc5987))

## Alternatives to this package

 - [github.com/tent/http-link-go](https://github.com/tent/http-link-go)
 - [github.com/swhite24/link](https://github.com/swhite24/link)

## License (MIT)

Copyright (c) 2015-2019 [Peter Hellberg](https://c7.se)

> Permission is hereby granted, free of charge, to any person obtaining
> a copy of this software and associated documentation files (the
> "Software"), to deal in the Software without restriction, including
> without limitation the rights to use, copy, modify, merge, publish,
> distribute, sublicense, and/or sell copies of the Software, and to
> permit persons to whom the Software is furnished to do so, subject to
> the following conditions:

> The above copyright notice and this permission notice shall be
> included in all copies or substantial portions of the Software.

> THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
> EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
> MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
> NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
> LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
> OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
> WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
