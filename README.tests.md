# Sourcegraph tests

Most of the tests in this repository are written using the Go testing
package. This document describes conventions for writing these tests.


## Running tests

PhantomJS is required for browser tests. Install with `npm install -g
phantomjs`. You may have to prefix that with `sudo`. Afterwards,
`phantomjs` should be in your `$PATH`.

By default, you also need to be running the webpack dev server (`cd
app && npm start` in a separate window).

The Makefile defines 4 test targets:

* `make test`: runs all tests, including integration tests. This can
  take a while (~3min). During development, it's quicker to run one of the
  other targets.
* `make smtest`: runs only unit tests. This is fast (~5sec).
* `make mdtest`: runs unit tests, exec tests, DB tests, and network
  tests, but not super-long integration tests (e.g., browser and build
  tests). This is moderately fast (~30sec).
* `make lgtest`: same as `make test`, included for completeness.

(See
[Test Sizes on the Google testing blog](http://googletesting.blogspot.com/2010/12/test-sizes.html)
for why we chose these names for our groups of tests.)


## Writing Go tests

We use [Go build tags](http://golang.org/pkg/go/build/) to selectively
compile and execute tests.

Each `_test.go` file contains zero or more build tags describing the
kinds of tests contained in the file.

* **no build tag**: fast unit tests that only check Go logic and do
  not use the database, network, or external programs. This is the
  most common kind of test.
* **exectest**: tests that spawn a `src` child process
* **pgsqltest**: tests that require access to the PostgreSQL database.
* **nettest**: tests that require access to the network. We currently
  don't distinguish between tests that require localhost access and
  tests that require Internet access, but we may do so later.
* **githubtest**: tests that require access to the GitHub API.
* **buildtest**: tests that execute srclib to build repositories.

**Why?** Using build tags encourages writing tests that use the minimum set of
resources (and therefore execute as quickly as possible). To enforce
the use of these build tags, test helper functions that give access to
potentially each kind of resource (e.g., funcs to initialize a test DB
or a GitHub API client) are also defined in files with the
corresponding build tag. This means that the compiler will prevent you
from using a resource that you haven't explicitly declared. (If we
didn't use build tags and just used `-test.run=`, then it would be far
easier to accidentally write a unit test that relies on slow external
resources.)
