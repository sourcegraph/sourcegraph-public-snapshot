package python

import (
	"testing"

	"sourcegraph.com/sourcegraph/srclib/graph"
)

func TestDefFormatter_Name(t *testing.T) {
	tests := []struct {
		def  *graph.Def
		want map[graph.Qualification]string
	}{{
		def:  defInfo{Name: "name"}.Def(),
		want: map[graph.Qualification]string{graph.Unqualified: "name"},
	}, {
		def: defInfo{Repo: "g.com/o/r", TreePath: "a/b/c", File: "a/b.py"}.Def(),
		want: map[graph.Qualification]string{
			graph.ScopeQualified: "c", graph.DepQualified: "b.c", graph.RepositoryWideQualified: "a.b.c", graph.LanguageWideQualified: "g.com/o/r/a.b.c",
		},
	}, {
		def: defInfo{Repo: "g.com/o/r", TreePath: "a/b/c", File: "a.py"}.Def(),
		want: map[graph.Qualification]string{
			graph.ScopeQualified: "b.c", graph.DepQualified: "a.b.c", graph.RepositoryWideQualified: "a.b.c", graph.LanguageWideQualified: "g.com/o/r/a.b.c",
		},
	}, {
		def: defInfo{Repo: "g.com/o/r", TreePath: "aa/a/b", File: "aa/a.py"}.Def(),
		want: map[graph.Qualification]string{
			graph.ScopeQualified: "b", graph.DepQualified: "a.b", graph.RepositoryWideQualified: "aa.a.b", graph.LanguageWideQualified: "g.com/o/r/aa.a.b",
		},
	}, {
		def: defInfo{Repo: "g.com/o/r", TreePath: "a/b", File: "a/__init__.py"}.Def(),
		want: map[graph.Qualification]string{
			graph.ScopeQualified: "b", graph.DepQualified: "a.b", graph.RepositoryWideQualified: "a.b", graph.LanguageWideQualified: "g.com/o/r/a.b",
		},
	}}

	for _, test := range tests {
		sf := newDefFormatter(test.def)
		for qual, expName := range test.want {
			name := sf.Name(qual)
			if expName != name {
				t.Errorf("%v qual %q: want %q but got %q", test.def, qual, expName, name)
			}
		}
	}
}

type defInfo struct {
	SID         int
	Repo        string
	CommitID    string
	UnitType    string
	Unit        string
	Path        string
	File        string
	Name        string
	TreePath    string
	NotExported bool
	Data        []byte
}

func (s defInfo) Def() *graph.Def {
	repo := s.Repo
	if repo == "" {
		repo = "r"
	}
	unitType := s.UnitType
	if unitType == "" {
		unitType = "t"
	}
	unit := s.Unit
	if unit == "" {
		unit = "u"
	}
	data := s.Data
	if data == nil {
		data = []byte(`{}`)
	}
	return &graph.Def{
		DefKey:   graph.DefKey{Repo: repo, CommitID: s.CommitID, UnitType: unitType, Unit: unit, Path: string(s.Path)},
		Name:     s.Name,
		File:     s.File,
		TreePath: s.TreePath,
		Exported: !s.NotExported,
		Data:     data,
	}
}
