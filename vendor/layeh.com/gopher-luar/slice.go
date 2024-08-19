package luar

import (
	"reflect"

	"github.com/yuin/gopher-lua"
)

func sliceIndex(L *lua.LState) int {
	ref, mt := check(L, 1)
	key := L.CheckAny(2)

	switch converted := key.(type) {
	case lua.LNumber:
		index := int(converted)
		if index < 1 || index > ref.Len() {
			L.ArgError(2, "index out of range")
		}
		val := ref.Index(index - 1)
		if (val.Kind() == reflect.Struct || val.Kind() == reflect.Array) && val.CanAddr() {
			val = val.Addr()
		}
		L.Push(New(L, val.Interface()))
	case lua.LString:
		if fn := mt.method(string(converted)); fn != nil {
			L.Push(fn)
			return 1
		}
		return 0
	default:
		L.ArgError(2, "must be a number or string")
	}
	return 1
}

func sliceNewIndex(L *lua.LState) int {
	ref, _ := check(L, 1)
	index := L.CheckInt(2)
	value := L.CheckAny(3)

	if index < 1 || index > ref.Len() {
		L.ArgError(2, "index out of range")
	}
	val, err := lValueToReflect(L, value, ref.Type().Elem(), nil)
	if err != nil {
		L.ArgError(3, err.Error())
	}
	ref.Index(index - 1).Set(val)
	return 0
}

func sliceLen(L *lua.LState) int {
	ref, _ := check(L, 1)

	L.Push(lua.LNumber(ref.Len()))
	return 1
}

func sliceCall(L *lua.LState) int {
	ref, _ := check(L, 1)

	i := 0
	fn := func(L *lua.LState) int {
		if i >= ref.Len() {
			return 0
		}
		item := ref.Index(i).Interface()
		L.Push(lua.LNumber(i + 1))
		L.Push(New(L, item))
		i++
		return 2
	}

	L.Push(L.NewFunction(fn))
	return 1
}

func sliceAdd(L *lua.LState) int {
	ref, _ := check(L, 1)
	item := L.CheckAny(2)

	hint := ref.Type().Elem()
	value, err := lValueToReflect(L, item, hint, nil)
	if err != nil {
		L.ArgError(2, err.Error())
	}

	ref = reflect.Append(ref, value)
	L.Push(New(L, ref.Interface()))
	return 1
}
