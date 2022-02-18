package golang

import "github.com/sourcegraph/sourcegraph/lib/codeintel/lsif_typed"

type Dependency struct {
	Package *lsif_typed.Package
	Sources []*lsif_typed.SourceFile
}

type reproDependency struct {
	Package *lsif_typed.Package
	Sources []*reproSourceFile
}
