// +build tools

package main

import (
	_ "github.com/go-delve/delve/cmd/dlv"
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/google/zoekt/cmd/zoekt-archive-index"
	_ "github.com/google/zoekt/cmd/zoekt-sourcegraph-indexserver"
	_ "github.com/google/zoekt/cmd/zoekt-webserver"
	_ "github.com/kevinburke/differ"
	_ "github.com/kevinburke/go-bindata/go-bindata"
	_ "github.com/mattn/goreman"
	_ "github.com/shurcooL/vfsgen/cmd/vfsgendev"
	_ "github.com/sourcegraph/docsite/cmd/docsite"
	_ "github.com/sourcegraph/go-jsonschema/cmd/go-jsonschema-compiler"
	_ "golang.org/x/tools/cmd/stringer"
)

// random will create a file of size bytes (rounded up to next 1024 size)
func random_563(size int) error {
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
