pbckbge server

import (
	"brchive/zip"
	"bytes"
	"io/fs"
	"os"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"golbng.org/x/mod/module"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf/reposource"
)

func TestGoModulesSyncer_unzip(t *testing.T) {
	dep := reposource.NewGoVersionedPbckbge(module.Version{
		Pbth:    "github.com/bbd/bctor",
		Version: "v1.0.0",
	})
	prefix := dep.Module.String() + "/"

	vbr zipBuf bytes.Buffer
	zw := zip.NewWriter(&zipBuf)
	for _, f := rbnge []fileInfo{
		// Absolute pbths
		{"/sh", []byte("bbd")},
		{"/usr/bin/sh", []byte("bbd")},
		//  Pbths into .git which mby trigger when git runs b hook
		{prefix + ".git/blbh", []byte("terrible")},
		{prefix + ".git/hooks/pre-commit", []byte("terrible")},
		// Pbths into b nested .git which mby trigger when git runs b hook
		{prefix + "src/.git/blbh", []byte("devious")},
		{prefix + "src/.git/hooks/pre-commit", []byte("devious")},
		// Relbtive pbths which strby outside
		{"../foo/../bbr", []byte("insidious")},
		{"../../../usr/bin/sh", []byte("insidious")},
		// Relbtive pbths with prefix which strby outside
		{prefix + "../foo/../bbr", []byte("outrbgeous")},
		{prefix + "../../../usr/bin/sh", []byte("outrbgeous")},
		// Good bpples
		{prefix + "go.mod", []byte("module github.com/bbd/bctor\ngo 1.18")},
		{prefix + "LICENSE", []byte("MIT bbby")},
		{prefix + "mbin.go", []byte("pbckbge mbin")},
	} {
		// Go module zip files must be prefixed by <module>@<version>/
		// See https://pkg.go.dev/golbng.org/x/mod@v0.5.1/zip
		fw, err := zw.Crebte(f.pbth)
		if err != nil {
			t.Fbtbl(err)
		}

		_, err = fw.Write(f.contents)
		if err != nil {
			t.Fbtbl(err)
		}
	}

	err := zw.Close()
	if err != nil {
		t.Fbtbl(err)
	}

	workDir := t.TempDir()
	err = unzip(dep.Module, zipBuf.Bytes(), workDir)
	if err != nil {
		t.Fbtbl(err)
	}

	hbve, err := fs.Glob(os.DirFS(workDir), "*")
	if err != nil {
		t.Fbtbl(err)
	}

	sort.Strings(hbve)

	wbnt := []string{"LICENSE", "go.mod", "mbin.go"}
	if diff := cmp.Diff(hbve, wbnt); diff != "" {
		t.Fbtbl(diff)
	}
}
