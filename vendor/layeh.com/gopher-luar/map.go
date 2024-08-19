package luar

import (
	"reflect"

	"github.com/yuin/gopher-lua"
)

func mapIndex(L *lua.LState) int {
	ref, mt := check(L, 1)
	key := L.CheckAny(2)

	convertedKey, err := lValueToReflect(L, key, ref.Type().Key(), nil)
	if err == nil {
		item := ref.MapIndex(convertedKey)
		if item.IsValid() {
			L.Push(New(L, item.Interface()))
			return 1
		}
	}

	if lstring, ok := key.(lua.LString); ok {
		if fn := mt.method(string(lstring)); fn != nil {
			L.Push(fn)
			return 1
		}
	}

	return 0
}

func mapNewIndex(L *lua.LState) int {
	ref, _ := check(L, 1)
	key := L.CheckAny(2)
	value := L.CheckAny(3)

	keyHint := ref.Type().Key()
	convertedKey, err := lValueToReflect(L, key, keyHint, nil)
	if err != nil {
		L.ArgError(2, err.Error())
	}
	var convertedValue reflect.Value
	if value != lua.LNil {
		convertedValue, err = lValueToReflect(L, value, ref.Type().Elem(), nil)
		if err != nil {
			L.ArgError(3, err.Error())
		}
	}
	ref.SetMapIndex(convertedKey, convertedValue)
	return 0
}

func mapLen(L *lua.LState) int {
	ref, _ := check(L, 1)

	L.Push(lua.LNumber(ref.Len()))
	return 1
}
