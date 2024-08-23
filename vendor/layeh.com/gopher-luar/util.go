package luar

import (
	"fmt"
	"reflect"
	"unicode"
	"unicode/utf8"

	"github.com/yuin/gopher-lua"
)

func check(L *lua.LState, idx int) (ref reflect.Value, mt *Metatable) {
	ud := L.CheckUserData(idx)
	ref = reflect.ValueOf(ud.Value)
	mt = &Metatable{LTable: ud.Metatable.(*lua.LTable)}
	return
}

func tostring(L *lua.LState) int {
	ud := L.CheckUserData(1)
	if stringer, ok := ud.Value.(fmt.Stringer); ok {
		L.Push(lua.LString(stringer.String()))
	} else {
		L.Push(lua.LString(ud.String()))
	}
	return 1
}

func getUnexportedName(name string) string {
	first, n := utf8.DecodeRuneInString(name)
	if n == 0 {
		return name
	}
	return string(unicode.ToLower(first)) + name[n:]
}
