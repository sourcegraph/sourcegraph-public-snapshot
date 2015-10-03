# vfsgen [![Build Status](https://travis-ci.org/shurcooL/vfsgen.svg?branch=master)](https://travis-ci.org/shurcooL/vfsgen) [![GoDoc](https://godoc.org/github.com/shurcooL/vfsgen?status.svg)](https://godoc.org/github.com/shurcooL/vfsgen)

Package vfsgen takes an input http.FileSystem (likely at `go generate` time) and
generates Go code that statically implements the given http.FileSystem.

Features:

-	Efficient generated code without unneccessary overhead.

-	Uses gzip compression internally (selectively, only for files that compress well).

-	Enables direct access to internal gzip compressed bytes via an optional interface.

-	Outputs `gofmt`ed Go code.

Installation
------------

```bash
go get -u github.com/shurcooL/vfsgen
```

Usage
-----

This code will generate an assets_vfsdata.go file with `var assets http.FileSystem = ...` that statically implements the contents of "assets" directory.

```Go
var fs http.FileSystem = http.Dir("assets")

err := vfsgen.Generate(fs, vfsgen.Options{})
if err != nil {
	log.Fatalln(err)
}
```

Then, in your program, you can use `assets` as any other [`http.FileSystem`](https://godoc.org/net/http#FileSystem), for example:

```Go
file, err := assets.Open("/some/file.txt")
if err != nil { ... }
defer file.Close()
```

```Go
http.Handle("/assets/", http.FileServer(assets))
```

### `go generate` Usage

vfsgen is great to use with go generate directives. The code above can go in an assets_gen.go file, which can then be invoked via "//go:generate go run assets_gen.go". The input virtual filesystem can read directly from disk, or it can be more involved.

By using build tags, you can create a development mode where assets are loaded directly from disk via `http.Dir`, but then statically implemented for final releases.

See [shurcooL/Go-Package-Store#38](https://github.com/shurcooL/Go-Package-Store/pull/38) for a complete example of such use.

### Additional Embedded Information

All compressed files implement this interface for efficient direct access to the internal compressed bytes:

```Go
interface {
	// GzipBytes returns gzip compressed contents of the file.
	GzipBytes() []byte
}
```

Files that have been determined to not be worth gzip compressing (their compressed size is larger than original) implement this interface:

```Go
interface {
	// NotWorthGzipCompressing indicates the file is not worth gzip compressing.
	NotWorthGzipCompressing()
}
```

Comparison
----------

vfsgen aims to be conceptually simple to use. The [`http.FileSystem`](https://godoc.org/net/http#FileSystem) abstraction is central to vfsgen. It's used as both input for code generation, and as output in the generated code.

That enables great flexibility through orthogonality, since helpers and wrappers can operate on `http.FileSystem` without knowing about vfsgen. If you want, you can perform pre-processing, minifying assets, merging folders, filtering out files and otherwise modifying input via generic `http.FileSystem` middleware.

It avoids unneccessary overhead by merging what was previously done with two distinct packages into a single package.

It strives to be the best in its class in terms of code quality and efficiency of generated code. However, if your use goals are different, there are other similar packages that may fit your needs better.

### Alternatives

-	[`go-bindata`](https://github.com/jteeuwen/go-bindata) - Reads from disk, generates Go code that provides access to data via a [custom API](https://github.com/jteeuwen/go-bindata#accessing-an-asset).
-	[`go-bindata-assetfs`](https://github.com/elazarl/go-bindata-assetfs) - Takes output of go-bindata and provides a wrapper that implements `http.FileSystem` interface (the same as what vfsgen outputs directly).
-	[`becky`](https://github.com/tv42/becky) - Embeds assets as string literals in Go source.
-	[`statik`](https://github.com/rakyll/statik) - Embeds a directory of static files to be accessed via `http.FileSystem` interface (sounds very similar to vfsgen); implementation sourced from [camlistore](https://camlistore.org).
-	[`go.rice`](https://github.com/GeertJohan/go.rice) - Makes working with resources such as HTML, JS, CSS, images and templates very easy.

Attribution
-----------

This package was originally based on the excellent work by [@jteeuwen](https://github.com/jteeuwen) on [`go-bindata`](https://github.com/jteeuwen/go-bindata) and [@elazarl](https://github.com/elazarl) on [`go-bindata-assetfs`](https://github.com/elazarl/go-bindata-assetfs).

License
-------

-	[MIT License](http://opensource.org/licenses/mit-license.php)
