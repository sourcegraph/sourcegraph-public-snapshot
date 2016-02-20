// Package sample provides a Go def formatter for sample source unit
// defs.
package sample

import (
	"fmt"

	"sourcegraph.com/sourcegraph/srclib/graph"
)

const sampleSourceUnitType = "sample"

func init() {
	graph.RegisterMakeDefFormatter(sampleSourceUnitType, func(*graph.Def) graph.DefFormatter { return testFormatter{} })
}

type testFormatter struct{}

func (_ testFormatter) Name(qual graph.Qualification) string {
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

func (_ testFormatter) Type(qual graph.Qualification) string {
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

func (_ testFormatter) Language() string             { return "lang" }
func (_ testFormatter) DefKeyword() string           { return "defkw" }
func (_ testFormatter) NameAndTypeSeparator() string { return "_" }
func (_ testFormatter) Kind() string                 { return "kind" }
