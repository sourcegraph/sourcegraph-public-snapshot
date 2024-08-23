package valast

import "reflect"

// isAddressableKind reports if v would be encoded as a Go literal which is addressable or not.
// For example, &struct{}{}, &map[string]string{}, &[]string{} are all addressable - but &"string",
// &5, &1.345, &myBool(true) are not.
func isAddressableKind(v reflect.Kind) bool {
	return v != reflect.Bool &&
		v != reflect.Int &&
		v != reflect.Int8 &&
		v != reflect.Int16 &&
		v != reflect.Int32 &&
		v != reflect.Int64 &&
		v != reflect.Uint &&
		v != reflect.Uint8 &&
		v != reflect.Uint16 &&
		v != reflect.Uint32 &&
		v != reflect.Uint64 &&
		v != reflect.Uintptr &&
		v != reflect.Float32 &&
		v != reflect.Float64 &&
		v != reflect.Complex64 &&
		v != reflect.Complex128 &&
		v != reflect.Ptr &&
		v != reflect.String &&
		v != reflect.UnsafePointer
}

// valueLess tells if i is less than j, according to normal Go less-than < operator rules. Values
// that are unsortable according to Go rules will always yield true.
//
// The two values must be of the same kind or a panic will occur.
func valueLess(i, j reflect.Value) bool {
	ii := unexported(i)
	switch ii.Kind() {
	case reflect.Bool:
		x := 0
		if ii.Bool() {
			x = 1
		}
		y := 0
		if unexported(j).Bool() {
			y = 1
		}
		return x < y
	case reflect.Int:
		return ii.Int() < unexported(j).Int()
	case reflect.Int8:
		return ii.Int() < unexported(j).Int()
	case reflect.Int16:
		return ii.Int() < unexported(j).Int()
	case reflect.Int32:
		return ii.Int() < unexported(j).Int()
	case reflect.Int64:
		return ii.Int() < unexported(j).Int()
	case reflect.Uint:
		return ii.Uint() < unexported(j).Uint()
	case reflect.Uint8:
		return ii.Uint() < unexported(j).Uint()
	case reflect.Uint16:
		return ii.Uint() < unexported(j).Uint()
	case reflect.Uint32:
		return ii.Uint() < unexported(j).Uint()
	case reflect.Uint64:
		return ii.Uint() < unexported(j).Uint()
	case reflect.Uintptr:
		return ii.Uint() < unexported(j).Uint()
	case reflect.Float32:
		return ii.Float() < unexported(j).Float()
	case reflect.Float64:
		return ii.Float() < unexported(j).Float()
	case reflect.Ptr:
		return ii.Pointer() < unexported(j).Pointer()
	case reflect.String:
		return ii.String() < unexported(j).String()
	case reflect.UnsafePointer:
		return ii.Pointer() < unexported(j).Pointer()
	case reflect.Complex64:
		return true
	case reflect.Complex128:
		return true
	case reflect.Array:
		return true
	case reflect.Map:
		return true
	case reflect.Interface:
		return true
	case reflect.Slice:
		return true
	case reflect.Struct:
		return true
	default:
		// never here
		return true
	}
}
