package parse

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/sourcegraph/sourcegraph/schema"
)

// diff returns names of the Go fields that have different values between the
// two configurations.
func diff(before, after *schema.SiteConfiguration) (fields map[string]struct{}) {
	fields = make(map[string]struct{})
	beforeFields := getJSONFields(before)
	afterFields := getJSONFields(after)
	for fieldName, beforeField := range beforeFields {
		afterField := afterFields[fieldName]
		if !reflect.DeepEqual(beforeField, afterField) {
			fields[fieldName] = struct{}{}
		}
	}
	return fields
}

func getJSONFields(vv interface{}) (fields map[string]interface{}) {
	fields = make(map[string]interface{})
	v := reflect.ValueOf(vv).Elem()
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		tag := v.Type().Field(i).Tag.Get("json")
		if tag == "" {
			// should never happen, and if it does this func cannot work.
			panic(fmt.Sprintf("missing json struct field tag on %T field %q", v.Interface(), v.Type().Field(i).Name))
		}
		if ef, ok := f.Interface().(*schema.ExperimentalFeatures); ok && ef != nil {
			for fieldName, fieldValue := range getJSONFields(ef) {
				fields["experimentalFeatures::"+fieldName] = fieldValue
			}
			continue
		}
		fieldName := parseJSONTag(tag)
		fields[fieldName] = f.Interface()
	}
	return fields
}

// parseJSONTag parses a JSON struct field tag to return the JSON field name.
func parseJSONTag(tag string) string {
	if idx := strings.Index(tag, ","); idx != -1 {
		return tag[:idx]
	}
	return tag
}
