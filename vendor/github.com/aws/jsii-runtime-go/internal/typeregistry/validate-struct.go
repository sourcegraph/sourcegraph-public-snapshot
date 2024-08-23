package typeregistry

import (
	"fmt"
	"reflect"
)

// ValidateStruct runs validations on the supplied struct to determine whether
// it is valid. In particular, it checks union-typed properties to ensure the
// provided value is of one of the allowed types.
//
// May panic if v is not a pointer to a struct value.
func (t *TypeRegistry) ValidateStruct(v interface{}, d func() string) error {
	rt := reflect.TypeOf(v).Elem()

	info, ok := t.structInfo[rt]
	if !ok {
		return fmt.Errorf("%v: %v is not a know struct type", d(), rt)
	}

	// There may not be a validator (type is simple enough, etc...).
	if info.Validator != nil {
		return info.Validator(v, d)
	}

	return nil
}
