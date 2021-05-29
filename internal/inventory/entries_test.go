package inventory

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/vcs/util"
)

func TestContext_Entries(t *testing.T) {
	var (
		readTreeCalls      []string
		newFileReaderCalls []string
		cacheGetCalls      []string
		cacheSetCalls      = map[string]Inventory{}
	)
	c := Context{
		ReadTree: func(ctx context.Context, path string) ([]os.FileInfo, error) {
			readTreeCalls = append(readTreeCalls, path)
			switch path {
			case "d":
				return []os.FileInfo{
					&util.FileInfo{Name_: "d/a", Mode_: os.ModeDir},
					&util.FileInfo{Name_: "d/b.go", Size_: 12},
				}, nil
			case "d/a":
				return []os.FileInfo{&util.FileInfo{Name_: "d/a/c.m", Size_: 24}}, nil
			default:
				panic("unhandled mock ReadTree " + path)
			}
		},
		NewFileReader: func(ctx context.Context, path string) (io.ReadCloser, error) {
			newFileReaderCalls = append(newFileReaderCalls, path)
			var data []byte
			switch path {
			case "f.go":
				data = []byte("package f")
			case "d/b.go":
				data = []byte("package main")
			case "d/a/c.m":
				data = []byte("@interface X:NSObject {}")
			default:
				panic("unhandled mock ReadFile " + path)
			}
			return ioutil.NopCloser(bytes.NewReader(data)), nil
		},
		CacheGet: func(e os.FileInfo) (Inventory, bool) {
			cacheGetCalls = append(cacheGetCalls, e.Name())
			return Inventory{}, false
		},
		CacheSet: func(e os.FileInfo, inv Inventory) {
			if _, ok := cacheSetCalls[e.Name()]; ok {
				t.Fatalf("already stored %q in cache", e.Name())
			}
			cacheSetCalls[e.Name()] = inv
		},
	}

	inv, err := c.Entries(context.Background(),
		&util.FileInfo{Name_: "d", Mode_: os.ModeDir},
		&util.FileInfo{Name_: "f.go", Mode_: 0, Size_: 1 /* HACK to force read */},
	)
	if err != nil {
		t.Fatal(err)
	}
	if want := (Inventory{
		Languages: []Lang{
			{Name: "Go", TotalBytes: 21, TotalLines: 2},
			{Name: "Objective-C", TotalBytes: 24, TotalLines: 1},
		},
	}); !reflect.DeepEqual(inv, want) {
		t.Fatalf("got  %#v\nwant %#v", inv, want)
	}

	// Check that our mocks were called as expected.
	if want := []string{"d", "d/a"}; !reflect.DeepEqual(readTreeCalls, want) {
		t.Errorf("ReadTree calls: got %q, want %q", readTreeCalls, want)
	}
	if want := []string{
		// We need to read all files to get line counts
		"d/a/c.m",
		"d/b.go",
		"f.go",
	}; !reflect.DeepEqual(newFileReaderCalls, want) {
		t.Errorf("GetFileReader calls: got %q, want %q", newFileReaderCalls, want)
	}
	if want := []string{"d", "d/a", "f.go"}; !reflect.DeepEqual(cacheGetCalls, want) {
		t.Errorf("CacheGet calls: got %q, want %q", cacheGetCalls, want)
	}
	want := map[string]Inventory{
		"d": {
			Languages: []Lang{
				{Name: "Objective-C", TotalBytes: 24, TotalLines: 1},
				{Name: "Go", TotalBytes: 12, TotalLines: 1},
			},
		},
		"d/a": {
			Languages: []Lang{
				{Name: "Objective-C", TotalBytes: 24, TotalLines: 1},
			},
		},
		"f.go": {
			Languages: []Lang{
				{Name: "Go", TotalBytes: 9, TotalLines: 1},
			},
		},
	}
	if diff := cmp.Diff(want, cacheSetCalls); diff != "" {
		t.Error(diff)
	}
}
