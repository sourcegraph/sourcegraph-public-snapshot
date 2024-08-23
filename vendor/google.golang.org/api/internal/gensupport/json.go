// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gensupport

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// MarshalJSON returns a JSON encoding of schema containing only selected fields.
// A field is selected if any of the following is true:
//   - it has a non-empty value
//   - its field name is present in forceSendFields and it is not a nil pointer or nil interface
//   - its field name is present in nullFields.
//
// The JSON key for each selected field is taken from the field's json: struct tag.
func MarshalJSON(schema interface{}, forceSendFields, nullFields []string) ([]byte, error) {
	if len(forceSendFields) == 0 && len(nullFields) == 0 {
		return json.Marshal(schema)
	}

	mustInclude := make(map[string]bool)
	for _, f := range forceSendFields {
		mustInclude[f] = true
	}
	useNull := make(map[string]bool)
	useNullMaps := make(map[string]map[string]bool)
	for _, nf := range nullFields {
		parts := strings.SplitN(nf, ".", 2)
		field := parts[0]
		if len(parts) == 1 {
			useNull[field] = true
		} else {
			if useNullMaps[field] == nil {
				useNullMaps[field] = map[string]bool{}
			}
			useNullMaps[field][parts[1]] = true
		}
	}

	dataMap, err := schemaToMap(schema, mustInclude, useNull, useNullMaps)
	if err != nil {
		return nil, err
	}
	return json.Marshal(dataMap)
}

func schemaToMap(schema interface{}, mustInclude, useNull map[string]bool, useNullMaps map[string]map[string]bool) (map[string]interface{}, error) {
	m := make(map[string]interface{})
	s := reflect.ValueOf(schema)
	st := s.Type()

	for i := 0; i < s.NumField(); i++ {
		jsonTag := st.Field(i).Tag.Get("json")
		if jsonTag == "" {
			continue
		}
		tag, err := parseJSONTag(jsonTag)
		if err != nil {
			return nil, err
		}
		if tag.ignore {
			continue
		}

		v := s.Field(i)
		f := st.Field(i)

		if useNull[f.Name] {
			if !isEmptyValue(v) {
				return nil, fmt.Errorf("field %q in NullFields has non-empty value", f.Name)
			}
			m[tag.apiName] = nil
			continue
		}

		if !includeField(v, f, mustInclude) {
			continue
		}

		// If map fields are explicitly set to null, use a map[string]interface{}.
		if f.Type.Kind() == reflect.Map && useNullMaps[f.Name] != nil {
			ms, ok := v.Interface().(map[string]string)
			if !ok {
				mi, err := initMapSlow(v, f.Name, useNullMaps)
				if err != nil {
					return nil, err
				}
				m[tag.apiName] = mi
				continue
			}
			mi := map[string]interface{}{}
			for k, v := range ms {
				mi[k] = v
			}
			for k := range useNullMaps[f.Name] {
				mi[k] = nil
			}
			m[tag.apiName] = mi
			continue
		}

		// nil maps are treated as empty maps.
		if f.Type.Kind() == reflect.Map && v.IsNil() {
			m[tag.apiName] = map[string]string{}
			continue
		}

		// nil slices are treated as empty slices.
		if f.Type.Kind() == reflect.Slice && v.IsNil() {
			m[tag.apiName] = []bool{}
			continue
		}

		if tag.stringFormat {
			m[tag.apiName] = formatAsString(v, f.Type.Kind())
		} else {
			m[tag.apiName] = v.Interface()
		}
	}
	return m, nil
}

// initMapSlow uses reflection to build up a map object. This is slower than
// the default behavior so it should be used only as a fallback.
func initMapSlow(rv reflect.Value, fieldName string, useNullMaps map[string]map[string]bool) (map[string]interface{}, error) {
	mi := map[string]interface{}{}
	iter := rv.MapRange()
	for iter.Next() {
		k, ok := iter.Key().Interface().(string)
		if !ok {
			return nil, fmt.Errorf("field %q has keys in NullFields but is not a map[string]any", fieldName)
		}
		v := iter.Value().Interface()
		mi[k] = v
	}
	for k := range useNullMaps[fieldName] {
		mi[k] = nil
	}
	return mi, nil
}

// formatAsString returns a string representation of v, dereferencing it first if possible.
func formatAsString(v reflect.Value, kind reflect.Kind) string {
	if kind == reflect.Ptr && !v.IsNil() {
		v = v.Elem()
	}

	return fmt.Sprintf("%v", v.Interface())
}

// jsonTag represents a restricted version of the struct tag format used by encoding/json.
// It is used to describe the JSON encoding of fields in a Schema struct.
type jsonTag struct {
	apiName      string
	stringFormat bool
	ignore       bool
}

// parseJSONTag parses a restricted version of the struct tag format used by encoding/json.
// The format of the tag must match that generated by the Schema.writeSchemaStruct method
// in the api generator.
func parseJSONTag(val string) (jsonTag, error) {
	if val == "-" {
		return jsonTag{ignore: true}, nil
	}

	var tag jsonTag

	i := strings.Index(val, ",")
	if i == -1 || val[:i] == "" {
		return tag, fmt.Errorf("malformed json tag: %s", val)
	}

	tag = jsonTag{
		apiName: val[:i],
	}

	switch val[i+1:] {
	case "omitempty":
	case "omitempty,string":
		tag.stringFormat = true
	default:
		return tag, fmt.Errorf("malformed json tag: %s", val)
	}

	return tag, nil
}

// Reports whether the struct field "f" with value "v" should be included in JSON output.
func includeField(v reflect.Value, f reflect.StructField, mustInclude map[string]bool) bool {
	// The regular JSON encoding of a nil pointer is "null", which means "delete this field".
	// Therefore, we could enable field deletion by honoring pointer fields' presence in the mustInclude set.
	// However, many fields are not pointers, so there would be no way to delete these fields.
	// Rather than partially supporting field deletion, we ignore mustInclude for nil pointer fields.
	// Deletion will be handled by a separate mechanism.
	if f.Type.Kind() == reflect.Ptr && v.IsNil() {
		return false
	}

	// The "any" type is represented as an interface{}.  If this interface
	// is nil, there is no reasonable representation to send.  We ignore
	// these fields, for the same reasons as given above for pointers.
	if f.Type.Kind() == reflect.Interface && v.IsNil() {
		return false
	}

	return mustInclude[f.Name] || !isEmptyValue(v)
}

// isEmptyValue reports whether v is the empty value for its type.  This
// implementation is based on that of the encoding/json package, but its
// correctness does not depend on it being identical. What's important is that
// this function return false in situations where v should not be sent as part
// of a PATCH operation.
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}
