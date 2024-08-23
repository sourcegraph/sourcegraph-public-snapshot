package slog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"golang.org/x/xerrors"
)

// Map represents an ordered map of fields.
type Map []Field

var _ json.Marshaler = Map(nil)

// MarshalJSON implements json.Marshaler.
//
// It is guaranteed to return a nil error.
// Any error marshalling a field will become the field's value.
//
// Every field value is encoded with the following process:
//
// 1. json.Marshaller is handled.
//
// 2. xerrors.Formatter is handled.
//
// 3. structs that have a field with a json tag are encoded with json.Marshal.
//
// 4. error and fmt.Stringer is handled.
//
// 5. slices and arrays go through the encode function for every element.
//
// 6. For values that cannot be encoded with json.Marshal, fmt.Sprintf("%+v") is used.
//
// 7. json.Marshal(v) is used for all other values.
func (m Map) MarshalJSON() ([]byte, error) {
	b := &bytes.Buffer{}
	b.WriteByte('{')
	for i, f := range m {
		b.WriteByte('\n')
		b.Write(encode(f.Name))
		b.WriteByte(':')
		b.Write(encode(f.Value))

		if i < len(m)-1 {
			b.WriteByte(',')
		}
	}
	b.WriteByte('}')

	return b.Bytes(), nil
}

func marshalList(rv reflect.Value) []byte {
	b := &bytes.Buffer{}
	b.WriteByte('[')
	for i := 0; i < rv.Len(); i++ {
		b.WriteByte('\n')
		b.Write(encode(rv.Index(i).Interface()))

		if i < rv.Len()-1 {
			b.WriteByte(',')
		}
	}
	b.WriteByte(']')

	return b.Bytes()
}

func encode(v interface{}) []byte {
	switch v := v.(type) {
	case json.Marshaler:
		return encodeJSON(v)
	case xerrors.Formatter:
		return encode(errorChain(v))
	}

	rv := reflect.Indirect(reflect.ValueOf(v))
	if !rv.IsValid() {
		return encodeJSON(v)
	}

	if rv.Kind() == reflect.Struct {
		b, ok := encodeStruct(rv)
		if ok {
			return b
		}
	}

	switch v.(type) {
	case error, fmt.Stringer:
		return encode(fmt.Sprint(v))
	}

	switch rv.Type().Kind() {
	case reflect.Slice:
		if !rv.IsNil() {
			return marshalList(rv)
		}
	case reflect.Array:
		return marshalList(rv)
	case reflect.Struct, reflect.Chan, reflect.Complex64, reflect.Complex128, reflect.Func:
		// These types cannot be directly encoded with json.Marshal.
		// See https://golang.org/pkg/encoding/json/#Marshal
		return encodeJSON(fmt.Sprintf("%+v", v))
	}

	return encodeJSON(v)
}

func encodeStruct(rv reflect.Value) ([]byte, bool) {
	if rv.Kind() == reflect.Struct {
		for i := 0; i < rv.NumField(); i++ {
			ft := rv.Type().Field(i)
			// Found a field with a json tag.
			if ft.Tag.Get("json") != "" {
				return encodeJSON(rv.Interface()), true
			}
		}
	}

	return nil, false
}

func encodeJSON(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		return encode(M(
			Error(xerrors.Errorf("failed to marshal to JSON: %w", err)),
			F("type", reflect.TypeOf(v)),
			F("value", fmt.Sprintf("%+v", v)),
		))
	}
	return b
}

func errorChain(f xerrors.Formatter) []interface{} {
	var errs []interface{}

	next := error(f)
	for {
		f, ok := next.(xerrors.Formatter)
		if !ok {
			errs = append(errs, next)
			return errs
		}

		p := &xerrorPrinter{}
		next = f.FormatError(p)
		errs = append(errs, p.e)
	}
}

type wrapError struct {
	Msg string `json:"msg"`
	Fun string `json:"fun"`
	// file:line
	Loc string `json:"loc"`
}

type xerrorPrinter struct {
	e wrapError
}

func (p *xerrorPrinter) Print(v ...interface{}) {
	s := fmt.Sprint(v...)
	p.write(s)
}

func (p *xerrorPrinter) Printf(f string, v ...interface{}) {
	s := fmt.Sprintf(f, v...)
	p.write(s)
}

func (p *xerrorPrinter) Detail() bool {
	return true
}

func (p *xerrorPrinter) write(s string) {
	s = strings.TrimSpace(s)
	switch {
	case p.e.Msg == "":
		p.e.Msg = s
	case p.e.Fun == "":
		p.e.Fun = s
	case p.e.Loc == "":
		p.e.Loc = s
	}
}

func (m Map) append(m2 Map) Map {
	m3 := make(Map, 0, len(m)+len(m2))
	m3 = append(m3, m...)
	m3 = append(m3, m2...)
	return m3
}
