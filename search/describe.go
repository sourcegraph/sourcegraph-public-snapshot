package search

import (
	"fmt"
	"path"
	"strings"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

// Describe returns a string that describes in English what the query
// searches for.
func Describe(resolvedTokens []sourcegraph.Token) string {
	var terms []string
	var scope []string
	for _, tok := range resolvedTokens {
		switch tok := tok.(type) {
		case sourcegraph.RepoToken:
			if tok.URI != "" {
				scope = append(scope, "repository "+tok.URI)
			}
		case sourcegraph.RevToken:
			if tok.Rev != "" {
				scope = append(scope, "at revision "+tok.Rev)
			}
		case sourcegraph.UnitToken:
			if tok.Name != "" {
				s := "in "
				if tok.UnitType != "" {
					s += tok.UnitType + " "
				} else {
					s += "package "
				}
				s += path.Base(tok.Name)
				scope = append(scope, s)
			}
		case sourcegraph.FileToken:
			if tok.Path != "" {
				scope = append(scope, "in "+tok.Path)
			}
		case sourcegraph.UserToken:
			if tok.Login != "" {
				scope = append(scope, "repositories owned by "+tok.Login)
			}
		case sourcegraph.Term:
			if tok.Token() != "" {
				terms = append(terms, tok.Token())
			}
		default:
			panic(fmt.Sprintf("Describe: unexpected token type %T", tok))
		}
	}

	if len(terms) > 0 {
		return fmt.Sprintf("Definitions and files matching %q in %s", strings.Join(terms, " "), strings.Join(scope, " "))
	}
	return "Definitions in " + strings.Join(scope, " ")
}

// Revision returns the revision string that the search query was performed
// against.
func Revision(resolvedTokens []sourcegraph.Token) string {
	for _, tok := range resolvedTokens {
		if v, ok := tok.(sourcegraph.RevToken); ok && v.Rev != "" {
			return v.Rev
		}
	}
	return ""
}
