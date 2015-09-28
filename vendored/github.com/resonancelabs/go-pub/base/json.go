// Sanitize an object for encoding into JSON (e.g. thrift data types)

package base

import (
	"fmt"
	"reflect"

	"github.com/golang/glog"
)

// A reflection-based, generic function for converting an arbitrary Go object into
// a JSON-encodable interface.  This function exists, for example, to avoid writing
// boilerplate conversion functions every thrift type that needs to be converted to
// for the JSON-based APIs.
//
// The two specific needs this addresses are: (1) structs are automatically remapped
// to map[string]interface{} with lowercase, underscore casing since json tags cannot
// be added to the thrift types and (2) all maps are converted to map[string]interface{}
// since Go uses a strict JSON encoder that only accepts string keys -- and there's no
// way to extend thrift types to implement the MarshalJSON interface.
//
//
// whitelist
//      A map of the form "packageName.StructName.FieldName":true of fields to
//      include in the sanitized data.  Any field not in the whitelist is removed.
//      If whitelist is nil, all fields will be included.
//
func SanitizeForJSON(p interface{}, whitelist map[string]bool) interface{} {
	value := reflect.ValueOf(p)

	switch value.Kind() {

	case reflect.Invalid:
		glog.Warningf("Unexpected reflect.Invalid in data structure")
		return nil

	case reflect.Chan:
		glog.Warningf("Unexpected reflect.Chan in data structure")
		return nil

	case reflect.Func:
		glog.Warningf("Unexpected reflect.Func in data structure")
		return nil

	case reflect.Ptr:
		if value.IsNil() {
			return nil
		}
		return SanitizeForJSON(value.Elem().Interface(), whitelist)

	case reflect.Slice:
		// Rebuild the slice, recursively processing the elements in the slice
		count := value.Len()
		newSlice := make([]interface{}, count)
		for i := 0; i < count; i++ {
			sliceValue := value.Index(i)
			if sliceValue.CanInterface() {
				newSlice[i] = SanitizeForJSON(sliceValue.Interface(), whitelist)
			}
		}
		return newSlice

	case reflect.Map:
		// Rebuild the map, forcing the keys to strings for JSON encoding
		// convenience
		stringMap := map[string]interface{}{}
		keys := value.MapKeys()
		for _, k := range keys {
			mapValue := value.MapIndex(k)
			if k.CanInterface() && mapValue.CanInterface() {
				stringKey := fmt.Sprint(k.Interface())
				stringMap[stringKey] = SanitizeForJSON(mapValue.Interface(), whitelist)
			}
		}
		return stringMap

	case reflect.Struct:
		// Rebuild the struct as a map, using lowercase underscore for the
		// names.
		valueType := value.Type()

		count := value.NumField()
		m := make(map[string]interface{})
		for i := 0; i < count; i++ {
			fieldType := valueType.Field(i)
			fieldValue := value.Field(i)

			if fieldValue.Kind() == reflect.Ptr {
				fieldValue = fieldValue.Elem()
			}

			// Exclude fields not in the whitelist
			if whitelist != nil {
				// Note: valueType.String() will return "<packagename>.<StructName>" so
				// the final fieldString will have three parts
				fieldString := fmt.Sprintf("%v.%v", valueType, fieldType.Name)
				if !whitelist[fieldString] {
					continue
				}
			}

			name := ToUnderscoreCase(fieldType.Name)
			if fieldValue.IsValid() && fieldValue.CanInterface() {
				m[name] = SanitizeForJSON(fieldValue.Interface(), whitelist)
			}
		}
		return m

	default:
		return value.Interface()
	}
}
