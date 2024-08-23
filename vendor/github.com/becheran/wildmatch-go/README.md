# wildmatch go

[![](https://godoc.org/github.com/becheran/wildmatch-go?status.svg)](https://godoc.org/github.com/becheran/wildmatch-go)
[![GoDev](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/becheran/wildmatch-go)
[![Sourcegraph](https://sourcegraph.com/github.com/becheran/wildmatch-go/-/badge.svg)](https://sourcegraph.com/github.com/becheran/wildmatch-go?badge)
[![Go Report Card](https://goreportcard.com/badge/becheran/wildmatch-go)](https://goreportcard.com/report/becheran/wildmatch-go)
![GitHub Workflow Status](https://github.com/becheran/wildmatch-go/workflows/CI/badge.svg)

*golang* library of the original *rust* [wildmatch library](https://github.com/becheran/wildmatch).

``` sh
go get github.com/becheran/wildmatch-go
```

Match strings against a simple wildcard pattern. Tests a wildcard pattern `p` against an input string `s`. Returns true only when `p` matches the entirety of `s`.

See also the example described on [wikipedia](https://en.wikipedia.org/wiki/Matching_wildcards) for matching wildcards.

- `?` matches exactly one occurrence of any character.
- `*` matches arbitrary many (including zero) occurrences of any character.
- No escape characters are defined.

For example the pattern `ca?` will match `cat` or `car`. The pattern `https://*` will match all https urls, such as `https://google.de` or `https://github.com/becheran/wildmatch`.

The library only depends on the go standard library.
