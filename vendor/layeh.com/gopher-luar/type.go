package luar

import (
	"reflect"

	"github.com/yuin/gopher-lua"
)

func checkType(L *lua.LState, idx int) reflect.Type {
	ud := L.CheckUserData(idx)
	return ud.Value.(reflect.Type)
}

func typeCall(L *lua.LState) int {
	ref := checkType(L, 1)

	var value reflect.Value
	switch ref.Kind() {
	case reflect.Chan:
		buffer := L.OptInt(2, 0)

		if buffer < 0 {
			L.ArgError(2, "negative buffer size")
		}
		if ref.ChanDir() != reflect.BothDir {
			L.RaiseError("unidirectional channel type")
		}

		value = reflect.MakeChan(ref, buffer)
	case reflect.Map:
		value = reflect.MakeMap(ref)
	case reflect.Slice:
		length := L.OptInt(2, 0)
		capacity := L.OptInt(3, length)

		if length < 0 {
			L.ArgError(2, "negative length")
		}
		if capacity < 0 {
			L.ArgError(3, "negative capacity")
		}
		if length > capacity {
			L.RaiseError("length > capacity")
		}

		value = reflect.MakeSlice(ref, length, capacity)
	default:
		value = reflect.New(ref)
	}
	L.Push(New(L, value.Interface()))
	return 1
}

func typeEq(L *lua.LState) int {
	type1 := checkType(L, 1)
	type2 := checkType(L, 2)
	L.Push(lua.LBool(type1 == type2))
	return 1
}
