package jsonschema

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
)

type Enum interface{}
type EnumList []Enum
type Format string
type PrimitiveType string
type PrimitiveTypeList []PrimitiveType

const (
	UnspecifiedType PrimitiveType = "unspecified"
	NullType                      = "null"
	BooleanType                   = "boolean"
	ObjectType                    = "object"
	ArrayType                     = "array"
	NumberType                    = "number"
	StringType                    = "string"
	IntegerType                   = "integer"
)

func (t PrimitiveType) String() string {
	return string(t)
}

func (l PrimitiveTypeList) Len() int {
	return len(l)
}

func (l PrimitiveTypeList) Less(i, j int) bool {
	return l[i] < l[j]
}

func (l PrimitiveTypeList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

func (l *PrimitiveTypeList) MarshalJSON() ([]byte, error) {
	if len(*l) == 1 {
		return json.Marshal((*l)[0])
	}
	return json.Marshal([]PrimitiveType(*l))
}

func (l *PrimitiveTypeList) UnmarshalJSON(buf []byte) error {
	var sl []string
	if len(buf) > 0 && buf[0] == '[' {
		if err := json.Unmarshal(buf, &sl); err != nil {
			return errors.Wrap(err, `failed to parse primitive types list`)
		}
	} else {
		var s string
		if err := json.Unmarshal(buf, &s); err != nil {
			return errors.Wrap(err, `failed to parse primitive types list`)
		}
		sl = []string{s}
	}

	ptl := make(PrimitiveTypeList, 0, len(sl))
	for _, s := range sl {
		var pt PrimitiveType
		switch s {
		case "null":
			pt = NullType
		case "boolean":
			pt = BooleanType
		case "object":
			pt = ObjectType
		case "array":
			pt = ArrayType
		case "number":
			pt = NumberType
		case "string":
			pt = StringType
		case "integer":
			pt = IntegerType
		default:
			return fmt.Errorf(`invalid primitive type: %s`, s)
		}
		ptl = append(ptl, pt)
	}

	*l = ptl
	return nil
}
