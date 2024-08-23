package luar

import (
	"reflect"

	"github.com/yuin/gopher-lua"
)

func addMethods(L *lua.LState, c *Config, vtype reflect.Type, tbl *lua.LTable, ptrReceiver bool) {
	for i := 0; i < vtype.NumMethod(); i++ {
		method := vtype.Method(i)
		if method.PkgPath != "" {
			continue
		}
		namesFn := c.MethodNames
		if namesFn == nil {
			namesFn = defaultMethodNames
		}
		fn := funcWrapper(L, method.Func, ptrReceiver)
		for _, name := range namesFn(vtype, method) {
			tbl.RawSetString(name, fn)
		}
	}
}

func collectFields(vtype reflect.Type, current []int) map[string]reflect.StructField {
	m := make(map[string]reflect.StructField)

	var subFields []map[string]reflect.StructField

	for i, n := 0, vtype.NumField(); i < n; i++ {
		field := vtype.Field(i)

		if field.PkgPath == "" {
			field.Index = append(current[:len(current):len(current)], i)
			m[field.Name] = field
		}

		if field.Anonymous {
			t := field.Type
			if t.Kind() != reflect.Struct {
				if t.Kind() != reflect.Ptr || t.Elem().Kind() != reflect.Struct {
					continue
				}
				t = field.Type.Elem()
			}
			r := collectFields(t, append(current[:len(current):len(current)], i))
			subFields = append(subFields, r)
		}
	}

	m2 := make(map[string]reflect.StructField)
	for i := 0; i < len(subFields); i++ {
		for name, value := range subFields[i] {
			if _, ok := m2[name]; !ok {
				m2[name] = value
			} else {
				m2[name] = reflect.StructField{}
			}
		}
	}

	for name, value := range m2 {
		if len(value.Index) > 0 {
			if _, ok := m[name]; !ok {
				m[name] = value
			}
		}
	}

	return m
}

func addFields(L *lua.LState, c *Config, vtype reflect.Type, tbl *lua.LTable) {
	namesFn := c.FieldNames
	if namesFn == nil {
		namesFn = defaultFieldNames
	}

	for _, field := range collectFields(vtype, nil) {
		aliases := namesFn(vtype, field)
		if len(aliases) > 0 {
			ud := L.NewUserData()
			ud.Value = field.Index
			for _, alias := range aliases {
				tbl.RawSetString(alias, ud)
			}
		}
	}
}

func getMetatable(L *lua.LState, vtype reflect.Type) *lua.LTable {
	config := GetConfig(L)

	if v := config.regular[vtype]; v != nil {
		return v
	}

	var (
		mt      *lua.LTable
		methods = L.CreateTable(0, vtype.NumMethod())
	)

	switch vtype.Kind() {
	case reflect.Array:
		mt = L.CreateTable(0, 7)

		mt.RawSetString("__index", L.NewFunction(arrayIndex))
		mt.RawSetString("__len", L.NewFunction(arrayLen))
		mt.RawSetString("__call", L.NewFunction(arrayCall))
		mt.RawSetString("__eq", L.NewFunction(arrayEq))

		addMethods(L, config, vtype, methods, false)
	case reflect.Chan:
		mt = L.CreateTable(0, 8)

		mt.RawSetString("__index", L.NewFunction(chanIndex))
		mt.RawSetString("__len", L.NewFunction(chanLen))
		mt.RawSetString("__eq", L.NewFunction(chanEq))
		mt.RawSetString("__call", L.NewFunction(chanCall))
		mt.RawSetString("__unm", L.NewFunction(chanUnm))

		addMethods(L, config, vtype, methods, false)
	case reflect.Map:
		mt = L.CreateTable(0, 7)

		mt.RawSetString("__index", L.NewFunction(mapIndex))
		mt.RawSetString("__newindex", L.NewFunction(mapNewIndex))
		mt.RawSetString("__len", L.NewFunction(mapLen))
		mt.RawSetString("__call", L.NewFunction(mapCall))

		addMethods(L, config, vtype, methods, false)
	case reflect.Slice:
		mt = L.CreateTable(0, 8)

		mt.RawSetString("__index", L.NewFunction(sliceIndex))
		mt.RawSetString("__newindex", L.NewFunction(sliceNewIndex))
		mt.RawSetString("__len", L.NewFunction(sliceLen))
		mt.RawSetString("__call", L.NewFunction(sliceCall))
		mt.RawSetString("__add", L.NewFunction(sliceAdd))

		addMethods(L, config, vtype, methods, false)
	case reflect.Struct:
		mt = L.CreateTable(0, 6)

		fields := L.CreateTable(0, vtype.NumField())
		addFields(L, config, vtype, fields)
		mt.RawSetString("fields", fields)

		mt.RawSetString("__index", L.NewFunction(structIndex))
		mt.RawSetString("__eq", L.NewFunction(structEq))

		addMethods(L, config, vtype, methods, false)
	case reflect.Ptr:
		switch vtype.Elem().Kind() {
		case reflect.Array:
			mt = L.CreateTable(0, 10)

			mt.RawSetString("__index", L.NewFunction(arrayPtrIndex))
			mt.RawSetString("__newindex", L.NewFunction(arrayPtrNewIndex))
			mt.RawSetString("__call", L.NewFunction(arrayCall)) // same as non-pointer
			mt.RawSetString("__len", L.NewFunction(arrayLen))   // same as non-pointer
		case reflect.Struct:
			mt = L.CreateTable(0, 8)

			mt.RawSetString("__index", L.NewFunction(structPtrIndex))
			mt.RawSetString("__newindex", L.NewFunction(structPtrNewIndex))
		default:
			mt = L.CreateTable(0, 7)

			mt.RawSetString("__index", L.NewFunction(ptrIndex))
		}

		mt.RawSetString("__eq", L.NewFunction(ptrEq))
		mt.RawSetString("__pow", L.NewFunction(ptrPow))
		mt.RawSetString("__unm", L.NewFunction(ptrUnm))

		addMethods(L, config, vtype, methods, true)
	default:
		panic("unexpected kind " + vtype.Kind().String())
	}

	mt.RawSetString("__tostring", L.NewFunction(tostring))
	mt.RawSetString("__metatable", lua.LString("gopher-luar"))
	mt.RawSetString("methods", methods)

	config.regular[vtype] = mt
	return mt
}

func getTypeMetatable(L *lua.LState, t reflect.Type) *lua.LTable {
	config := GetConfig(L)

	if v := config.types; v != nil {
		return v
	}

	mt := L.CreateTable(0, 3)
	mt.RawSetString("__call", L.NewFunction(typeCall))
	mt.RawSetString("__eq", L.NewFunction(typeEq))
	mt.RawSetString("__metatable", lua.LString("gopher-luar"))

	config.types = mt
	return mt
}
