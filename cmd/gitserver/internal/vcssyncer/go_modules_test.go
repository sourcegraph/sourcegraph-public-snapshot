package vcssyncer

import (
	"archive/zip"
	"bytes"
	"io/fs"
	"os"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/mod/module"

	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
)

func TestGoModulesSyncer_unzip(t *testing.T) {
	dep := reposource.NewGoVersionedPackage(module.Version{
		Path:    "github.com/bad/actor",
		Version: "v1.0.0",
	})
	prefix := dep.Module.String() + "/"

	var zipBuf bytes.Buffer
	zw := zip.NewWriter(&zipBuf)
	for _, f := range []fileInfo{
		// Absolute paths
		{"/sh", []byte("bad")},
		{"/usr/bin/sh", []byte("bad")},
		//  Paths into .git which may trigger when git runs a hook
		{prefix + ".git/blah", []byte("terrible")},
		{prefix + ".git/hooks/pre-commit", []byte("terrible")},
		// Paths into a nested .git which may trigger when git runs a hook
		{prefix + "src/.git/blah", []byte("devious")},
		{prefix + "src/.git/hooks/pre-commit", []byte("devious")},
		// Relative paths which stray outside
		{"../foo/../bar", []byte("insidious")},
		{"../../../usr/bin/sh", []byte("insidious")},
		// Relative paths with prefix which stray outside
		{prefix + "../foo/../bar", []byte("outrageous")},
		{prefix + "../../../usr/bin/sh", []byte("outrageous")},
		// Good apples
		{prefix + "go.mod", []byte("module github.com/bad/actor\ngo 1.18")},
		{prefix + "LICENSE", []byte("MIT baby")},
		{prefix + "main.go", []byte("package main")},
	} {
		// Go module zip files must be prefixed by <module>@<version>/
		// See https://pkg.go.dev/golang.org/x/mod@v0.5.1/zip
		fw, err := zw.Create(f.path)
		if err != nil {
			t.Fatal(err)
		}

		_, err = fw.Write(f.contents)
		if err != nil {
			t.Fatal(err)
		}
	}

	err := zw.Close()
	if err != nil {
		t.Fatal(err)
	}

	workDir := t.TempDir()
	err = unzip(dep.Module, bytes.NewReader(zipBuf.Bytes()), workDir, workDir)
	if err != nil {
		t.Fatal(err)
	}

	have, err := fs.Glob(os.DirFS(workDir), "*")
	if err != nil {
		t.Fatal(err)
	}

	sort.Strings(have)

	want := []string{"LICENSE", "go.mod", "main.go"}
	if diff := cmp.Diff(have, want); diff != "" {
		t.Fatal(diff)
	}
}
