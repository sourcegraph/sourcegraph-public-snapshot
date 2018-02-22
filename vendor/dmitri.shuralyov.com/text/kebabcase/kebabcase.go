// Package kebabcase provides a parser for identifier names
// using kebab-case naming convention.
//
// Reference: https://en.wikipedia.org/wiki/Naming_convention_(programming)#Multiple-word_identifiers.
package kebabcase

import (
	"strings"

	"github.com/shurcooL/graphql/ident"
)

// Parse parses a kebab-case identifier name.
//
// E.g., "client-mutation-id" -> {"client", "mutation", "id"}.
func Parse(name string) ident.Name {
	return ident.Name(strings.Split(name, "-"))
}
