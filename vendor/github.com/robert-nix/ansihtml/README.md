# ansihtml

[![Travis](https://img.shields.io/travis/robert-nix/ansihtml.svg)](https://travis-ci.org/robert-nix/ansihtml/) [![codecov](https://codecov.io/gh/robert-nix/ansihtml/branch/master/graph/badge.svg)](https://codecov.io/gh/robert-nix/ansihtml) [![PkgGoDev](https://pkg.go.dev/badge/github.com/robert-nix/ansihtml)](https://pkg.go.dev/github.com/robert-nix/ansihtml)

> Go package to parse ANSI escape sequences to HTML.

## Usage

```go
html := ansihtml.ConvertToHTML([]byte("\x1b[33mThis text is yellow.\x1b[m"))
// html: `<span style="color:olive;">This text is yellow.</span>`

html := ansihtml.ConvertToHTMLWithClasses([]byte("\x1b[31mThis text is red."), "ansi-", false)
// html: `<span class="ansi-fg-red">This text is red.</span>`
```
