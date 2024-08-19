// Copyright 2019 The GopherWasm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build wasm
// +build go1.12

package js

import (
	"syscall/js"
)

type Func = js.Func

func FuncOf(fn func(this Value, args []Value) interface{}) Func {
	return js.FuncOf(fn)
}
