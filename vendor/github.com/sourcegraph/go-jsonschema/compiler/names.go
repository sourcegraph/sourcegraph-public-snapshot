package compiler

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/sourcegraph/go-jsonschema/jsonschema"
)

func goNameForSchema(schema *jsonschema.Schema, location schemaLocation) (string, error) {
	var name string
	if schema.Title != nil {
		name = *schema.Title
	}
	if name == "" {
		// Take the nearest ancestor reference token that is defined by the schema author and is not
		// a JSON Schema keyword (e.g., use a property name, not "properties" or "items" itself).
		for i := len(location.rel) - 1; i >= 0; i-- {
			refToken := location.rel[i]
			if refToken.Name != "" && !refToken.Keyword {
				name = refToken.Name
				break
			}
		}
	}
	if name == "" {
		return "", fmt.Errorf("schema at %q has no viable name", jsonschema.EncodeReferenceTokens(location.rel))
	}

	return toGoName(name, "Schema_"), nil
}

// toGoName converts name to a nice-looking Go exported identifier. The prefix (which must itself be
// a valid, exported Go identifier) is prepended if needed to produce a valid, exported Go
// identifier.
func toGoName(name, prefix string) string {
	// See https://golang.org/ref/spec#Identifiers for the Go identifier specification. The spec for
	// the first character is even stricter (it must be a Unicode letter).
	isValidGoIdentifierChar := func(r rune) bool {
		return unicode.IsLetter(r) || unicode.IsNumber(r) || r == '_'
	}

	titleNext := true
	initial := true
	name = strings.Map(
		func(r rune) rune {
			if initial {
				initial = false
				if unicode.IsLetter(r) {
					prefix = ""
				}
			}
			if !isValidGoIdentifierChar(r) {
				titleNext = true
				return -1
			}
			if titleNext {
				titleNext = false
				return unicode.ToTitle(r)
			}
			return r
		},
		name)
	return prefix + name
}
