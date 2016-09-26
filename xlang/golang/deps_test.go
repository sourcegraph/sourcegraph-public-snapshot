package golang

import (
	"fmt"
	"go/build"
	"reflect"
	"testing"
)

func TestImportDeps(t *testing.T) {
	imported := map[string]struct{}{}
	importPackage := func(path, srcDir string, mode build.ImportMode) (*build.Package, error) {
		imported[path] = struct{}{}
		switch path {
		case "a":
			return &build.Package{Imports: []string{"b", "c"}}, nil
		case "b":
			return &build.Package{}, nil
		case "c":
			return &build.Package{Imports: []string{"d"}}, nil
		case "d":
			return &build.Package{Imports: []string{"e"}}, nil
		case "e":
			return &build.Package{Imports: []string{"f"}}, nil
		case "f":
			return &build.Package{}, nil
		default:
			return nil, fmt.Errorf("package not found: %q", path)
		}
	}
	if err := doDeps(&build.Package{Imports: []string{"a"}}, 0, importPackage); err != nil {
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
