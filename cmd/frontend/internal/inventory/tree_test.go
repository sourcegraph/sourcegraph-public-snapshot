package inventory

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/vcs/util"
)

func TestContext_Tree(t *testing.T) {
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
			case "":
				return []os.FileInfo{
					&util.FileInfo{Name_: "a", Mode_: os.ModeDir},
					&util.FileInfo{Name_: "b.go", Size_: 12},
				}, nil
			case "a":
				return []os.FileInfo{&util.FileInfo{Name_: "a/c.m", Size_: 24}}, nil
			default:
				panic("unhandled mock ReadTree " + path)
			}
		},
		NewFileReader: func(ctx context.Context, path string) (io.ReadCloser, error) {
			newFileReaderCalls = append(newFileReaderCalls, path)
			var data []byte
			switch path {
			case "b.go":
				data = []byte("package main")
			case "a/c.m":
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

	inv, err := c.Tree(context.Background(), &util.FileInfo{Name_: "", Mode_: os.ModeDir})
	if err != nil {
		t.Fatal(err)
	}
	if want := (Inventory{
		Languages: []Lang{
			{Name: "Objective-C", TotalBytes: 24, TotalLines: 1},
			{Name: "Go", TotalBytes: 12, TotalLines: 1},
		},
	}); !reflect.DeepEqual(inv, want) {
		t.Errorf("got  %#v\nwant %#v", inv, want)
	}

	// Check that our mocks were called as expected.
	if want := []string{"", "a"}; !reflect.DeepEqual(readTreeCalls, want) {
		t.Errorf("ReadTree calls: got %q, want %q", readTreeCalls, want)
	}
	if want := []string{
		// We need to read all files to get line counts
		"a/c.m",
		"b.go",
	}; !reflect.DeepEqual(newFileReaderCalls, want) {
		t.Errorf("GetFileReader calls: got %q, want %q", newFileReaderCalls, want)
	}
	if want := []string{"", "a"}; !reflect.DeepEqual(cacheGetCalls, want) {
		t.Errorf("CacheGet calls: got %q, want %q", cacheGetCalls, want)
	}
	if want := map[string]Inventory{
		"": {
			Languages: []Lang{
				{Name: "Objective-C", TotalBytes: 24, TotalLines: 1},
				{Name: "Go", TotalBytes: 12, TotalLines: 1},
			},
		},
		"a": {
			Languages: []Lang{
				{Name: "Objective-C", TotalBytes: 24, TotalLines: 1},
			},
		},
	}; !reflect.DeepEqual(cacheSetCalls, want) {
		t.Errorf("CacheGet calls: got %+v, want %+v", cacheSetCalls, want)
	}
}
