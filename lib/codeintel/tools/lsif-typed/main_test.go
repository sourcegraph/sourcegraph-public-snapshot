package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	lsiftypedtesting "github.com/sourcegraph/sourcegraph/lib/codeintel/lsif-typed-testing"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/protocol/reader"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif_typed"
	reproLang "github.com/sourcegraph/sourcegraph/lib/codeintel/repro_lang/bindings/golang"
	"github.com/stretchr/testify/require"
)

// TestLsifTypedSnapshots runs all the snapshot tests from the "snapshot-input" directory.
func TestLsifTypedSnapshots(t *testing.T) {
	lsiftypedtesting.SnapshotTest(t, func(inputDirectory, outputDirectory string, sources []*lsif_typed.SourceFile) []*lsif_typed.SourceFile {
		testName := filepath.Base(inputDirectory)
		var dependencies []*reproLang.Dependency
		rootDirectory := filepath.Dir(inputDirectory)
		dirs, err := os.ReadDir(rootDirectory)
		require.Nil(t, err)
		for _, dir := range dirs {
			if !dir.IsDir() {
				continue
			}
			if dir.Name() == testName {
				continue
			}
			dependencyRoot := filepath.Join(rootDirectory, dir.Name())
			dependencySources, err := lsif_typed.NewSourcesFromDirectory(dependencyRoot)
			require.Nil(t, err)
			dependencies = append(dependencies, &reproLang.Dependency{
				Package: &lsif_typed.Package{
					Manager: "repro_manager",
					Name:    dir.Name(),
					Version: "1.0.0",
				},
				Sources: dependencySources,
			})
		}
		index, err := reproLang.Index("file:/"+inputDirectory, testName, sources, dependencies)
		require.Nil(t, err)
		symbolFormatter := lsif_typed.DescriptorOnlyFormatter
		symbolFormatter.IncludePackageName = func(name string) bool { return name != testName }
		snapshots, err := lsiftypedtesting.FormatSnapshots(index, "#", symbolFormatter)
		require.Nil(t, err)
		index.Metadata.ProjectRoot = "file:/root"
		lsif, err := reader.ConvertTypedIndexToGraphIndex(index)
		require.Nil(t, err)
		var obtained bytes.Buffer
		err = reader.WriteNDJSON(reader.ElementsToJsonElements(lsif), &obtained)
		require.Nil(t, err)
		snapshots = append(snapshots, lsif_typed.NewSourceFile(
			filepath.Join(outputDirectory, "dump.lsif"),
			"dump.lsif",
			obtained.String(),
		))
		return snapshots
	})
}
