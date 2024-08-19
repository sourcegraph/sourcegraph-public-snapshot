package internal

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/packages"
)

// GoPackage is the scope of *packages.GoPackage that is needed for cff.
type GoPackage struct {
	PkgPath         string
	Errors          []packages.Error
	CompiledGoFiles []string
	Syntax          []*ast.File
	Types           *types.Package
	TypesInfo       *types.Info
}

func NewPackage(pkg *packages.Package) *GoPackage {
	return &GoPackage{
		PkgPath:         pkg.PkgPath,
		Errors:          pkg.Errors,
		CompiledGoFiles: pkg.CompiledGoFiles,
		Syntax:          pkg.Syntax,
		Types:           pkg.Types,
		TypesInfo:       pkg.TypesInfo,
	}
}
