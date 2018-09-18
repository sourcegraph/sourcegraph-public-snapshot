package server

import (
	"fmt"
	"go/build"
	"reflect"
	"sort"
	"testing"

	sort8 "go4.org/sort"
)

func TestUnvendoredPath(t *testing.T) {
	tests := map[string]string{
		// valid cases:
		"github.com/slimsag/foobar/vendor/github.com/gorilla/mux": "github.com/gorilla/mux",
		"/a/vendor/github.com/gorilla/mux":                        "github.com/gorilla/mux",

		// invalid cases:
		"vendor/github.com/gorilla/mux":                 "vendor/github.com/gorilla/mux",
		"/a/vendor/weird/vendor/github.com/gorilla/mux": "weird/vendor/github.com/gorilla/mux",
		"":         "",
		"vendor":   "vendor",
		"/vendor":  "/vendor",
		"vendor/":  "vendor/",
		"/vendor/": "",
	}
	for input, want := range tests {
		got := unvendoredPath(input)
		if got != want {
			t.Logf("got  %q", got)
			t.Logf("want %q", want)
			t.Fail()
		}
	}
}

func TestDepsReferences(t *testing.T) {
	importPackage := func(path, srcDir string, mode build.ImportMode) (*build.Package, error) {
		switch path {
		case "a":
			return &build.Package{ImportPath: "a", Imports: []string{"b", "c"}, Dir: "/a"}, nil
		case "b":
			return &build.Package{ImportPath: "b", Imports: []string{"d"}, Dir: "/b"}, nil
		case "c":
			return &build.Package{ImportPath: "c", Imports: []string{"d"}, Dir: "/c"}, nil
		case "d":
			return &build.Package{ImportPath: "d", Imports: []string{"e", "f"}, Dir: "/d"}, nil
		case "e":
			return &build.Package{ImportPath: "e", Imports: []string{"f", "e"}, Dir: "/e"}, nil
		case "f":
			return &build.Package{ImportPath: "f", Imports: []string{"d"}, Dir: "/f"}, nil
		default:
			return nil, fmt.Errorf("package not found: %q", path)
		}
	}

	dc := newDepCache()
	dc.collectReferences = true
	if err := doDeps(&build.Package{Imports: []string{"a"}, Dir: "/"}, 0, dc, importPackage); err != nil {
		t.Error(err)
	}
	referencesTest(t, dc, nil)
}

func TestDepCacheReferences(t *testing.T) {
	pkg := map[string]*build.Package{
		"":  &build.Package{Imports: []string{"a"}, Dir: "/"},
		"a": &build.Package{ImportPath: "a", Imports: []string{"b", "c"}, Dir: "/a"},
		"b": &build.Package{ImportPath: "b", Imports: []string{"d"}, Dir: "/b"},
		"c": &build.Package{ImportPath: "c", Imports: []string{"d"}, Dir: "/c"},
		"d": &build.Package{ImportPath: "d", Imports: []string{"e", "f"}, Dir: "/d"},
		"e": &build.Package{ImportPath: "e", Imports: []string{"f", "e"}, Dir: "/e"},
		"f": &build.Package{ImportPath: "f", Imports: []string{"d"}, Dir: "/f"},
	}
	dc := newDepCache()
	dc.entryPackageDirs = append(dc.entryPackageDirs, pkg[""].Dir)
	dc.seen = map[string][]importRecord{
		"/": []importRecord{{pkg: pkg[""], imports: pkg["a"]}},
		"/a": []importRecord{
			{pkg: pkg["a"], imports: pkg["b"]},
			{pkg: pkg["a"], imports: pkg["c"]},
		},
		"/b": []importRecord{{pkg: pkg["b"], imports: pkg["d"]}},
		"/c": []importRecord{{pkg: pkg["c"], imports: pkg["d"]}},
		"/d": []importRecord{
			{pkg: pkg["d"], imports: pkg["f"]},
			{pkg: pkg["d"], imports: pkg["e"]},
		},
		"/e": []importRecord{
			{pkg: pkg["e"], imports: pkg["e"]},
			{pkg: pkg["e"], imports: pkg["f"]},
		},
		"/f": []importRecord{{pkg: pkg["f"], imports: pkg["d"]}},
	}
	referencesTest(t, dc, nil)
}

func referencesTest(t *testing.T, dc *depCache, want map[string][]goDependencyReference) {
	references := map[string][]goDependencyReference{}
	emitRef := func(path string, r goDependencyReference) {
		references[path] = append(references[path], r)
	}
	dc.references(emitRef, 100)

	r := func(abs string, depth int) goDependencyReference {
		// TODO(slimsag): write unit tests for vendor / Pkg field
		return goDependencyReference{absolute: abs, pkg: abs, depth: depth}
	}
	if want == nil {
		want = map[string][]goDependencyReference{
			// Import graph:
			//
			//       '/' (root)
			//        |
			//        a
			//        |\
			//        b c
			//         \|
			//    .>.   d <<<<<<.
			//    |  \ / \    | |
			//    .<< e   f >>^ |
			//        |         |
			//        f >>>>>>>>^
			//
			"/":  []goDependencyReference{r("a", 0), r("b", 1), r("c", 1), r("d", 2), r("e", 3), r("f", 3)},
			"/a": []goDependencyReference{r("b", 0), r("c", 0), r("d", 1), r("e", 2), r("f", 2)},
			"/b": []goDependencyReference{r("d", 0), r("e", 1), r("f", 1)},
			"/c": []goDependencyReference{r("d", 0), r("e", 1), r("f", 1)},
			"/d": []goDependencyReference{r("e", 0), r("f", 0)},
			"/e": []goDependencyReference{r("f", 0)},
		}
	}
	for _, s := range references {
		sort8.Slice(s, func(i, j int) bool {
			if s[i].depth != s[j].depth {
				return s[i].depth < s[j].depth
			}
			return s[i].absolute < s[j].absolute
		})
	}
	if !reflect.DeepEqual(references, want) {
		t.Logf("got:")
		for k, v := range references {
			t.Logf("\t%q:%v", k, v)
		}
		t.Logf("want:")
		for k, v := range want {
			t.Logf("\t%q:%v", k, v)
		}

		for k, v := range references {
			if v2, ok := want[k]; ok {
				if !reflect.DeepEqual(v, v2) {
					t.Logf("\n  got  %q:%v\n  want %q:%v", k, v, k, v2)
				}
			}
		}
		t.FailNow()
	}
}

func TestDepCacheReferencesStableOrder(t *testing.T) {
	pkg := map[string]*build.Package{
		"sync":                    &build.Package{ImportPath: "sync", Dir: "/goroot/src/sync", Imports: []string{"internal/race", "runtime", "sync/atomic", "unsafe"}},
		"unsafe":                  &build.Package{ImportPath: "unsafe", Dir: "/goroot/src/unsafe", Imports: []string(nil)},
		"sync/atomic":             &build.Package{ImportPath: "sync/atomic", Dir: "/goroot/src/sync/atomic", Imports: []string{"unsafe"}},
		"runtime/internal/atomic": &build.Package{ImportPath: "runtime/internal/atomic", Dir: "/goroot/src/runtime/internal/atomic", Imports: []string{"unsafe"}},
		"runtime/internal/sys":    &build.Package{ImportPath: "runtime/internal/sys", Dir: "/goroot/src/runtime/internal/sys", Imports: []string(nil)},
		"internal/race":           &build.Package{ImportPath: "internal/race", Dir: "/goroot/src/internal/race", Imports: []string{"unsafe"}},
		"runtime":                 &build.Package{ImportPath: "runtime", Dir: "/goroot/src/runtime", Imports: []string{"runtime/internal/atomic", "runtime/internal/sys", "unsafe"}},
	}

	importPackage := func(path, srcDir string, mode build.ImportMode) (*build.Package, error) {
		p, ok := pkg[path]
		if !ok {
			return nil, fmt.Errorf("package not found: %q", path)
		}
		return p, nil
	}

	wantRecords := map[string][]importRecord{
		"/goroot/src/runtime/internal/atomic": []importRecord{{pkg: pkg["runtime/internal/atomic"], imports: pkg["unsafe"]}},
		"/goroot/src/internal/race":           []importRecord{{pkg: pkg["internal/race"], imports: pkg["unsafe"]}},
		"/goroot/src/sync/atomic":             []importRecord{{pkg: pkg["sync/atomic"], imports: pkg["unsafe"]}},
		"/goroot/src/sync": []importRecord{
			{pkg: pkg["sync"], imports: pkg["internal/race"]},
			{pkg: pkg["sync"], imports: pkg["runtime"]},
			{pkg: pkg["sync"], imports: pkg["sync/atomic"]},
			{pkg: pkg["sync"], imports: pkg["unsafe"]},
		},
		"/goroot/src/runtime": []importRecord{
			{pkg: pkg["runtime"], imports: pkg["unsafe"]},
			{pkg: pkg["runtime"], imports: pkg["runtime/internal/atomic"]},
			{pkg: pkg["runtime"], imports: pkg["runtime/internal/sys"]},
		},
		"/goroot/src/unsafe":               []importRecord{},
		"/goroot/src/runtime/internal/sys": []importRecord{},
	}

	r := func(abs string, depth int) goDependencyReference {
		return goDependencyReference{absolute: abs, pkg: abs, depth: depth}
	}
	wantRefs := map[string][]goDependencyReference{
		"/goroot/src/runtime":                 []goDependencyReference{r("runtime/internal/atomic", 0), r("runtime/internal/sys", 0), r("unsafe", 0)},
		"/goroot/src/sync":                    []goDependencyReference{r("internal/race", 0), r("runtime", 0), r("sync/atomic", 0), r("unsafe", 0), r("runtime/internal/atomic", 1), r("runtime/internal/sys", 1)},
		"/goroot/src/internal/race":           []goDependencyReference{r("unsafe", 0)},
		"/goroot/src/sync/atomic":             []goDependencyReference{r("unsafe", 0)},
		"/goroot/src/runtime/internal/atomic": []goDependencyReference{r("unsafe", 0)},
	}

	for i := 0; i < 50; i++ {
		dc := newDepCache()
		dc.collectReferences = true
		if err := doDeps(pkg["sync"], 0, dc, importPackage); err != nil {
			t.Error(err)
			return
		}

		for _, s := range dc.seen {
			sort.Sort(sortedImportRecord(s))
		}
		for _, s := range wantRecords {
			sort.Sort(sortedImportRecord(s))
		}

		if !reflect.DeepEqual(dc.seen, wantRecords) {
			t.Log("test run", i)
			t.Logf("got records:")
			for k, v := range dc.seen {
				t.Logf("\t%q:%+v", k, v)
			}
			t.Logf("want records:")
			for k, v := range wantRecords {
				t.Logf("\t%q:%+v", k, v)
			}

			for k, v := range dc.seen {
				if v2, ok := wantRecords[k]; ok {
					if !reflect.DeepEqual(v, v2) {
						t.Logf("\n  got  %q:%+v\n  want %q:%+v", k, v, k, v2)
					}
				}
			}
			t.FailNow()
		}

		referencesTest(t, dc, wantRefs)
	}
}
