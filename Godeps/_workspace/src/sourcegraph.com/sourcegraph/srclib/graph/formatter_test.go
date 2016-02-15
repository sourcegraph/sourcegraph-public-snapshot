package graph

import (
	"fmt"
	"testing"
)

type testFormatter struct{}

func (_ testFormatter) Name(qual Qualification) string {
	switch qual {
	case Unqualified:
		return "name"
	case ScopeQualified:
		return "scope.name"
	case DepQualified:
		return "imp.scope.name"
	case RepositoryWideQualified:
		return "dir/lib.scope.name"
	case LanguageWideQualified:
		return "lib.scope.name"
	}
	panic("Name: unrecognized Qualification: " + fmt.Sprint(qual))
}

func (_ testFormatter) Type(qual Qualification) string {
	switch qual {
	case Unqualified:
		return "typeName"
	case ScopeQualified:
		return "scope.typeName"
	case DepQualified:
		return "imp.scope.typeName"
	case RepositoryWideQualified:
		return "dir/lib.scope.typeName"
	case LanguageWideQualified:
		return "lib.scope.typeName"
	}
	panic("Type: unrecognized Qualification: " + fmt.Sprint(qual))
}

func (_ testFormatter) Language() string             { return "lang" }
func (_ testFormatter) DefKeyword() string           { return "defkw" }
func (_ testFormatter) NameAndTypeSeparator() string { return "_" }
func (_ testFormatter) Kind() string                 { return "kind" }

func TestPrintFormatter(t *testing.T) {
	const unitType = "TestFormatter"
	RegisterMakeDefFormatter("TestFormatter", func(*Def) DefFormatter { return testFormatter{} })
	def := &Def{DefKey: DefKey{UnitType: unitType}}
	tests := []struct {
		format string
		want   string
	}{
		{"%n", "name"},
		{"%.0n", "name"},
		{"%.1n", "scope.name"},
		{"%.2n", "imp.scope.name"},
		{"%.3n", "dir/lib.scope.name"},
		{"%.4n", "lib.scope.name"},
		{"%t", "typeName"},
		{"%.0t", "typeName"},
		{"%.1t", "scope.typeName"},
		{"%.2t", "imp.scope.typeName"},
		{"%.3t", "dir/lib.scope.typeName"},
		{"%.4t", "lib.scope.typeName"},
		{"% t", "_typeName"},
		{"%w", "defkw"},
		{"%k", "kind"},
	}
	for _, test := range tests {
		str := fmt.Sprintf(test.format, PrintFormatter(def))
		if str != test.want {
			t.Errorf("Sprintf(%q, def): got %q, want %q", test.format, str, test.want)
		}
	}
}
