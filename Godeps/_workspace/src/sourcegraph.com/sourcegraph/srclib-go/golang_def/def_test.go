package golang_def

import (
	"encoding/json"
	"testing"

	"sourcegraph.com/sourcegraph/srclib-go/gog/definfo"
	"sourcegraph.com/sourcegraph/srclib/graph"
)

func defInfo(si DefData) json.RawMessage {
	b, err := json.Marshal(si)
	if err != nil {
		panic(err)
	}
	return b
}

func TestDefFormatter(t *testing.T) {
	tests := []struct {
		def       *graph.Def
		wantNames map[graph.Qualification]string
		wantTypes map[graph.Qualification]string
	}{
		{
			// unqualified
			def: &graph.Def{
				Name: "name",
				Data: json.RawMessage(`{}`),
			},
			wantNames: map[graph.Qualification]string{graph.Unqualified: "name"},
		},
		{
			// qualify methods with receiver
			def: &graph.Def{
				Name: "name",
				Data: defInfo(DefData{DefInfo: definfo.DefInfo{Receiver: "*T", Kind: definfo.Method}}),
			},
			wantNames: map[graph.Qualification]string{graph.ScopeQualified: "(*T).name"},
		},
		{
			// all funcs are at pkg scope
			def: &graph.Def{
				Name: "name",
				Data: defInfo(DefData{DefInfo: definfo.DefInfo{PkgName: "mypkg", Kind: definfo.Func}}),
			},
			wantNames: map[graph.Qualification]string{graph.ScopeQualified: "name"},
		},
		{
			// qualify funcs with pkg
			def: &graph.Def{
				Name: "Name",
				Data: defInfo(DefData{DefInfo: definfo.DefInfo{PkgName: "mypkg", Kind: definfo.Func}}),
			},
			wantNames: map[graph.Qualification]string{graph.DepQualified: "mypkg.Name"},
		},
		{
			// qualify methods with receiver pkg
			def: &graph.Def{
				Name: "Name",
				Data: defInfo(DefData{DefInfo: definfo.DefInfo{Receiver: "*T", PkgName: "mypkg", Kind: definfo.Method}}),
			},
			wantNames: map[graph.Qualification]string{graph.DepQualified: "(*mypkg.T).Name"},
		},
		{
			// qualify pkgs with import path relative to repo root
			def: &graph.Def{
				DefKey: graph.DefKey{Repo: "example.com/foo"},
				Name:   "subpkg",
				Kind:   "package",
				Data: defInfo(DefData{
					PackageImportPath: "example.com/foo/mypkg/subpkg",
					DefInfo: definfo.DefInfo{
						PkgName: "subpkg", Kind: definfo.Package,
						TypeString: "struct {x example.com/foo/mypkg/subpkg.T1; w example.com/foo/mypkg.T2; y example.com/foo/mypkg/subpkg/subsubpkg.T2}",
					},
				}),
			},
			wantNames: map[graph.Qualification]string{graph.RepositoryWideQualified: "mypkg/subpkg"},
			wantTypes: map[graph.Qualification]string{
				graph.RepositoryWideQualified: " struct {x mypkg/subpkg.T1; w mypkg.T2; y mypkg/subpkg/subsubpkg.T2}",
			},
		},
		{
			// qualify funcs with import path
			def: &graph.Def{
				Name: "Name",
				Data: defInfo(DefData{PackageImportPath: "a/b", DefInfo: definfo.DefInfo{PkgName: "x", Kind: definfo.Func}}),
			},
			wantNames: map[graph.Qualification]string{graph.LanguageWideQualified: "a/b.Name"},
		},
		{
			// qualify methods with receiver pkg
			def: &graph.Def{
				Name: "Name",
				Data: defInfo(DefData{PackageImportPath: "a/b", DefInfo: definfo.DefInfo{Receiver: "*T", PkgName: "x", Kind: definfo.Method}}),
			},
			wantNames: map[graph.Qualification]string{graph.LanguageWideQualified: "(*a/b.T).Name"},
		},
		{
			// qualify pkgs with full import path
			def: &graph.Def{
				Name: "x",
				Data: defInfo(DefData{PackageImportPath: "a/b", DefInfo: definfo.DefInfo{PkgName: "x", Kind: definfo.Package}}),
			},
			wantNames: map[graph.Qualification]string{graph.LanguageWideQualified: "a/b"},
		},
	}
	for _, test := range tests {
		sf := newDefFormatter(test.def)
		for qual, want := range test.wantNames {
			name := sf.Name(qual)
			if name != want {
				t.Errorf("%v qual %q: got name %q, want %q", test.def, qual, name, want)
			}
		}
		for qual, want := range test.wantTypes {
			typ := sf.Type(qual)
			if typ != want {
				t.Errorf("%v qual %q: got type %q, want %q", test.def, qual, typ, want)
			}
		}
	}
}
