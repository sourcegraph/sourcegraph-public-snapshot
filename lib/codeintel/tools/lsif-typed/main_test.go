package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/protocol/reader"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsiftyped"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsiftyped/testutil"
	reproLang "github.com/sourcegraph/sourcegraph/lib/codeintel/reprolang/bindings/golang"
	"github.com/stretchr/testify/require"
)

// TestLsifTypedSnapshots runs all the snapshot tests from the "snapshot-input" directory.
func TestLsifTypedSnapshots(t *testing.T) {
	testutil.SnapshotTest(t, func(inputDirectory, outputDirectory string, sources []*lsiftyped.SourceFile) []*lsiftyped.SourceFile {
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
			dependencySources, err := lsiftyped.NewSourcesFromDirectory(dependencyRoot)
			require.Nil(t, err)
			dependencies = append(dependencies, &reproLang.Dependency{
				Package: &lsiftyped.Package{
					Manager: "repro_manager",
					Name:    dir.Name(),
					Version: "1.0.0",
				},
				Sources: dependencySources,
			})
		}
		index, err := reproLang.Index("file:/"+inputDirectory, testName, sources, dependencies)
		require.Nil(t, err)
		symbolFormatter := lsiftyped.DescriptorOnlyFormatter
		symbolFormatter.IncludePackageName = func(name string) bool { return name != testName }
		snapshots, err := testutil.FormatSnapshots(index, "#", symbolFormatter)
		require.Nil(t, err)
		index.Metadata.ProjectRoot = "file:/root"
		lsif, err := reader.ConvertTypedIndexToGraphIndex(index)
		require.Nil(t, err)
		var obtained bytes.Buffer
		err = reader.WriteNDJSON(reader.ElementsToJsonElements(lsif), &obtained)
		require.Nil(t, err)
		snapshots = append(snapshots, lsiftyped.NewSourceFile(
			filepath.Join(outputDirectory, "dump.lsif"),
			"dump.lsif",
			obtained.String(),
		))
		return snapshots
	})
}
