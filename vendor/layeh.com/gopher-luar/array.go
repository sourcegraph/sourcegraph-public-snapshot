package luar

import (
	"reflect"

	"github.com/yuin/gopher-lua"
)

func arrayIndex(L *lua.LState) int {
	ref, mt := check(L, 1)
	ref = reflect.Indirect(ref)
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

func arrayPtrIndex(L *lua.LState) int {
	ref, mt := check(L, 1)
	ref = ref.Elem()
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

		mt = MT(L, ref.Interface())
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

func arrayPtrNewIndex(L *lua.LState) int {
	ref, _ := check(L, 1)
	ref = ref.Elem()

	index := L.CheckInt(2)
	value := L.CheckAny(3)
	if index < 1 || index > ref.Len() {
		L.ArgError(2, "index out of range")
	}
	hint := ref.Type().Elem()
	val, err := lValueToReflect(L, value, hint, nil)
	if err != nil {
		L.ArgError(3, err.Error())
	}
	ref.Index(index - 1).Set(val)
	return 0
}

func arrayLen(L *lua.LState) int {
	ref, _ := check(L, 1)
	ref = reflect.Indirect(ref)

	L.Push(lua.LNumber(ref.Len()))
	return 1
}

func arrayCall(L *lua.LState) int {
	ref, _ := check(L, 1)
	ref = reflect.Indirect(ref)

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

func arrayEq(L *lua.LState) int {
	ref1, _ := check(L, 1)
	ref2, _ := check(L, 2)

	L.Push(lua.LBool(ref1.Interface() == ref2.Interface()))
	return 1
}
