// Copyright 2016 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bigquery

import (
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"regexp"
	"strings"
	"time"

	"cloud.google.com/go/civil"
	"cloud.google.com/go/internal/fields"
	bq "google.golang.org/api/bigquery/v2"
)

// See https://cloud.google.com/bigquery/docs/reference/standard-sql/data-types#timestamp-type.
var (
	timestampFormat = "2006-01-02 15:04:05.999999-07:00"
	dateTimeFormat  = "2006-01-02 15:04:05"
)

var (
	// See https://cloud.google.com/bigquery/docs/reference/rest/v2/tables#schema.fields.name
	validFieldName = regexp.MustCompile("^[a-zA-Z_][a-zA-Z0-9_]{0,127}$")
)

const nullableTagOption = "nullable"

func bqTagParser(t reflect.StructTag) (name string, keep bool, other interface{}, err error) {
	name, keep, opts, err := fields.ParseStandardTag("bigquery", t)
	if err != nil {
		return "", false, nil, err
	}
	if name != "" && !validFieldName.MatchString(name) {
		return "", false, nil, invalidFieldNameError(name)
	}
	for _, opt := range opts {
		if opt != nullableTagOption {
			return "", false, nil, fmt.Errorf(
				"bigquery: invalid tag option %q. The only valid option is %q",
				opt, nullableTagOption)
		}
	}
	return name, keep, opts, nil
}

type invalidFieldNameError string

func (e invalidFieldNameError) Error() string {
	return fmt.Sprintf("bigquery: invalid name %q of field in struct", string(e))
}

var fieldCache = fields.NewCache(bqTagParser, nil, nil)

var (
	int64ParamType      = &bq.QueryParameterType{Type: "INT64"}
	float64ParamType    = &bq.QueryParameterType{Type: "FLOAT64"}
	boolParamType       = &bq.QueryParameterType{Type: "BOOL"}
	stringParamType     = &bq.QueryParameterType{Type: "STRING"}
	bytesParamType      = &bq.QueryParameterType{Type: "BYTES"}
	dateParamType       = &bq.QueryParameterType{Type: "DATE"}
	timeParamType       = &bq.QueryParameterType{Type: "TIME"}
	dateTimeParamType   = &bq.QueryParameterType{Type: "DATETIME"}
	timestampParamType  = &bq.QueryParameterType{Type: "TIMESTAMP"}
	numericParamType    = &bq.QueryParameterType{Type: "NUMERIC"}
	bigNumericParamType = &bq.QueryParameterType{Type: "BIGNUMERIC"}
	geographyParamType  = &bq.QueryParameterType{Type: "GEOGRAPHY"}
	intervalParamType   = &bq.QueryParameterType{Type: "INTERVAL"}
	jsonParamType       = &bq.QueryParameterType{Type: "JSON"}
)

var (
	typeOfDate                = reflect.TypeOf(civil.Date{})
	typeOfTime                = reflect.TypeOf(civil.Time{})
	typeOfDateTime            = reflect.TypeOf(civil.DateTime{})
	typeOfGoTime              = reflect.TypeOf(time.Time{})
	typeOfRat                 = reflect.TypeOf(&big.Rat{})
	typeOfIntervalValue       = reflect.TypeOf(&IntervalValue{})
	typeOfQueryParameterValue = reflect.TypeOf(&QueryParameterValue{})
)

// A QueryParameter is a parameter to a query.
type QueryParameter struct {
	// Name is used for named parameter mode.
	// It must match the name in the query case-insensitively.
	Name string

	// Value is the value of the parameter.
	//
	// When you create a QueryParameter to send to BigQuery, the following Go types
	// are supported, with their corresponding Bigquery types:
	// int, int8, int16, int32, int64, uint8, uint16, uint32: INT64
	//   Note that uint, uint64 and uintptr are not supported, because
	//   they may contain values that cannot fit into a 64-bit signed integer.
	// float32, float64: FLOAT64
	// bool: BOOL
	// string: STRING
	// []byte: BYTES
	// time.Time: TIMESTAMP
	// *big.Rat: NUMERIC
	// *IntervalValue: INTERVAL
	// Arrays and slices of the above.
	// Structs of the above. Only the exported fields are used.
	//
	// For scalar values, you can supply the Null types within this library
	// to send the appropriate NULL values (e.g. NullInt64, NullString, etc).
	//
	// To specify query parameters explicitly rather by inference, *QueryParameterValue can be used.
	// For example, a BIGNUMERIC can be specified like this:
	// &QueryParameterValue{
	//		Type: StandardSQLDataType{
	//			TypeKind: "BIGNUMERIC",
	//		},
	//		Value: BigNumericString(*big.Rat),
	//	}
	//
	// When a QueryParameter is returned inside a QueryConfig from a call to
	// Job.Config:
	// Integers are of type int64.
	// Floating-point values are of type float64.
	// Arrays are of type []interface{}, regardless of the array element type.
	// Structs are of type map[string]interface{}.
	//
	// When valid (non-null) Null types are sent, they come back as the Go types indicated
	// above.  Null strings will report in query statistics as a valid empty
	// string.
	Value interface{}
}

// QueryParameterValue is a go type for representing a explicit typed QueryParameter.
type QueryParameterValue struct {
	// Type specifies the parameter type. See StandardSQLDataType for more.
	// Scalar parameters and more complex types can be defined within this field.
	// See examples on the value fields.
	Type StandardSQLDataType

	// Value is the value of the parameter, if a simple scalar type.
	// The default behavior for scalar values is to do type inference
	// and format it accordingly.
	// Because of that, depending on the parameter type, is recommended
	// to send value as a String.
	// We provide some formatter functions for some types:
	//   CivilTimeString(civil.Time)
	//   CivilDateTimeString(civil.DateTime)
	//   NumericString(*big.Rat)
	//   BigNumericString(*big.Rat)
	//   IntervalString(*IntervalValue)
	//
	// Example:
	//
	// &QueryParameterValue{
	// 		Type: StandardSQLDataType{
	//			TypeKind: "BIGNUMERIC",
	//		},
	//		Value: BigNumericString(*big.Rat),
	//	}
	Value interface{}

	// ArrayValue is the array of values for the parameter.
	//
	// Must be used with QueryParameterValue.Type being a StandardSQLDataType
	// with ArrayElementType filled with the given element type.
	//
	// Example of an array of strings :
	// &QueryParameterValue{
	//		Type: &StandardSQLDataType{
	// 			ArrayElementType: &StandardSQLDataType{
	//				TypeKind: "STRING",
	//			},
	//		},
	//		ArrayValue: []QueryParameterValue{
	//			{Value: "a"},
	//			{Value: "b"},
	//		},
	//	}
	//
	// Example of an array of structs :
	// &QueryParameterValue{
	//		Type: &StandardSQLDataType{
	// 			ArrayElementType: &StandardSQLDataType{
	//	 			StructType: &StandardSQLDataType{
	//					Fields: []*StandardSQLField{
	//						{
	//							Name: "NumberField",
	//							Type: &StandardSQLDataType{
	//								TypeKind: "INT64",
	//							},
	//						},
	//					},
	//				},
	//			},
	// 		},
	//		ArrayValue: []QueryParameterValue{
	//			{StructValue: map[string]QueryParameterValue{
	//				"NumberField": {
	//					Value: int64(42),
	//				},
	// 			}},
	// 			{StructValue: map[string]QueryParameterValue{
	//				"NumberField": {
	//					Value: int64(43),
	//				},
	// 			}},
	//		},
	//	}
	ArrayValue []QueryParameterValue

	// StructValue is the struct field values for the parameter.
	//
	// Must be used with QueryParameterValue.Type being a StandardSQLDataType
	// with StructType filled with the given field types.
	//
	// Example:
	//
	// &QueryParameterValue{
	//		Type: &StandardSQLDataType{
	// 			StructType{
	//				Fields: []*StandardSQLField{
	//					{
	//						Name: "StringField",
	//						Type: &StandardSQLDataType{
	//							TypeKind: "STRING",
	//						},
	//					},
	//					{
	//						Name: "NumberField",
	//						Type: &StandardSQLDataType{
	//							TypeKind: "INT64",
	//						},
	//					},
	//				},
	//			},
	//		},
	//		StructValue: []map[string]QueryParameterValue{
	//			"NumberField": {
	//				Value: int64(42),
	//			},
	//			"StringField": {
	//				Value: "Value",
	//			},
	//		},
	//	}
	StructValue map[string]QueryParameterValue
}

func (p QueryParameterValue) toBQParamType() *bq.QueryParameterType {
	return p.Type.toBQParamType()
}

func (p QueryParameterValue) toBQParamValue() (*bq.QueryParameterValue, error) {
	if len(p.ArrayValue) > 0 {
		pv := &bq.QueryParameterValue{}
		pv.ArrayValues = []*bq.QueryParameterValue{}
		for _, v := range p.ArrayValue {
			val, err := v.toBQParamValue()
			if err != nil {
				return nil, err
			}
			pv.ArrayValues = append(pv.ArrayValues, val)
		}
		return pv, nil
	}
	if len(p.StructValue) > 0 {
		pv := &bq.QueryParameterValue{}
		pv.StructValues = map[string]bq.QueryParameterValue{}
		for name, param := range p.StructValue {
			v, err := param.toBQParamValue()
			if err != nil {
				return nil, err
			}
			pv.StructValues[name] = *v
		}
		return pv, nil
	}
	pv, err := paramValue(reflect.ValueOf(p.Value))
	if err != nil {
		return nil, err
	}
	return pv, nil
}

func (p QueryParameter) toBQ() (*bq.QueryParameter, error) {
	v := reflect.ValueOf(p.Value)
	pv, err := paramValue(v)
	if err != nil {
		return nil, err
	}
	pt, err := paramType(reflect.TypeOf(p.Value), v)
	if err != nil {
		return nil, err
	}
	return &bq.QueryParameter{
		Name:           p.Name,
		ParameterValue: pv,
		ParameterType:  pt,
	}, nil
}

func paramType(t reflect.Type, v reflect.Value) (*bq.QueryParameterType, error) {
	if t == nil {
		return nil, errors.New("bigquery: nil parameter")
	}
	switch t {
	case typeOfDate, typeOfNullDate:
		return dateParamType, nil
	case typeOfTime, typeOfNullTime:
		return timeParamType, nil
	case typeOfDateTime, typeOfNullDateTime:
		return dateTimeParamType, nil
	case typeOfGoTime, typeOfNullTimestamp:
		return timestampParamType, nil
	case typeOfRat:
		return numericParamType, nil
	case typeOfIntervalValue:
		return intervalParamType, nil
	case typeOfNullBool:
		return boolParamType, nil
	case typeOfNullFloat64:
		return float64ParamType, nil
	case typeOfNullInt64:
		return int64ParamType, nil
	case typeOfNullString:
		return stringParamType, nil
	case typeOfNullGeography:
		return geographyParamType, nil
	case typeOfNullJSON:
		return jsonParamType, nil
	case typeOfQueryParameterValue:
		return v.Interface().(*QueryParameterValue).toBQParamType(), nil
	}
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint8, reflect.Uint16, reflect.Uint32:
		return int64ParamType, nil

	case reflect.Float32, reflect.Float64:
		return float64ParamType, nil

	case reflect.Bool:
		return boolParamType, nil

	case reflect.String:
		return stringParamType, nil

	case reflect.Slice:
		if t.Elem().Kind() == reflect.Uint8 {
			return bytesParamType, nil
		}
		fallthrough

	case reflect.Array:
		et, err := paramType(t.Elem(), v)
		if err != nil {
			return nil, err
		}
		return &bq.QueryParameterType{Type: "ARRAY", ArrayType: et}, nil

	case reflect.Ptr:
		if t.Elem().Kind() != reflect.Struct {
			break
		}
		t = t.Elem()
		fallthrough

	case reflect.Struct:
		var fts []*bq.QueryParameterTypeStructTypes
		fields, err := fieldCache.Fields(t)
		if err != nil {
			return nil, err
		}
		for _, f := range fields {
			prefixes := []string{"*", "[]"} // check pointer and arrays
			for _, prefix := range prefixes {
				if strings.TrimPrefix(t.String(), prefix) == strings.TrimPrefix(f.Type.String(), prefix) {
					return nil, fmt.Errorf("bigquery: Go type %s cannot be represented as a parameter due to an attribute cycle/recursion detected", t)
				}
			}
			pt, err := paramType(f.Type, v)
			if err != nil {
				return nil, err
			}
			fts = append(fts, &bq.QueryParameterTypeStructTypes{
				Name: f.Name,
				Type: pt,
			})
		}
		return &bq.QueryParameterType{Type: "STRUCT", StructTypes: fts}, nil
	}
	return nil, fmt.Errorf("bigquery: Go type %s cannot be represented as a parameter type", t)
}

func paramValue(v reflect.Value) (*bq.QueryParameterValue, error) {
	res := &bq.QueryParameterValue{}
	if !v.IsValid() {
		return res, errors.New("bigquery: nil parameter")
	}
	t := v.Type()
	switch t {

	// Handle all the custom null types as a group first, as they all have the same logic when invalid.
	case typeOfNullInt64,
		typeOfNullString,
		typeOfNullGeography,
		typeOfNullFloat64,
		typeOfNullBool,
		typeOfNullTimestamp,
		typeOfNullDate,
		typeOfNullTime,
		typeOfNullDateTime,
		typeOfNullJSON:
		// Shared:  If the Null type isn't valid, we have no value to send.
		// However, the backend requires us to send the QueryParameterValue with
		// the fields empty.
		if !v.FieldByName("Valid").Bool() {
			// Ensure we don't send a default value by using NullFields in the JSON
			// serialization.
			res.NullFields = append(res.NullFields, "Value")
			return res, nil
		}
		// For cases where the Null type is valid, populate the scalar value as needed.
		switch t {
		case typeOfNullInt64:
			res.Value = fmt.Sprint(v.FieldByName("Int64").Interface())
		case typeOfNullString:
			res.Value = fmt.Sprint(v.FieldByName("StringVal").Interface())
		case typeOfNullGeography:
			res.Value = fmt.Sprint(v.FieldByName("GeographyVal").Interface())
		case typeOfNullJSON:
			res.Value = fmt.Sprint(v.FieldByName("JSONVal").Interface())
		case typeOfNullFloat64:
			res.Value = fmt.Sprint(v.FieldByName("Float64").Interface())
		case typeOfNullBool:
			res.Value = fmt.Sprint(v.FieldByName("Bool").Interface())
		case typeOfNullTimestamp:
			res.Value = v.FieldByName("Timestamp").Interface().(time.Time).Format(timestampFormat)
		case typeOfNullDate:
			res.Value = v.FieldByName("Date").Interface().(civil.Date).String()
		case typeOfNullTime:
			res.Value = CivilTimeString(v.FieldByName("Time").Interface().(civil.Time))
		case typeOfNullDateTime:
			res.Value = CivilDateTimeString(v.FieldByName("DateTime").Interface().(civil.DateTime))
		}
		// We expect to produce a value in all these cases, so force send if the result is the empty
		// string.
		if res.Value == "" {
			res.ForceSendFields = append(res.ForceSendFields, "Value")
		}
		return res, nil

	case typeOfDate:
		res.Value = v.Interface().(civil.Date).String()
		return res, nil
	case typeOfTime:
		// civil.Time has nanosecond resolution, but BigQuery TIME only microsecond.
		// (If we send nanoseconds, then when we try to read the result we get "query job
		// missing destination table").
		res.Value = CivilTimeString(v.Interface().(civil.Time))
		return res, nil

	case typeOfDateTime:
		res.Value = CivilDateTimeString(v.Interface().(civil.DateTime))
		return res, nil

	case typeOfGoTime:
		res.Value = v.Interface().(time.Time).Format(timestampFormat)
		return res, nil

	case typeOfRat:
		// big.Rat types don't communicate scale or precision, so we cannot
		// disambiguate between NUMERIC and BIGNUMERIC.  For now, we'll continue
		// to honor previous behavior and send as Numeric type.
		res.Value = NumericString(v.Interface().(*big.Rat))
		return res, nil
	case typeOfIntervalValue:
		res.Value = IntervalString(v.Interface().(*IntervalValue))
		return res, nil
	case typeOfQueryParameterValue:
		return v.Interface().(*QueryParameterValue).toBQParamValue()
	}
	switch t.Kind() {
	case reflect.Slice:
		if t.Elem().Kind() == reflect.Uint8 {
			res.Value = base64.StdEncoding.EncodeToString(v.Interface().([]byte))
			return res, nil
		}
		fallthrough

	case reflect.Array:
		var vals []*bq.QueryParameterValue
		for i := 0; i < v.Len(); i++ {
			val, err := paramValue(v.Index(i))
			if err != nil {
				return nil, err
			}
			vals = append(vals, val)
		}
		return &bq.QueryParameterValue{ArrayValues: vals}, nil

	case reflect.Ptr:
		if t.Elem().Kind() != reflect.Struct {
			return res, fmt.Errorf("bigquery: Go type %s cannot be represented as a parameter value", t)
		}
		t = t.Elem()
		v = v.Elem()
		if !v.IsValid() {
			// nil pointer becomes empty value
			return res, nil
		}
		fallthrough

	case reflect.Struct:
		fields, err := fieldCache.Fields(t)
		if err != nil {
			return nil, err
		}
		res.StructValues = map[string]bq.QueryParameterValue{}
		for _, f := range fields {
			fv := v.FieldByIndex(f.Index)
			fp, err := paramValue(fv)
			if err != nil {
				return nil, err
			}
			res.StructValues[f.Name] = *fp
		}
		return res, nil
	}
	// None of the above: assume a scalar type. (If it's not a valid type,
	// paramType will catch the error.)
	res.Value = fmt.Sprint(v.Interface())
	// Ensure empty string values are sent.
	if res.Value == "" {
		res.ForceSendFields = append(res.ForceSendFields, "Value")
	}
	return res, nil
}

func bqToQueryParameter(q *bq.QueryParameter) (QueryParameter, error) {
	p := QueryParameter{Name: q.Name}
	val, err := convertParamValue(q.ParameterValue, q.ParameterType)
	if err != nil {
		return QueryParameter{}, err
	}
	p.Value = val
	return p, nil
}

var paramTypeToFieldType = map[string]FieldType{
	int64ParamType.Type:      IntegerFieldType,
	float64ParamType.Type:    FloatFieldType,
	boolParamType.Type:       BooleanFieldType,
	stringParamType.Type:     StringFieldType,
	bytesParamType.Type:      BytesFieldType,
	dateParamType.Type:       DateFieldType,
	timeParamType.Type:       TimeFieldType,
	numericParamType.Type:    NumericFieldType,
	bigNumericParamType.Type: BigNumericFieldType,
	geographyParamType.Type:  GeographyFieldType,
	intervalParamType.Type:   IntervalFieldType,
	jsonParamType.Type:       JSONFieldType,
}

// Convert a parameter value from the service to a Go value. This is similar to, but
// not quite the same as, converting data values.  Namely, rather than returning nil
// directly, we wrap them in the appropriate Null types (NullInt64, etc).
func convertParamValue(qval *bq.QueryParameterValue, qtype *bq.QueryParameterType) (interface{}, error) {
	switch qtype.Type {
	case "ARRAY":
		if qval == nil {
			return []interface{}(nil), nil
		}
		return convertParamArray(qval.ArrayValues, qtype.ArrayType)
	case "STRUCT":
		if qval == nil {
			return map[string]interface{}(nil), nil
		}
		return convertParamStruct(qval.StructValues, qtype.StructTypes)
	case "TIMESTAMP":
		if isNullScalar(qval) {
			return NullTimestamp{Valid: false}, nil
		}
		formats := []string{timestampFormat, time.RFC3339Nano, dateTimeFormat}
		var lastParseErr error
		for _, format := range formats {
			t, err := time.Parse(format, qval.Value)
			if err != nil {
				lastParseErr = err
				continue
			}
			return t, nil
		}
		return nil, lastParseErr

	case "DATETIME":
		if isNullScalar(qval) {
			return NullDateTime{Valid: false}, nil
		}
		return parseCivilDateTime(qval.Value)
	default:
		if isNullScalar(qval) {
			switch qtype.Type {
			case "INT64":
				return NullInt64{Valid: false}, nil
			case "STRING":
				return NullString{Valid: false}, nil
			case "FLOAT64":
				return NullFloat64{Valid: false}, nil
			case "BOOL":
				return NullBool{Valid: false}, nil
			case "DATE":
				return NullDate{Valid: false}, nil
			case "TIME":
				return NullTime{Valid: false}, nil
			case "GEOGRAPHY":
				return NullGeography{Valid: false}, nil
			case "JSON":
				return NullJSON{Valid: false}, nil
			}

		}
		return convertBasicType(qval.Value, paramTypeToFieldType[qtype.Type])
	}
}

// isNullScalar determines if the input is meant to represent a null scalar
// value.
func isNullScalar(qval *bq.QueryParameterValue) bool {
	if qval == nil {
		return true
	}
	for _, v := range qval.NullFields {
		if v == "Value" {
			return true
		}
	}
	return false
}

// convertParamArray converts a query parameter array value to a Go value. It
// always returns a []interface{}.
func convertParamArray(elVals []*bq.QueryParameterValue, elType *bq.QueryParameterType) ([]interface{}, error) {
	var vals []interface{}
	for _, el := range elVals {
		val, err := convertParamValue(el, elType)
		if err != nil {
			return nil, err
		}
		vals = append(vals, val)
	}
	return vals, nil
}

// convertParamStruct converts a query parameter struct value into a Go value. It
// always returns a map[string]interface{}.
func convertParamStruct(sVals map[string]bq.QueryParameterValue, sTypes []*bq.QueryParameterTypeStructTypes) (map[string]interface{}, error) {
	vals := map[string]interface{}{}
	for _, st := range sTypes {
		if sv, ok := sVals[st.Name]; ok {
			val, err := convertParamValue(&sv, st.Type)
			if err != nil {
				return nil, err
			}
			vals[st.Name] = val
		} else {
			vals[st.Name] = nil
		}
	}
	return vals, nil
}
