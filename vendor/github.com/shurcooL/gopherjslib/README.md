# gopherjslib [![Build Status](https://travis-ci.org/shurcooL/gopherjslib.svg?branch=master)](https://travis-ci.org/shurcooL/gopherjslib) [![GoDoc](https://godoc.org/github.com/shurcooL/gopherjslib?status.svg)](https://godoc.org/github.com/shurcooL/gopherjslib)

Package gopherjslib provides helpers for in-process GopherJS compilation.

All of them take the optional *Options argument. It can be used to set
a different GOROOT or GOPATH directory or to enable minification.

Example compiling Go code:

	import "github.com/shurcooL/gopherjslib"

	...

	code := strings.NewReader(`
		package main
		import "github.com/gopherjs/gopherjs/js"
		func main() { println(js.Global.Get("window")) }
	`)

	var out bytes.Buffer

	err := gopherjslib.Build(code, &out, nil) // <- default options

Example compiling multiple files:

	var out bytes.Buffer

	builder := gopherjslib.NewBuilder(&out, nil)

	fileA := strings.NewReader(`
		package main
		import "github.com/gopherjs/gopherjs/js"
		func a() { println(js.Global.Get("window")) }
	`)

	builder.Add("a.go", fileA)

	// and so on for each file,then

	err = builder.Build()

Installation
------------

```bash
go get -u github.com/shurcooL/gopherjslib
```
