# Overview

When `GOOS=js` and `GOARCH=wasm`, Go's compiler targets WebAssembly Binary
format (%.wasm).

Wazero's "github.com/tetratelabs/wazero/experimental/gojs" package allows you to run
a `%.wasm` file compiled by Go.  This is similar to what is implemented in
[wasm_exec.js][1]. See https://wazero.io/languages/go/ for more.

## Example

wazero includes an [example](example) that implements the `cat` utility.

## Experimental

Go defines js "EXPERIMENTAL... exempt from the Go compatibility promise."
Accordingly, wazero cannot guarantee this will work from release to release,
or that usage will be relatively free of bugs. Moreover, [`GOOS=wasip1`][2]
will be shipped in Go 1.21. wazero will remove this package after Go 1.22 is
released.

Due to these concerns and the relatively high implementation overhead, most
will choose TinyGo instead of gojs.

[1]: https://github.com/golang/go/blob/go1.20/misc/wasm/wasm_exec.js
[2]: https://github.com/golang/go/issues/58141
