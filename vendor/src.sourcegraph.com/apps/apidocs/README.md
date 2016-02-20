# apidocs

This is a [Sourcegraph app](https://src.sourcegraph.com/sourcegraph) which
automatically generates API documentation from your code using the power of the
graph!

## Overview

It makes use of the [(*sourcegraph.DefsClient).List](https://godoc.org/src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph#DefsClient)
method in order to list and find documentation for all the definitions
(functions, variables, etc) in your code and allows you to view the API for any
given directory of code.

# Development

In order to develop easier, you can first `go generate -tags=dev ./...` which
will cause assets to be loaded from disk instead of statically. After
development you should always `go generate ./...` without `-tags=dev` and commit
the result.

## TODO

- Form a better UI with something more realistic (React, GopherJS, etc) instead
  of the highly-repetitive Go template (`home.html`) we use currently. Split
  apart the stylesheet from `home.html` etc.
- Remove the `grep -r HACK .` hacks.
- Address the `grep -r TODO .` todos.
