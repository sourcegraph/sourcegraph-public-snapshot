package luar

import (
	"reflect"

	"github.com/yuin/gopher-lua"
)

func structIndex(L *lua.LState) int {
	ref, mt := check(L, 1)
	key := L.CheckString(2)

	if fn := mt.method(key); fn != nil {
		L.Push(fn)
		return 1
	}

	ref = reflect.Indirect(ref)
	index := mt.fieldIndex(key)
	if index == nil {
		return 0
	}
	field := ref.FieldByIndex(index)
	if !field.CanInterface() {
		L.RaiseError("cannot interface field " + key)
	}

	if (field.Kind() == reflect.Struct || field.Kind() == reflect.Array) && field.CanAddr() {
		field = field.Addr()
	}
	L.Push(New(L, field.Interface()))
	return 1
}

func structPtrIndex(L *lua.LState) int {
	ref, mt := check(L, 1)
	key := L.CheckString(2)

	if fn := mt.method(key); fn != nil {
		L.Push(fn)
		return 1
	}

	ref = ref.Elem()
	mt = MT(L, ref.Interface())
	if fn := mt.method(key); fn != nil {
		L.Push(fn)
		return 1
	}

	index := mt.fieldIndex(key)
	if index == nil {
		return 0
	}
	field := ref.FieldByIndex(index)
	if !field.CanInterface() {
		L.RaiseError("cannot interface field " + key)
	}

	if (field.Kind() == reflect.Struct || field.Kind() == reflect.Array) && field.CanAddr() {
		field = field.Addr()
	}
	L.Push(New(L, field.Interface()))
	return 1
}

func structPtrNewIndex(L *lua.LState) int {
	ref, mt := check(L, 1)
	key := L.CheckString(2)
	value := L.CheckAny(3)

	ref = ref.Elem()
	mt = MT(L, ref.Interface())

	index := mt.fieldIndex(key)
	if index == nil {
		L.RaiseError("unknown field " + key)
	}
	field := ref.FieldByIndex(index)
	if !field.CanSet() {
		L.RaiseError("cannot set field " + key)
	}
	val, err := lValueToReflect(L, value, field.Type(), nil)
	if err != nil {
		L.ArgError(2, err.Error())
	}
	field.Set(val)
	return 0
}

func structEq(L *lua.LState) int {
	ref1, _ := check(L, 1)
	ref2, _ := check(L, 2)

	L.Push(lua.LBool(ref1.Interface() == ref2.Interface()))
	return 1
}
