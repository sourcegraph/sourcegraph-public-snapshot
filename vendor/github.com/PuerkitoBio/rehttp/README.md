# rehttp [![build status](https://secure.travis-ci.org/PuerkitoBio/rehttp.png)](http://travis-ci.org/PuerkitoBio/rehttp) [![Go Reference](https://pkg.go.dev/badge/rehttp.svg)](https://pkg.go.dev/rehttp)

Package rehttp implements an HTTP Transport (an `http.RoundTripper`) that handles retries. See the [godoc][] for details.

Please note that rehttp requires Go1.6+, because it uses the `http.Request.Cancel` field to check for cancelled requests. It *should* work on Go1.5, but only if there is no timeout set on the `*http.Client`. Go's stdlib will return an error on the first request if that's the case, because it requires a `RoundTripper` that implements the (now deprecated in Go1.6) `CancelRequest` method.

On Go1.7+, it uses the context returned by `http.Request.Context` to check for cancelled requests.

## Installation

    $ go get github.com/PuerkitoBio/rehttp

## License

The [BSD 3-Clause license][bsd].

[bsd]: http://opensource.org/licenses/BSD-3-Clause
[godoc]: http://godoc.org/github.com/PuerkitoBio/rehttp
