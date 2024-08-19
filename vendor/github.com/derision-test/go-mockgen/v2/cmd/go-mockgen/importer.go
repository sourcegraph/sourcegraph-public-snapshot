package main

import (
	"fmt"
	"go/token"
	"go/types"
	"os"
	"strings"

	"golang.org/x/tools/go/gcexportdata"
)

type importer struct {
	importToArchive map[string]string
	stdlibRoot      string
	fset            *token.FileSet
	imports         map[string]*types.Package
}

func newImporter(fset *token.FileSet, archives []archive, root string) (types.Importer, error) {
	imp := &importer{
		importToArchive: make(map[string]string, len(archives)),
		fset:            fset,
		imports:         make(map[string]*types.Package),
		stdlibRoot:      root,
	}

	for _, archive := range archives {
		imp.importToArchive[archive.ImportMap] = archive.File
	}
	return imp, nil
}

func (i *importer) Import(path string) (*types.Package, error) {
	if pkg, ok := i.imports[path]; ok && pkg.Complete() {
		return pkg, nil
	}

	if path == "unsafe" {
		// Special case: go/types has pre-defined type information for unsafe.
		// See https://github.com/golang/go/issues/13882.
		return types.Unsafe, nil
	}

	if isStdlibImport(path) {
		archiveFile := fmt.Sprintf("%v/%v.a", i.stdlibRoot, path)
		return i.readArchive(archiveFile, path)
	}

	if archive, ok := i.importToArchive[path]; ok {
		return i.readArchive(archive, path)
	}
	return nil, fmt.Errorf("package %q not found in read archives: please double check dependencies for the go-mockgen bazel rule", path)
}

func (i *importer) readArchive(archiveFile, path string) (p *types.Package, err error) {
	f, err := os.Open(archiveFile)
	if err != nil {
		return nil, err
	}
	defer func() { f.Close() }()

	r, err := gcexportdata.NewReader(f)
	if err != nil {
		return nil, err
	}

	return gcexportdata.Read(r, i.fset, i.imports, path)
}

func isStdlibImport(path string) bool {
	if i := strings.IndexByte(path, '/'); i >= 0 {
		path = path[:i]
	}

	// If the prefix of the import path contains a ".", it should be considered
	// to be a external package (not part of Go standard lib).
	return !strings.Contains(path, ".")
}
