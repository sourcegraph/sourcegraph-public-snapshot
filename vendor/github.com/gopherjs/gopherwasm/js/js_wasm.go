// Copyright 2018 The GopherWasm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build wasm

package js

import (
	"syscall/js"
)

type Type = js.Type

const (
	TypeUndefined = js.TypeUndefined
	TypeNull      = js.TypeNull
	TypeBoolean   = js.TypeBoolean
	TypeNumber    = js.TypeNumber
	TypeString    = js.TypeString
	TypeSymbol    = js.TypeSymbol
	TypeObject    = js.TypeObject
	TypeFunction  = js.TypeFunction
)

func Global() Value {
	return js.Global()
}

func Null() Value {
	return js.Null()
}

func Undefined() Value {
	return js.Undefined()
}

type Error = js.Error

type Value = js.Value

func ValueOf(x interface{}) Value {
	return js.ValueOf(x)
}

type TypedArray = js.TypedArray

func TypedArrayOf(slice interface{}) TypedArray {
	return js.TypedArrayOf(slice)
}

func GetInternalObject(v Value) interface{} {
	return v
}
