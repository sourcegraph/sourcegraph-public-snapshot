# valast - convert Go values to their AST <a href="https://hexops.com"><img align="right" alt="Hexops logo" src="https://raw.githubusercontent.com/hexops/media/master/readme.svg"></img></a>

<a href="https://pkg.go.dev/github.com/hexops/valast"><img src="https://pkg.go.dev/badge/badge/github.com/hexops/valast.svg" alt="Go Reference" align="right"></a>
  
[![Go CI](https://github.com/hexops/valast/workflows/Go%20CI/badge.svg)](https://github.com/hexops/valast/actions) [![codecov](https://codecov.io/gh/hexops/valast/branch/main/graph/badge.svg?token=Iw1FdYk0m8)](https://codecov.io/gh/hexops/valast) [![Go Report Card](https://goreportcard.com/badge/github.com/hexops/valast)](https://goreportcard.com/report/github.com/hexops/valast)

Valast converts Go values at runtime into their `go/ast` equivalent, e.g.:

```Go
x := &foo.Bar{
    a: "hello world!",
    B: 1.234,
}
fmt.Println(valast.String(x))
```

Prints string:

```Go
&foo.Bar{a: "hello world!", B: 1.234}
```

## What is this useful for?

This can be useful for debugging and testing, you may think of it as a more comprehensive and configurable version of the `fmt` package's `%+v` and `%#v` formatting directives. It is similar to e.g. `repr` in Python.

## Features

- Produces Go code via a `go/ast`, defers formatting to the best-in-class Go formatter [gofumpt](https://github.com/mvdan/gofumpt).
- Fully handles unexported fields, types, and values (optional.)
- Strong emphasis on being used for producing valid Go code that can be copy & pasted directly into e.g. tests.
- [Extensively tested](https://github.com/hexops/valast/tree/main/testdata), over 88 tests and handling numerous edge cases (such as pointers to unaddressable literal values like `&"foo"` properly, and even [finding bugs in alternative packages'](https://github.com/shurcooL/go-goon/issues/15)).
- Provide custom AST representations for your types with `valast.RegisterType(...)`.

## Alternatives comparison

The following are alternatives to Valast, making note of the differences we found that let us to create Valast:

- [github.com/davecgh/go-spew](https://github.com/davecgh/go-spew)
    - [may be inactive](https://github.com/davecgh/go-spew/issues/128)
    - Produces Go-like output, but not Go syntax.
- [github.com/shurcooL/go-goon](https://github.com/shurcooL/go-goon) (based on go-spew)
    - Produces valid Go syntax, but not via a `go/ast`.
    - [Produces less idiomatic/terse results](https://github.com/shurcooL/go-goon/issues/11))
    - Was deprecated in favor of valast.
- [github.com/alecthomas/repr](https://github.com/alecthomas/repr)
    - Produces Go syntax, but not always valid code (e.g. can emit illegal `&23`, whereas Valast will emit a valid expression `valast.Addr(23).(int)`), not via a `go/ast`.
    - [Does not handle unexported fields/types/values.](https://github.com/alecthomas/repr/pull/13)

You may also wish to look at [autogold](https://github.com/hexops/autogold) and [go-cmp](https://github.com/google/go-cmp), which aim to solve the "compare Go values in a test" problem.
