package repro

import "github.com/sourcegraph/scip/bindings/go/scip"

type Dependency struct {
	Package *scip.Package
	Sources []*scip.SourceFile
}

type reproDependency struct {
	Package *scip.Package
	Sources []*reproSourceFile
}
