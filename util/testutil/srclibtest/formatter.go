package srclibtest

import (
	"fmt"

	"sourcegraph.com/sourcegraph/srclib/graph"
)

type Formatter struct{}

func (_ Formatter) Name(qual graph.Qualification) string {
	switch qual {
	case graph.Unqualified:
		return "name"
	case graph.ScopeQualified:
		return "scope.name"
	case graph.DepQualified:
		return "imp.scope.name"
	case graph.RepositoryWideQualified:
		return "dir/lib.scope.name"
	case graph.LanguageWideQualified:
		return "lib.scope.name"
	}
	panic("Name: unrecognized Qualification: " + fmt.Sprint(qual))
}

func (_ Formatter) Type(qual graph.Qualification) string {
	switch qual {
	case graph.Unqualified:
		return "typeName"
	case graph.ScopeQualified:
		return "scope.typeName"
	case graph.DepQualified:
		return "imp.scope.typeName"
	case graph.RepositoryWideQualified:
		return "dir/lib.scope.typeName"
	case graph.LanguageWideQualified:
		return "lib.scope.typeName"
	}
	panic("Type: unrecognized Qualification: " + fmt.Sprint(qual))
}

func (_ Formatter) Language() string             { return "lang" }
func (_ Formatter) DefKeyword() string           { return "defkw" }
func (_ Formatter) NameAndTypeSeparator() string { return "_" }
func (_ Formatter) Kind() string                 { return "kind" }
