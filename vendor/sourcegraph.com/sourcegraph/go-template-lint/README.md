# go-template-lint

`go-template-lint` is a linter for Go
[text/template](http://golang.org/pkg/text/template/) (and
html/template) template files.


## Checks

* unused template functions (e.g., your `FuncMap` defines `f` but none
  of your templates call it)


## Usage

```
go get sourcegraph.com/sourcegraph/go-template-lint
go-template-lint -f=<file-with-FuncMap.go> -t=<file-with-template-[][]string-list> -td=<base-template-dir>
```

The `file-with-FuncMap.go` option should be a Go source file that
contains a `FuncMap` literal, such as:

```
package foo

import "text/template" // html/template and/or other import aliases are also detected

// ...
// can be nested in any block
  template.FuncMap{
    "f": myFunc,
    "g": func(v string) string { /* ... */ },
  }
// ...
```

The `file-with-template-[][]string-list.go` option should be a Go
source file that contains a list of top-level templates and other
template files (relative to the Go file) to include, such as:

```
package foo

// ...
// can be nested in any block
  [][]string{
    {"profile.html", "common.html", "layout.html"},
    {"edit.html", "common.html", "layout.html"},
  }
// ...
```

The `base-template-dir` should be the directory that contains your Go
templates and that the template filenames in your code are relative
to. For example, if the template files above (profile,html,
common.html, etc.) were stored in `app/templates`, we'd use
`-td=app/templates`.
