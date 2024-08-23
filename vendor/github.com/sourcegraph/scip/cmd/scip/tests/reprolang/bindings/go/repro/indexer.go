package repro

import (
	"context"

	"github.com/sourcegraph/scip/bindings/go/scip"
)

// Index returns an SCIP index for all of the provided source files, which should use the syntax of
// the "reprolang" programming language. Search for files with the `*.repro` file extension to see examples
// of how reprolang programs looks like. Search for "grammar.js" to see the tree-sitter grammar of the reprolang syntax.
func Index(
	projectRoot, packageName string,
	sources []*scip.SourceFile,
	dependencies []*Dependency,
) (*scip.Index, error) {
	index := &scip.Index{
		Metadata: &scip.Metadata{
			Version: 0,
			ToolInfo: &scip.ToolInfo{
				Name:      "reprolang",
				Version:   "1.0.0",
				Arguments: []string{"arg1", "arg2"},
			},
			ProjectRoot:          projectRoot,
			TextDocumentEncoding: scip.TextEncoding_UTF8,
		},
		Documents:       nil,
		ExternalSymbols: nil,
	}

	ctx := &reproContext{
		globalScope: newScope(),
		pkg: &scip.Package{
			Manager: "repro-manager",
			Name:    packageName,
			Version: "1.0.0",
		},
	}

	// Phase 1: parse sources
	var reproSources []*reproSourceFile
	for _, source := range sources {
		reproSource, err := parseSourceFile(context.Background(), source)
		if err != nil {
			return nil, err
		}
		reproSources = append(reproSources, reproSource)
	}
	var reproDependencies []*reproDependency
	for _, dependency := range dependencies {
		dep := &reproDependency{Package: dependency.Package}
		reproDependencies = append(reproDependencies, dep)
		for _, source := range dependency.Sources {
			reproSource, err := parseSourceFile(context.Background(), source)
			if err != nil {
				return nil, err
			}
			dep.Sources = append(dep.Sources, reproSource)
		}
	}

	// Phase 2: resolve names for definitions
	for _, dependency := range reproDependencies {
		dependency.enterGlobalDefinitions(ctx)
	}
	for _, file := range reproSources {
		file.enterDefinitions(ctx)
	}

	// Phase 3: resolve names for references
	for _, file := range reproSources {
		file.resolveReferences(ctx)
	}

	// Phase 4: emit SCIP
	for _, file := range reproSources {
		scipDocument := &scip.Document{
			RelativePath: file.Source.RelativePath,
			Occurrences:  file.occurrences(),
			Symbols:      file.symbols(),
		}
		index.Documents = append(index.Documents, scipDocument)
	}

	return index, nil
}

type reproContext struct {
	globalScope *reproScope
	pkg         *scip.Package
}

type reproScope struct {
	names map[string]string
}

func newScope() *reproScope {
	return &reproScope{names: map[string]string{}}
}
