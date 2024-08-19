package luar

import (
	"reflect"

	"github.com/yuin/gopher-lua"
)

func checkPtr(L *lua.LState, idx int) (ref reflect.Value, mt *Metatable) {
	ud := L.CheckUserData(idx)
	ref = reflect.ValueOf(ud.Value)
	if expecting := reflect.Ptr; ref.Kind() != expecting {
		L.ArgError(idx, "expecting "+expecting.String())
	}
	mt = &Metatable{LTable: ud.Metatable.(*lua.LTable)}
	return
}

func ptrIndex(L *lua.LState) int {
	ref, mt := checkPtr(L, 1)
	key := L.CheckString(2)

	if fn := mt.method(key); fn != nil {
		L.Push(fn)
		return 1
	}

	// fallback to non-pointer method
	ref = ref.Elem()
	mt = MT(L, ref.Interface())
	if fn := mt.method(key); fn != nil {
		L.Push(fn)
		return 1
	}

	return 0
}

func ptrPow(L *lua.LState) int {
	ref, _ := checkPtr(L, 1)
	val := L.CheckAny(2)

	elem := ref.Elem()
	if !elem.CanSet() {
		L.RaiseError("unable to set pointer value")
	}
	value, err := lValueToReflect(L, val, elem.Type(), nil)
	if err != nil {
		L.ArgError(2, err.Error())
	}
	elem.Set(value)
	L.SetTop(1)
	return 1
}

func ptrUnm(L *lua.LState) int {
	ref, _ := checkPtr(L, 1)
	elem := ref.Elem()
	if !elem.CanInterface() {
		L.RaiseError("cannot interface pointer type " + elem.String())
	}
	L.Push(New(L, elem.Interface()))
	return 1
}

func ptrEq(L *lua.LState) int {
	ref1, _ := checkPtr(L, 1)
	ref2, _ := checkPtr(L, 2)

	L.Push(lua.LBool(ref1.Pointer() == ref2.Pointer()))
	return 1
}
