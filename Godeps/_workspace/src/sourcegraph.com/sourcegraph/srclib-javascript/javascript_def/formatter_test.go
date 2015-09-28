package javascript_def

import (
	"encoding/json"
	"testing"

	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sourcegraph/srclib/util/sqltypes"
)

func defDataJSON(si defData) sqltypes.JSON {
	b, err := json.Marshal(si)
	if err != nil {
		panic(err)
	}
	return b
}

func TestDefFormatter_Name(t *testing.T) {
	tests := []struct {
		def  *graph.Def
		qual graph.Qualification
		want string
	}{
		{
			// unqualified
			def: &graph.Def{
				Name: "name",
				Data: sqltypes.JSON(`{}`),
			},
			qual: graph.Unqualified,
			want: "name",
		},
		{
			// qualify defs with scope
			def: &graph.Def{
				Data: defDataJSON(defData{Key: DefPath{Path: "a.b"}}),
			},
			qual: graph.ScopeQualified,
			want: "a.b",
		},
		{
			// qualify file defs with scope
			def: &graph.Def{
				Data: defDataJSON(defData{Key: DefPath{Namespace: "file", Path: "a.b.@local123.c.d"}}),
			},
			qual: graph.ScopeQualified,
			want: "c.d",
		},
		{
			// qualify defs with module basename (dep-qualified)
			def: &graph.Def{
				Data: defDataJSON(defData{Key: DefPath{Path: "a.b", Module: "c/d"}}),
			},
			qual: graph.DepQualified,
			want: "d.a.b",
		},
		{
			// qualify defs with pkg root and module (repository-wide)
			def: &graph.Def{
				DefKey: graph.DefKey{Unit: "x/y"},
				Data:   defDataJSON(defData{Key: DefPath{Path: "a.b", Module: "c/d"}}),
			},
			qual: graph.RepositoryWideQualified,
			want: "x/y/c/d.a.b",
		},
		{
			// qualify defs with full path (lang-wide)
			def: &graph.Def{
				DefKey: graph.DefKey{Repo: "t/u", Unit: "x/y"},
				Data:   defDataJSON(defData{Key: DefPath{Path: "a.b", Module: "c/d"}}),
			},
			qual: graph.LanguageWideQualified,
			want: "t/u/x/y/c/d.a.b",
		},
	}
	for _, test := range tests {
		sf := newDefFormatter(test.def)
		name := sf.Name(test.qual)
		if name != test.want {
			t.Errorf("%v qual %q: got %q, want %q", test.def, test.qual, name, test.want)
		}
	}
}
