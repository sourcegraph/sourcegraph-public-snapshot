package def

import (
	"encoding/json"
	"testing"

	"github.com/jmoiron/sqlx/types"
	"sourcegraph.com/sourcegraph/srclib-go/gog/definfo"
	"sourcegraph.com/sourcegraph/srclib/graph"
)

func defInfo(si DefData) types.JsonText {
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
				Data: types.JsonText(`{}`),
			},
			qual: graph.Unqualified,
			want: "name",
		},
		{
			// qualify methods with receiver
			def: &graph.Def{
				Name: "name",
				Data: defInfo(DefData{DefInfo: definfo.DefInfo{Receiver: "*T", Kind: definfo.Method}}),
			},
			qual: graph.ScopeQualified,
			want: "(*T).name",
		},
		{
			// all funcs are at pkg scope
			def: &graph.Def{
				Name: "name",
				Data: defInfo(DefData{DefInfo: definfo.DefInfo{PkgName: "mypkg", Kind: definfo.Func}}),
			},
			qual: graph.ScopeQualified,
			want: "name",
		},
		{
			// qualify funcs with pkg
			def: &graph.Def{
				Name: "Name",
				Data: defInfo(DefData{DefInfo: definfo.DefInfo{PkgName: "mypkg", Kind: definfo.Func}}),
			},
			qual: graph.DepQualified,
			want: "mypkg.Name",
		},
		{
			// qualify methods with receiver pkg
			def: &graph.Def{
				Name: "Name",
				Data: defInfo(DefData{DefInfo: definfo.DefInfo{Receiver: "*T", PkgName: "mypkg", Kind: definfo.Method}}),
			},
			qual: graph.DepQualified,
			want: "(*mypkg.T).Name",
		},
		{
			// qualify pkgs with import path relative to repo root
			def: &graph.Def{
				DefKey: graph.DefKey{Repo: "example.com/foo"},
				Name:   "subpkg",
				Kind:   "package",
				Data:   defInfo(DefData{PackageImportPath: "example.com/foo/mypkg/subpkg", DefInfo: definfo.DefInfo{PkgName: "subpkg", Kind: definfo.Package}}),
			},
			qual: graph.RepositoryWideQualified,
			want: "foo/mypkg/subpkg",
		},
		{
			// qualify funcs with import path
			def: &graph.Def{
				Name: "Name",
				Data: defInfo(DefData{PackageImportPath: "a/b", DefInfo: definfo.DefInfo{PkgName: "x", Kind: definfo.Func}}),
			},
			qual: graph.LanguageWideQualified,
			want: "a/b.Name",
		},
		{
			// qualify methods with receiver pkg
			def: &graph.Def{
				Name: "Name",
				Data: defInfo(DefData{PackageImportPath: "a/b", DefInfo: definfo.DefInfo{Receiver: "*T", PkgName: "x", Kind: definfo.Method}}),
			},
			qual: graph.LanguageWideQualified,
			want: "(*a/b.T).Name",
		},
		{
			// qualify pkgs with full import path
			def: &graph.Def{
				Name: "x",
				Data: defInfo(DefData{PackageImportPath: "a/b", DefInfo: definfo.DefInfo{PkgName: "x", Kind: definfo.Package}}),
			},
			qual: graph.LanguageWideQualified,
			want: "a/b",
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
