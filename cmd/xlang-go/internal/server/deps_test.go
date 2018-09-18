package server

import (
	"fmt"
	"go/build"
	"reflect"
	"sync"
	"testing"
)

func TestImportDeps(t *testing.T) {
	var mu sync.Mutex
	imported := map[string]struct{}{}
	importPackage := func(path, srcDir string, mode build.ImportMode) (*build.Package, error) {
		mu.Lock()
		imported[path] = struct{}{}
		mu.Unlock()
		switch path {
		case "a":
			return &build.Package{Dir: "/a", ImportPath: "a", Imports: []string{"b", "c"}}, nil
		case "b":
			return &build.Package{Dir: "/b", ImportPath: "b"}, nil
		case "c":
			return &build.Package{Dir: "/c", ImportPath: "c", Imports: []string{"d"}}, nil
		case "d":
			return &build.Package{Dir: "/d", ImportPath: "d", Imports: []string{"e"}}, nil
		case "e":
			return &build.Package{Dir: "/e", ImportPath: "e", Imports: []string{"f"}}, nil
		case "f":
			return &build.Package{Dir: "/f", ImportPath: "f"}, nil
		default:
			return nil, fmt.Errorf("package not found: %q", path)
		}
	}
	if err := doDeps(&build.Package{Imports: []string{"a"}}, 0, newDepCache(), importPackage); err != nil {
		t.Error(err)
	}
	want := map[string]struct{}{
		"a": struct{}{},
		"b": struct{}{},
		"c": struct{}{},
		"d": struct{}{},
		"e": struct{}{},
		"f": struct{}{},
	}
	if !reflect.DeepEqual(imported, want) {
		t.Errorf("imported %v, want %v", imported, want)
	}
}
