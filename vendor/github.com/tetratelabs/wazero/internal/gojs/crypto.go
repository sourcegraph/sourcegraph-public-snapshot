package gojs

import (
	"context"

	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/internal/gojs/custom"
	"github.com/tetratelabs/wazero/internal/gojs/goos"
	"github.com/tetratelabs/wazero/internal/wasm"
)

// jsCrypto is used by crypto/rand.Read to gets random values.
//
// It has only one invocation pattern:
//
//	jsCrypto.Call("getRandomValues", a /* uint8Array */)
//
// This is defined as `Get("crypto")` in rand_js.go init
var jsCrypto = newJsVal(goos.RefJsCrypto, custom.NameCrypto).
	addFunction(custom.NameCryptoGetRandomValues, cryptoGetRandomValues{})

// cryptoGetRandomValues implements jsFn
type cryptoGetRandomValues struct{}

func (cryptoGetRandomValues) invoke(_ context.Context, mod api.Module, args ...interface{}) (interface{}, error) {
	randSource := mod.(*wasm.ModuleInstance).Sys.RandSource()

	r := args[0].(*goos.ByteArray)
	n, err := randSource.Read(r.Unwrap())
	return uint32(n), err
}
