package luasandbox

import lua "github.com/yuin/gopher-lua"

// Modified version of the libs available from gopher-lua@linit.go:luaLibs
var builtinLibs = []struct {
	libName string
	libFunc lua.LGFunction
}{
	{lua.BaseLibName, lua.OpenBase},
	{lua.ChannelLibName, lua.OpenChannel},
	{lua.CoroutineLibName, lua.OpenCoroutine},
	{lua.DebugLibName, lua.OpenDebug},
	{lua.LoadLibName, lua.OpenPackage},
	{lua.MathLibName, lua.OpenMath},
	{lua.StringLibName, lua.OpenString},
	{lua.TabLibName, lua.OpenTable},

	// Explicitly omitted
	// {lua.IoLibName, lua.OpenIo},
	// {lua.OsLibName, lua.OpenOs},
}
