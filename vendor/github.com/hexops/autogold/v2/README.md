# autogold - automatically update your Go tests <a href="https://hexops.com"><img align="right" alt="Hexops logo" src="https://raw.githubusercontent.com/hexops/media/master/readme.svg"></img></a>

<a href="https://pkg.go.dev/github.com/hexops/autogold"><img src="https://pkg.go.dev/badge/badge/github.com/hexops/autogold.svg" alt="Go Reference" align="right"></a>
  
[![Go CI](https://github.com/hexops/autogold/workflows/Go%20CI/badge.svg)](https://github.com/hexops/autogold/actions) [![codecov](https://codecov.io/gh/hexops/autogold/branch/main/graph/badge.svg)](https://codecov.io/gh/hexops/autogold) [![Go Report Card](https://goreportcard.com/badge/github.com/hexops/autogold)](https://goreportcard.com/report/github.com/hexops/autogold)

autogold makes `go test -update` automatically update your Go tests (golden files and Go values in e.g. `foo_test.go`).

~5m introduction available on YouTube:

[_"It's 2021: you shouldn't have to update Go tests manually"_](https://youtu.be/mvkIruEGRr4)

## Installation

```sh
go get -u github.com/hexops/autogold/v2
```

## Automatic golden files

Write in a Go test:

```Go
import "github.com/hexops/autogold/v2"
...
autogold.ExpectFile(t, got)
```

`go test -update` will now create/update a `testdata/<test name>.golden` file for you automatically. If your tests change over time you can use `go test -update -clean` to also have it remove _unused_ golden files.

## Automatic inline test updating

In a Go test, simply call `autogold.Expect(want).Equal(t, got)`, passing `nil` as the value you `want` initially:

```Go
func TestFoo(t *testing.T) {
	...
	autogold.Expect(nil).Equal(t, got)
}
```

Run `go test -update` and autogold will automatically update the `autogold.Expect(want)` Go syntax with the actual value your test `got`. It works with complex Go structs, slices, strings, etc.

## Diffs

Anytime your test produces a result that is unexpected, you'll get a nice diff showing exactly what changed. It does this by [converting values at runtime directly to a formatted Go AST](https://github.com/hexops/valast), and using the same [diffing library the Go language server uses](https://github.com/hexops/gotextdiff):

```
--- FAIL: TestFoo (0.08s)
    autogold.go:91: mismatch (-want +got):
        --- want
        +++ got
        @@ -1 +1 @@
        +&example.Baz{Name: "Jane", Age: 31}
```

## Subtesting

[Table-driven Go subtests](https://blog.golang.org/subtests) are supported nicely as you can call `.Equal(got)` later, so that `go test -update` will update your table-driven test values defined earlier for you:

```Go
func TestTime(t *testing.T) {
	testCases := []struct {
		gmt    string
		loc    string
		expect autogold.Value // autogold: the value we expect
	}{
		{"12:31", "Europe/Zuri", autogold.Expect(nil)},
		{"12:31", "America/New_York", autogold.Expect(nil)},
		{"08:08", "Australia/Sydney", autogold.Expect(nil)},
	}
	for _, tc := range testCases {
		t.Run(tc.loc, func(t *testing.T) {
			loc, err := time.LoadLocation(tc.loc)
			if err != nil {
				t.Fatal("could not load location")
			}
			gmt, _ := time.Parse("15:04", tc.gmt)
			got := gmt.In(loc).Format("15:04")

			tc.expect.Equal(t, got) // autogold: tell it the value our test produced
		})
	}
}
```

It works by finding the relevant `autogold.Expect(want)` call for you based on callstack information / matching line number in the file, and then rewrites the `nil` parameter (or any other value that was there.)

## What are golden files, when should they be used?

Golden files are used by the Go authors for testing [the standard library](https://golang.org/src/go/doc/doc_test.go), the [`gofmt` tool](https://github.com/golang/go/blob/master/src/cmd/gofmt/gofmt_test.go#L124-L130), etc. and are a common pattern in the Go community for snapshot testing. See also ["Testing with golden files in Go" - Chris Reeves](https://medium.com/soon-london/testing-with-golden-files-in-go-7fccc71c43d3)

_Golden files make the most sense when you'd otherwise have to write a complex multi-line string or large Go structure inline in your test, making it hard to read._

In most cases, you should prefer inline snapshots, subtest golden values, or traditional Go tests.

## Command line syntax: put `-update` at the end

`-update` should go at the end of your `go test` command, otherwise for some reason stdout will be considered a terminal and color will be turned on for libraries like [fatih/color](https://github.com/fatih/color). Example:

```
go test -count=1 -run TestSomething . -update
```

## Custom formatting

[valast](https://github.com/hexops/valast) is used to produce Go syntax at runtime for the Go value you provide. If the default output is not to your liking, you have options:

- **Pass a string to autogold**: It will be formatted as a Go string for you in the resulting `.golden` file / in Go tests.
- **Use your own formatting (JSON, etc.)**: Make your `got` value of type `autogold.Raw("foobar")`, and it will be used as-is for `.golden` files (not allowed with inline tests.)
- **Exclude unexported fields**: `autogold.ExpectFile(t, got, autogold.ExportedOnly())`

## Backwards compatibility

- As is the case with `gofmt`, different Go versions may produce different formattings (although rare.)
- Minor versions of autogold (e.g. `v1.0`, `v1.1`) may alter the _formatting_ of `.golden` files, although we will be mindful of such changes.
- Major versions of autogold (e.g. `v1`, `v2`) will be used for any major changes in output that would be _difficult to review_ (we expect this will be rare in practice.)

## Alternatives comparison

The following are alternatives to autogold, making note of the differences we found that let us to create autogold:

- [github.com/xorcare/golden](https://pkg.go.dev/github.com/xorcare/golden)
    - Supports `[]byte` inputs only, defers formatting to users.
    - Does not support inline snapshots / code updating.
- [github.com/sebdah/goldie](https://pkg.go.dev/github.com/sebdah/goldie/v2)
    - Supports `[]byte` inputs only, provides helpers for JSON, XML, etc.
    - Does not support inline snapshots / code updating.
- [github.com/bradleyjkemp/cupaloy](https://pkg.go.dev/github.com/bradleyjkemp/cupaloy/v2)
    - Works on `interface{}` inputs.
    - [Uses inactive go-spew project](https://github.com/davecgh/go-spew/issues/128) to format Go structs.
    - Does not support inline snapshots / code updating.

## Changelog

#### v2.2.1

Updated to valast v1.4.4:

* `valast.Addr` is replaced by `valast.Ptr`, which uses Go generics and looks cleaner.
* `time.Time` values are now supported.

#### v2.2.0

* If autogold is used in packages with an `-update` flag already defined, now no conflict occurs. This enables autogold to be used with other 'golden' packages without conflict.
* Fixed an issue where `_test` packages using types from non-test packages would sometimes result in the wrong package name qualifier.

#### v2.1.0

Added support for building in Bazel / working around a bug in Bazel / Go's `packages.Load` functionality. This feature can be enabled using `ENABLE_BAZEL_PACKAGES_LOAD_HACK=true`. For more details see [#40](https://github.com/hexops/autogold/pull/40) and [golang/go#57304](https://github.com/golang/go/issues/57304)

#### v2.0.3

Fixed an issue where updating inline tests could cause a deadlock.

#### v2.0.2

Writing a unique name with inline tests is no longer required. Previously you must write a unique name as the first parameter to `Want` and it must have been inside a `TestFoo` function for autogold to find it:

```go
func TestFoo(t *testing.T) {
	...
	autogold.Want("unique name inside TestFoo", want).Equal(t, nil)
}
```

This can be rewritten as `autogold.Expect(want).Equal(t, got)` and no unique name is required, the function call can be placed anywhere inside your Go test file as autogold will now update the invocation based on callstack information:

```go
func TestFoo(t *testing.T) {
	...
	autogold.Expect(want).Equal(t, nil)
}
```

Additionally, CLI flag behavior has been improved substantially based on experience working in very large enterprise Go codebases:

* `-update` now behaves like `-update-only`, it no longer removes unused golden files which is faster in very large codebases. Instead, you may use `-update -clean` to remove unused golden files. `-update-only` is removed.
* Previously autogold would fail tests when running `-update`, meaning you may need to run `go test -update` many times to get to your desired end-state if updating a lot of test cases. Now we match the behavior of OCaml expect tests in not failing tests by default (you can now specify `-fail-on-update`)
* `-fail-on-update` now uses `t.FailNow()` instead of `t.Fatal()` to allow as many tests as possible to succeed when performing an update operation.
* `autogold.Want` has been deprecated in favor of `autogold.Expect`
* Fixed `invalid cross-device link` errors on some systems.

Finally, please note the renaming of functions before and after:

* Inline tests: `autogold.Want` -> `autogold.Expect`
* File tests: `autogold.Equal` -> `autogold.ExpectFile`

##### Automating the migration with Comby

You can automatically migrate from v1 to v2 using the following [Comby](https://comby.dev) configuration:

<details>
<summary>autogold.comby.toml</summary>

```
# autogold.comby.toml
[update-imports]

match="\"github.com/hexops/autogold\""
rewrite="\"github.com/hexops/autogold/v2\""

[update-api-want]

match="autogold.Want(:[desc], :[v])"
rewrite="autogold.Expect(:[v])"

[update-api-equal]

match="autogold.Equal(:[v])"
rewrite="autogold.ExpectFile(:[v])"
```

Assuming Comby is available on your system, you can run the following command to apply the changes: 

```
$ go get -u github.com/hexops/autogold/v2
$ comby -config autogold.comby.toml -matcher .go -exclude-dir vendor,node_modules -in-place
$ go mod tidy
$ git diff # show the changes applied by comby
```

</details>

#### v1.3.1

* Improved Go code formatting (updated valast and gofumpt versions)
* Added usage of `t.Helper` to improve line:column information of test failures. 
* Fixed an issue where `-update` subtest names could collide and incorrectly fail tests.
* Fixed a data race when `-update` is used
* Diffs are now printed with ANSII color codes
