//+build !go1.12

package luar

import "github.com/yuin/gopher-lua"

func mapCall(L *lua.LState) int {
	ref, _ := check(L, 1)

	keys := ref.MapKeys()
	i := 0
	fn := func(L *lua.LState) int {
		if i >= len(keys) {
			return 0
		}
		L.Push(New(L, keys[i].Interface()))
		L.Push(New(L, ref.MapIndex(keys[i]).Interface()))
		i++
		return 2
	}
	L.Push(L.NewFunction(fn))
	return 1
}
