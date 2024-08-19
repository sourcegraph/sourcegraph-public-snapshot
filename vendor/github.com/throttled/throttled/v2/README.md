# Throttled [![Build Status](https://github.com/throttled/throttled/workflows/throttled%20CI/badge.svg)](https://github.com/throttled/throttled/actions) [![Go Reference](https://pkg.go.dev/badge/github.com/throttled/throttled/v2.svg)](https://pkg.go.dev/github.com/throttled/throttled/v2)

Package throttled implements rate limiting using the [generic cell rate
algorithm][gcra] to limit access to resources such as HTTP endpoints.

The 2.0.0 release made some major changes to the throttled API. If
this change broke your code in problematic ways or you wish a feature
of the old API had been retained, please open an issue.  We don't
guarantee any particular changes but would like to hear more about
what our users need. Thanks!

## Installation

Go Modules are required to use Throttled (check that there's a `go.mod` in your
package's root). Import Throttled:

``` go
import (
	"github.com/throttled/throttled/v2"
)
```

Then any of the standard Go tooling like `go build`, `go test`, will find the
package automatically.

You can also pull it into your project using `go get`:

```sh
go get -u github.com/throttled/throttled/v2
```

### Upgrading from the pre-Modules version

The current `/v2` of Throttled is perfectly compatible with the pre-Modules
version of Throttled, but when upgrading, you'll have to add `/v2` to your
imports. Sorry about the churn, but because Throttled was already on its
semantic version 2 by the time Go Modules came around, its tooling didn't play
nice because it expects the major version in the path to match the major in
its tags.

## Documentation

API documentation is available on [godoc.org][doc].

## Usage

This example demonstrates the usage of `HTTPLimiter` for rate-limiting access to
an `http.Handler` to 20 requests per path per minute with bursts of up to 5
additional requests:

```go
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/throttled/throttled/v2"
	"github.com/throttled/throttled/v2/store/memstore"
)

func myHandlerFunc(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, world")
}

func main() {
	store, err := memstore.NewCtx(65536)
	if err != nil {
		log.Fatal(err)
	}

	quota := throttled.RateQuota{
		MaxRate:  throttled.PerMin(20),
		MaxBurst: 5,
	}
	rateLimiter, err := throttled.NewGCRARateLimiterCtx(store, quota)
	if err != nil {
		log.Fatal(err)
	}

	httpRateLimiter := throttled.HTTPRateLimiterCtx{
		RateLimiter: rateLimiter,
		VaryBy:      &throttled.VaryBy{Path: true},
	}

	handler := http.HandlerFunc(myHandlerFunc)
	http.ListenAndServe(":8080", httpRateLimiter.RateLimit(handler))
}
```

### Upgrading to `context.Context` aware version of `throttled`

To upgrade to the new `context.Context` aware version of `throttled`, update
the package to the latest version and replace the following function with their
context-aware equivalent:

- `memstore.New` => `memstore.NewCtx`
- `goredisstore.New` => `goredisstore.NewCtx`
- `redigostore.New` => `redigostore.NewCtx`
- `throttled.NewGCRARateLimiter` => `throttled.NewGCRARateLimiterCtx`
- `throttled.HTTPRateLimiter` => `throttled.HTTPRateLimiterCtx`

Please note that not all stores make use of the passed `context.Context` yet.

## Related Projects

See [throttled/gcra][throttled-gcra] for a list of other projects related to
rate limiting and GCRA.

## Release

1. Update `CHANGELOG.md`. Please use semantic versioning and the existing
   conventions established in the file. Commit the changes with a message like
   `Bump version to 2.2.0`.
2. Tag `master` with a new version prefixed with `v`. For example, `v2.2.0`.
3. `git push origin master --tags`.
4. Publish a new release on the [releases] page. Copy the body from the
   contents of `CHANGELOG.md` for the version and follow other conventions from
   previous releases.

## License

The [BSD 3-clause license][bsd]. Copyright (c) 2014 Martin Angers and contributors.

[blog]: http://0value.com/throttled--guardian-of-the-web-server
[bsd]: https://opensource.org/licenses/BSD-3-Clause
[doc]: https://godoc.org/github.com/throttled/throttled
[gcra]: https://en.wikipedia.org/wiki/Generic_cell_rate_algorithm
[puerkitobio]: https://github.com/puerkitobio/
[pr]: https://github.com/throttled/throttled/compare
[releases]: https://github.com/throttled/throttled/releases
[throttled-gcra]: https://github.com/throttled/gcra

<!--
# vim: set tw=79:
-->
