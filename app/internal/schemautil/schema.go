// Package schemautil contains schema utilities for working with URL
// querystrings and form POST data.
package schemautil

import (
	"net/url"
	"reflect"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/errcode"

	"github.com/google/go-querystring/query"
	"github.com/gorilla/schema"
)

var decoder = schema.NewDecoder()

func init() {
	decoder.IgnoreUnknownKeys(true)
}

func Decode(dst interface{}, src url.Values) error {
	if err := decoder.Decode(dst, src); err != nil {
		return &errcode.HTTPErr{Status: 400, Err: err}
	}
	return nil
}

type schemaMatcher interface {
	matchesSchema(other interface{}) bool
}

func schemaMatches(schema, other interface{}) bool {
	if sm, ok := schema.(schemaMatcher); ok {
		return sm.matchesSchema(other)
	}
	schemaV, err := query.Values(schema)
	if err != nil {
		panic(err.Error())
	}
	otherV, err := query.Values(other)
	if err != nil {
		panic(err.Error())
	}
	return reflect.DeepEqual(schemaV, otherV)
}

// SchemaMatchesExceptListAndSortOptions returns true if schema is
// equal to other, except for fields that are related to pagination
// and sorting.
func SchemaMatchesExceptListAndSortOptions(schema, other interface{}) bool {
	clearListOpt := func(sch interface{}) interface{} {
		v, err := query.Values(sch)
		if err != nil {
			panic(err.Error())
		}
		delete(v, "Page")
		delete(v, "PerPage")
		delete(v, "Sort")
		delete(v, "Direction")

		t := reflect.TypeOf(sch)
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		v2 := reflect.New(t)

		if err := decoder.Decode(v2.Interface(), v); err != nil {
			panic(err.Error())
		}
		return v2.Interface()
	}
	return schemaMatches(clearListOpt(schema), clearListOpt(other))
}

// URLWithSchema returns a URL querystring (beginning with "?") that
// encodes v.
func URLWithSchema(v interface{}) string {
	qs, err := query.Values(v)
	if err != nil {
		panic(err.Error())
	}
	return "?" + qs.Encode()
}
