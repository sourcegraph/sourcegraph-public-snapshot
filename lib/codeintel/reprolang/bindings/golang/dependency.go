package golang

import "github.com/sourcegraph/sourcegraph/lib/codeintel/lsiftyped"

type Dependency struct {
	Package *lsiftyped.Package
	Sources []*lsiftyped.SourceFile
}

type reproDependency struct {
	Package *lsiftyped.Package
	Sources []*reproSourceFile
}
