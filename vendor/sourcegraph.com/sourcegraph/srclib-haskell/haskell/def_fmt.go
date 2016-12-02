// Package haskell provides a srclib def formatter for Haskell defs.
package haskell

import (
	"fmt"

	"strings"

	"sourcegraph.com/sourcegraph/srclib/graph"
)

const haskellSourceUnitType = "HaskellPackage"

func init() {
	graph.RegisterMakeDefFormatter(haskellSourceUnitType, func(d *graph.Def) graph.DefFormatter { return haskellFormatter{d} })
}

type haskellFormatter struct {
	d *graph.Def
}

func (f haskellFormatter) Name(qual graph.Qualification) string {
	// TODO(sqs): use data stored
	switch qual {
	case graph.Unqualified:
		return f.d.Name
	case graph.ScopeQualified:
		return strings.Replace(string(f.d.Path), "/", "::", -1)
	case graph.DepQualified:
		return strings.Replace(string(f.d.Path), "/", "::", -1)
	case graph.RepositoryWideQualified:
		return strings.Replace(string(f.d.Path), "/", "::", -1)
	case graph.LanguageWideQualified:
		return strings.Replace(string(f.d.Path), "/", "::", -1)
	}
	panic("Name: unrecognized Qualification: " + fmt.Sprint(qual))
}

func (_ haskellFormatter) Type(qual graph.Qualification) string {
	switch qual {
	case graph.Unqualified:
		return ""
	case graph.ScopeQualified:
		return ""
	case graph.DepQualified:
		return ""
	case graph.RepositoryWideQualified:
		return ""
	case graph.LanguageWideQualified:
		return ""
	}
	panic("Type: unrecognized Qualification: " + fmt.Sprint(qual))
}

func (_ haskellFormatter) Language() string             { return "Haskell" }
func (_ haskellFormatter) DefKeyword() string           { return "def" }
func (_ haskellFormatter) NameAndTypeSeparator() string { return "" }
func (f haskellFormatter) Kind() string                 { return f.d.Kind }
