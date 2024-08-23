package luar

import (
	"fmt"
	"reflect"

	"github.com/yuin/gopher-lua"
)

// New creates and returns a new lua.LValue for the given value. Values are
// converted in the following manner:
//
// A nil value (untyped, or a nil channel, function, map, pointer, or slice) is
// converted to lua.LNil.
//
// A lua.LValue value is returned without conversion.
//
// Boolean values are converted to lua.LBool.
//
// String values are converted to lua.LString.
//
// Real numeric values (ints, uints, and floats) are converted to lua.LNumber.
//
// Functions are converted to *lua.LFunction. When called from Lua, Lua values
// are converted to Go using the rules described in the package documentation,
// and Go return values converted to Lua values using the rules described by
// New.
//
// If a function has the signature:
//  func(*LState) int // *LState defined in this package, not in lua
// The argument and return value conversions described above are skipped, and
// the function is called with the arguments passed on the Lua stack. Return
// values are pushed to the stack and the number of return values is returned
// from the function.
//
// Arrays, channels, maps, pointers, slices, and structs are all converted to
// *lua.LUserData with its Value field set to value. The userdata's metatable
// is set to a table generated for value's type. The type's method set is
// callable from the Lua type. If the type implements the fmt.Stringer
// interface, that method will be used when the value is passed to the Lua
// tostring function.
//
// With arrays, the # operator returns the array's length. Array elements can
// be accessed with the index operator (array[index]). Calling an array
// (array()) returns an iterator over the array that can be used in a for loop.
// Two arrays of the same type can be compared for equality. Additionally, a
// pointer to an array allows the array elements to be modified
// (array[index] = value).
//
// With channels, the # operator returns the number of elements buffered in the
// channel. Two channels of the same type can be compared for equality (i.e. if
// they were created with the same make call). Calling a channel value with
// no arguments reads one element from the channel, returning the value and a
// boolean indicating if the channel is closed. Calling a channel value with
// one argument sends the argument to the channel. The channel's unary minus
// operator closes the channel (_ = -channel).
//
// With maps, the # operator returns the number of elements in the map. Map
// elements can be accessed using the index operator (map[key]) and also set
// (map[key] = value). Calling a map value returns an iterator over the map that
// can be used in a for loop. If a map's key type is string, map values take
// priority over methods.
//
// With slices, the # operator returns the length of the slice. Slice elements
// can be accessed using the index operator (slice[index]) and also set
// (slice[index] = value). Calling a slice returns an iterator over its elements
// that can be used in a for loop. Elements can be appended to a slice using the
// add operator (new_slice = slice + element).
//
// With structs, fields can be accessed using the index operator
// (struct[field]). As a special case, accessing field that is an array or
// struct field will return a pointer to that value. Structs of the same type
// can be tested for equality. Additionally, a pointer to a struct can have its
// fields set (struct[field] = value).
//
// Struct field accessibility can be changed by setting the field's luar tag.
// If the tag is empty (default), the field is accessed by its name and its
// name with a lowercase first letter (e.g. "Field1" would be accessible using
// "Field1" or "field1"). If the tag is "-", the field will not be accessible.
// Any other tag value makes the field accessible through that name.
//
// Pointer values can be compared for equality. The pointed to value can be
// changed using the pow operator (pointer = pointer ^ value). A pointer can be
// dereferenced using the unary minus operator (value = -pointer).
//
// All other values (complex numbers, unsafepointer, uintptr) are converted to
// *lua.LUserData with its Value field set to value and no custom metatable.
//
func New(L *lua.LState, value interface{}) lua.LValue {
	if value == nil {
		return lua.LNil
	}
	if lval, ok := value.(lua.LValue); ok {
		return lval
	}

	switch val := reflect.ValueOf(value); val.Kind() {
	case reflect.Bool:
		return lua.LBool(val.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return lua.LNumber(float64(val.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return lua.LNumber(float64(val.Uint()))
	case reflect.Float32, reflect.Float64:
		return lua.LNumber(val.Float())
	case reflect.Chan, reflect.Map, reflect.Ptr, reflect.Slice:
		if val.IsNil() {
			return lua.LNil
		}
		fallthrough
	case reflect.Array, reflect.Struct:
		ud := L.NewUserData()
		ud.Value = val.Interface()
		ud.Metatable = getMetatable(L, val.Type())
		return ud
	case reflect.Func:
		if val.IsNil() {
			return lua.LNil
		}
		return funcWrapper(L, val, false)
	case reflect.String:
		return lua.LString(val.String())
	default:
		ud := L.NewUserData()
		ud.Value = val.Interface()
		return ud
	}
}

// NewType returns a new type generator for the given value's type.
//
// When the returned lua.LValue is called, a new value will be created that is
// dependent on value's type:
//
// If value is a channel, the first argument optionally specifies the channel's
// buffer size (defaults to 1). The new channel is returned.
//
// If value is a map, a new map is returned.
//
// If value is a slice, the first argument optionally specifies the slices's
// length (defaults to 0), and the second argument optionally specifies the
// slice's capacity (defaults to the first argument). The new slice is returned.
//
// All other types return a new pointer to the zero value of value's type.
func NewType(L *lua.LState, value interface{}) lua.LValue {
	val := reflect.TypeOf(value)
	ud := L.NewUserData()
	ud.Value = val
	ud.Metatable = getTypeMetatable(L, val)

	return ud
}

type conversionError struct {
	Lua  lua.LValue
	Hint reflect.Type
}

func (c conversionError) Error() string {
	if _, isNil := c.Lua.(*lua.LNilType); isNil {
		return fmt.Sprintf("cannot use nil as type %s", c.Hint)
	}

	var val interface{}

	if userData, ok := c.Lua.(*lua.LUserData); ok {
		val = userData.Value
	} else {
		val = c.Lua
	}

	return fmt.Sprintf("cannot use %v (type %T) as type %s", val, val, c.Hint)
}

type structFieldError struct {
	Field string
	Type  reflect.Type
}

func (s structFieldError) Error() string {
	return `type ` + s.Type.String() + ` has no field ` + s.Field
}

func lValueToReflect(L *lua.LState, v lua.LValue, hint reflect.Type, tryConvertPtr *bool) (reflect.Value, error) {
	visited := make(map[*lua.LTable]reflect.Value)
	return lValueToReflectInner(L, v, hint, visited, tryConvertPtr)
}

func lValueToReflectInner(L *lua.LState, v lua.LValue, hint reflect.Type, visited map[*lua.LTable]reflect.Value, tryConvertPtr *bool) (reflect.Value, error) {
	if hint.Implements(refTypeLuaLValue) {
		return reflect.ValueOf(v), nil
	}

	isPtr := false

	switch converted := v.(type) {
	case lua.LBool:
		val := reflect.ValueOf(bool(converted))
		if !val.Type().ConvertibleTo(hint) {
			return reflect.Value{}, conversionError{
				Lua:  v,
				Hint: hint,
			}
		}
		return val.Convert(hint), nil
	case lua.LChannel:
		val := reflect.ValueOf(converted)
		if !val.Type().ConvertibleTo(hint) {
			return reflect.Value{}, conversionError{
				Lua:  v,
				Hint: hint,
			}
		}
		return val.Convert(hint), nil
	case lua.LNumber:
		val := reflect.ValueOf(float64(converted))
		if !val.Type().ConvertibleTo(hint) {
			return reflect.Value{}, conversionError{
				Lua:  v,
				Hint: hint,
			}
		}
		return val.Convert(hint), nil
	case *lua.LFunction:
		emptyIfaceHint := false
		switch {
		case hint == refTypeEmptyIface:
			emptyIfaceHint = true
			inOut := []reflect.Type{
				reflect.SliceOf(refTypeEmptyIface),
			}
			hint = reflect.FuncOf(inOut, inOut, true)
		case hint.Kind() != reflect.Func:
			return reflect.Value{}, conversionError{
				Lua:  v,
				Hint: hint,
			}
		}

		fn := func(args []reflect.Value) []reflect.Value {
			thread, cancelFunc := L.NewThread()
			defer thread.Close()
			if cancelFunc != nil {
				defer cancelFunc()
			}
			thread.Push(converted)
			defer thread.SetTop(0)

			argCount := 0
			for i, arg := range args {
				if i+1 == len(args) && hint.IsVariadic() {
					// arg is a varadic slice
					for j := 0; j < arg.Len(); j++ {
						arg := arg.Index(j)
						thread.Push(New(thread, arg.Interface()))
						argCount++
					}
					break
				}

				thread.Push(New(thread, arg.Interface()))
				argCount++
			}

			thread.Call(argCount, lua.MultRet)
			top := thread.GetTop()

			switch {
			case emptyIfaceHint:
				ret := reflect.MakeSlice(reflect.SliceOf(refTypeEmptyIface), top, top)

				for i := 1; i <= top; i++ {
					item, err := lValueToReflect(thread, thread.Get(i), refTypeEmptyIface, nil)
					if err != nil {
						panic(err)
					}
					ret.Index(i - 1).Set(item)
				}

				return []reflect.Value{ret}

			case top == hint.NumOut():
				ret := make([]reflect.Value, top)

				var err error
				for i := 1; i <= top; i++ {
					outHint := hint.Out(i - 1)
					item := thread.Get(i)
					ret[i-1], err = lValueToReflect(thread, item, outHint, nil)
					if err != nil {
						panic(err)
					}
				}

				return ret
			}

			panic(fmt.Errorf("expecting %d return values, got %d", hint.NumOut(), top))
		}
		return reflect.MakeFunc(hint, fn), nil
	case *lua.LNilType:
		switch hint.Kind() {
		case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice, reflect.UnsafePointer, reflect.Uintptr:
			return reflect.Zero(hint), nil
		}

		return reflect.Value{}, conversionError{
			Lua:  v,
			Hint: hint,
		}

	case *lua.LState:
		val := reflect.ValueOf(converted)
		if !val.Type().ConvertibleTo(hint) {
			return reflect.Value{}, conversionError{
				Lua:  v,
				Hint: hint,
			}
		}
		return val.Convert(hint), nil

	case lua.LString:
		val := reflect.ValueOf(string(converted))
		if !val.Type().ConvertibleTo(hint) {
			return reflect.Value{}, conversionError{
				Lua:  v,
				Hint: hint,
			}
		}
		return val.Convert(hint), nil

	case *lua.LTable:
		if existing := visited[converted]; existing.IsValid() {
			return existing, nil
		}

		if hint == refTypeEmptyIface {
			hint = reflect.MapOf(refTypeEmptyIface, refTypeEmptyIface)
		}

		switch {
		case hint.Kind() == reflect.Array:
			elemType := hint.Elem()
			length := converted.Len()
			if length != hint.Len() {
				return reflect.Value{}, conversionError{
					Lua:  v,
					Hint: hint,
				}
			}
			s := reflect.New(hint).Elem()
			visited[converted] = s

			for i := 0; i < length; i++ {
				value := converted.RawGetInt(i + 1)
				elemValue, err := lValueToReflectInner(L, value, elemType, visited, nil)
				if err != nil {
					return reflect.Value{}, err
				}
				s.Index(i).Set(elemValue)
			}

			return s, nil

		case hint.Kind() == reflect.Slice:
			elemType := hint.Elem()
			length := converted.Len()
			s := reflect.MakeSlice(hint, length, length)
			visited[converted] = s

			for i := 0; i < length; i++ {
				value := converted.RawGetInt(i + 1)
				elemValue, err := lValueToReflectInner(L, value, elemType, visited, nil)
				if err != nil {
					return reflect.Value{}, err
				}
				s.Index(i).Set(elemValue)
			}

			return s, nil

		case hint.Kind() == reflect.Map:
			keyType := hint.Key()
			elemType := hint.Elem()
			s := reflect.MakeMap(hint)
			visited[converted] = s

			for key := lua.LNil; ; {
				var value lua.LValue
				key, value = converted.Next(key)
				if key == lua.LNil {
					break
				}

				lKey, err := lValueToReflectInner(L, key, keyType, visited, nil)
				if err != nil {
					return reflect.Value{}, err
				}
				lValue, err := lValueToReflectInner(L, value, elemType, visited, nil)
				if err != nil {
					return reflect.Value{}, err
				}
				s.SetMapIndex(lKey, lValue)
			}

			return s, nil

		case hint.Kind() == reflect.Ptr && hint.Elem().Kind() == reflect.Struct:
			hint = hint.Elem()
			isPtr = true
			fallthrough
		case hint.Kind() == reflect.Struct:
			s := reflect.New(hint)
			visited[converted] = s

			t := s.Elem()

			mt := &Metatable{
				LTable: getMetatable(L, hint),
			}

			for key := lua.LNil; ; {
				var value lua.LValue
				key, value = converted.Next(key)
				if key == lua.LNil {
					break
				}
				if _, ok := key.(lua.LString); !ok {
					continue
				}

				fieldName := key.String()
				index := mt.fieldIndex(fieldName)
				if index == nil {
					return reflect.Value{}, structFieldError{
						Type:  hint,
						Field: fieldName,
					}
				}
				field := hint.FieldByIndex(index)

				lValue, err := lValueToReflectInner(L, value, field.Type, visited, nil)
				if err != nil {
					return reflect.Value{}, nil
				}
				t.FieldByIndex(field.Index).Set(lValue)
			}

			if isPtr {
				return s, nil
			}

			return t, nil
		}

		return reflect.Value{}, conversionError{
			Lua:  v,
			Hint: hint,
		}

	case *lua.LUserData:
		val := reflect.ValueOf(converted.Value)
		if tryConvertPtr != nil && val.Kind() != reflect.Ptr && hint.Kind() == reflect.Ptr && val.Type() == hint.Elem() {
			newVal := reflect.New(hint.Elem())
			newVal.Elem().Set(val)
			val = newVal
			*tryConvertPtr = true
		} else {
			if !val.Type().ConvertibleTo(hint) {
				return reflect.Value{}, conversionError{
					Lua:  converted,
					Hint: hint,
				}
			}
			val = val.Convert(hint)
			if tryConvertPtr != nil {
				*tryConvertPtr = false
			}
		}
		return val, nil
	}

	panic("never reaches")
}
