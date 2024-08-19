// Copyright 2019 The GopherWasm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build wasm
// +build !go1.12

package js

import (
	"syscall/js"
)

type Callback = js.Callback

type EventCallbackFlag = js.EventCallbackFlag

const (
	PreventDefault           = js.PreventDefault
	StopPropagation          = js.StopPropagation
	StopImmediatePropagation = js.StopImmediatePropagation
)

func NewCallback(f func([]Value)) Callback {
	return js.NewCallback(f)
}

func NewEventCallback(flags EventCallbackFlag, fn func(event Value)) Callback {
	return js.NewEventCallback(flags, fn)
}
