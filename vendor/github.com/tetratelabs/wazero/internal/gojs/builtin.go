package gojs

import (
	"github.com/tetratelabs/wazero/internal/gojs/config"
	"github.com/tetratelabs/wazero/internal/gojs/goos"
)

// newJsGlobal = js.Global() // js.go init
func newJsGlobal(config *config.Config) *jsVal {
	var fetchProperty interface{} = goos.Undefined
	proc := &processState{
		cwd:   config.Workdir,
		umask: config.Umask,
	}

	return newJsVal(goos.RefValueGlobal, "global").
		addProperties(map[string]interface{}{
			"Object":     objectConstructor,
			"Array":      arrayConstructor,
			"crypto":     jsCrypto,
			"Uint8Array": uint8ArrayConstructor,
			"fetch":      fetchProperty,
			"process":    newJsProcess(proc),
			"fs":         newJsFs(proc),
			"Date":       jsDateConstructor,
		})
}

var (
	// Values below are not built-in, but verifiable by looking at Go's source.
	// When marked "XX.go init", these are eagerly referenced during syscall.init

	// jsGo is not a constant

	// objectConstructor is used by js.ValueOf to make `map[string]any`.
	//	Get("Object") // js.go init
	objectConstructor = newJsVal(goos.RefObjectConstructor, "Object")

	// arrayConstructor is used by js.ValueOf to make `[]any`.
	//	Get("Array") // js.go init
	arrayConstructor = newJsVal(goos.RefArrayConstructor, "Array")

	// uint8ArrayConstructor = js.Global().Get("Uint8Array")
	//	// fs_js.go, rand_js.go init
	//
	// It has only one invocation pattern: `buf := uint8Array.New(len(b))`
	uint8ArrayConstructor = newJsVal(goos.RefUint8ArrayConstructor, "Uint8Array")
)
