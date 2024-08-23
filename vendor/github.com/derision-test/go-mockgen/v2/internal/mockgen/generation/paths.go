package generation

import (
	gotypes "go/types"
	"strings"

	"github.com/dave/jennifer/jen"
)

func generateQualifiedName(t *gotypes.Named, importPath, outputImportPath string) *jen.Statement {
	name := t.Obj().Name()

	if t.Obj().Pkg() == nil {
		return jen.Id(name)
	}

	if path := t.Obj().Pkg().Path(); path != "" {
		return jen.Qual(sanitizeImportPath(path, outputImportPath), name)
	}

	return jen.Qual(sanitizeImportPath(importPath, outputImportPath), name)
}

func sanitizeImportPath(path, outputImportPath string) string {
	path = stripVendor(path)
	if path == outputImportPath {
		return ""
	}

	return path
}

func stripVendor(path string) string {
	parts := strings.Split(path, "/vendor/")
	return parts[len(parts)-1]
}
