// Package schemautil contains schema utilities for working with URL
// querystrings and form POST data.
package schemautil

import (
	"net/url"
	"reflect"

	"src.sourcegraph.com/sourcegraph/errcode"

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

// currentQueryIsSupersetOf returns true iff currentQuery is a
// "superset" of other.  For example, in rough notation,
// currentQueryMatches(`?a=1&b=2`, `?b=2`) is true, but
// currentQueryMatches(`?a=1`, `?a=2`) and currentQueryMatches(`?a=1`,
// `?b=1`) are false.
func currentQueryIsSupersetOf(currentQuery url.Values, otherQuery url.Values) bool {
	for k, v := range currentQuery {
		v2, present := otherQuery[k]
		if present {
			delete(otherQuery, k)
			if !reflect.DeepEqual(v, v2) {
				return false
			}
		}
	}

	// reject if keys (with non-nil values) exist in other that aren't in currentQuery
	for k, v := range otherQuery {
		if v == nil {
			delete(otherQuery, k)
		}
	}
	if len(otherQuery) > 0 {
		return false
	}

	return true
}
