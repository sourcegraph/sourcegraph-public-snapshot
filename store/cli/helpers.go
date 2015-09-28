package cli

import (
	"fmt"
	"reflect"

	"sourcegraph.com/sourcegraph/go-flags"
)

// Helpers for the CLI config.

func implTypeNames(impls interface{}) []string {
	implsv := reflect.ValueOf(impls)
	names := make([]string, implsv.Len())
	for i := 0; i < implsv.Len(); i++ {
		names[i] = implTypeName(implsv.Index(i))
	}
	return names
}

func implTypeName(impl reflect.Value) string {
	if impl.Kind() == reflect.Interface {
		impl = impl.Elem()
	}
	if impl.Kind() == reflect.Ptr {
		impl = impl.Elem()
	}
	return fmt.Sprintf("%T", impl.Interface())
}

func defaultImplTypeName(impls interface{}) string {
	s := chooseStore("", impls)
	if s == nil {
		return ""
	}
	return implTypeName(reflect.ValueOf(s))
}

// chooseStore returns the element of stores (which should be a slice
// of some store interface type, such as []Users) for which
// implTypeName(x) == typeName.
func chooseStore(typeName string, stores interface{}) interface{} {
	storesv := reflect.ValueOf(stores)
	for i := 0; i < storesv.Len(); i++ {
		s := storesv.Index(i)
		if implTypeName(s) == typeName || /* choose last if none chosen yet */ i == storesv.Len()-1 {
			return s.Interface()
		}
	}
	return nil
}

func setDefaultIfNoneExist(o *flags.Option, defaults []string) {
	if len(o.Default) == 0 || (len(o.Default) == 1 && o.Default[0] == "") {
		o.Default = defaults
	}
}
