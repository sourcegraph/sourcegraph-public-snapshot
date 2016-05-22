package local

import (
	"reflect"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/util/testutil/srclibtest"
	"sourcegraph.com/sourcegraph/srclib/graph"
)

func TestPopulateDefFormatStrings(t *testing.T) {
	graph.RegisterMakeDefFormatter("test", func(*graph.Def) graph.DefFormatter { return srclibtest.Formatter{} })

	def := &sourcegraph.Def{
		Def: graph.Def{
			DefKey: graph.DefKey{Repo: "x.com/r", UnitType: "test", Unit: "u", Path: "p"},
		},
	}

	want := &graph.DefFormatStrings{
		Name: graph.QualFormatStrings{
			Unqualified:             "name",
			ScopeQualified:          "scope.name",
			DepQualified:            "imp.scope.name",
			RepositoryWideQualified: "dir/lib.scope.name",
			LanguageWideQualified:   "lib.scope.name",
		},
		Type: graph.QualFormatStrings{
			Unqualified:             "typeName",
			ScopeQualified:          "scope.typeName",
			DepQualified:            "imp.scope.typeName",
			RepositoryWideQualified: "dir/lib.scope.typeName",
			LanguageWideQualified:   "lib.scope.typeName",
		},
		Language:             "lang",
		DefKeyword:           "defkw",
		NameAndTypeSeparator: "_",
		Kind:                 "kind",
	}

	populateDefFormatStrings(def)

	if !reflect.DeepEqual(def.FmtStrings, want) {
		t.Errorf("got\n%+v\n\nwant\n%+v", def.FmtStrings, want)
	}
}

func TestPopulateDefFormatStrings_noneRegistered(t *testing.T) {
	def := &sourcegraph.Def{
		Def: graph.Def{
			DefKey: graph.DefKey{UnitType: "NoFormatterRegistered"},
		},
	}
	populateDefFormatStrings(def)
	if def.FmtStrings != nil {
		t.Errorf("got def.FmtStrings == %v, want nil (no formatter registered)", def.FmtStrings)
	}
}
