package vfsutil

import (
	"context"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/sourcegraph/ctxvfs"
)

func testVFS(t *testing.T, fs ctxvfs.FileSystem, want map[string]string) {
	tree, err := ctxvfs.ReadAllFiles(context.Background(), fs, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(tree) != len(want) {
		t.Errorf("got %d files, want %d files", len(tree), len(want))
	}
	for wantFile, wantContents := range want {
		contentsBytes, ok := tree[wantFile]
		if !ok {
			t.Errorf("missing file %s", wantFile)
			continue
		}
		contents := string(contentsBytes)

		if strings.HasSuffix(wantContents, "...") {
			// Allow specifying expected contents with "..." at the
			// end for ease of test creation.
			if len(contents) >= len(wantContents)+3 {
				contents = contents[:len(wantContents)-3] + "..."
			}
		}
		if contents != wantContents {
			t.Errorf("%s: got contents %q, want %q", wantFile, contents, wantContents)
		}
	}
	for file := range tree {
		if _, ok := want[file]; !ok {
			t.Errorf("extra file %s", file)
		}
	}
	if fsLister, ok := fs.(interface {
		ListAllFiles(context.Context) ([]string, error)
	}); ok {
		got, err := fsLister.ListAllFiles(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		wantList := make([]string, 0, len(want))
		for file := range want {
			wantList = append(wantList, strings.TrimPrefix(file, "/"))
		}
		sort.Strings(got)
		sort.Strings(wantList)
		if !reflect.DeepEqual(got, wantList) {
			t.Fatalf("ListAllFiles does not match want:\ngot:  %v\nwant: %v", got, wantList)
		}
	}
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_972(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
