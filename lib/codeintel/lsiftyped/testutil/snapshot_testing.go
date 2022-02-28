package testutil

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsiftyped"
)

var updateLsifSnapshots = flag.Bool("update-lsif-snapshots", false, "update LSIF snapshots files")

type indexFunction = func(inputDirectory, outputDirectory string, sources []*lsiftyped.SourceFile) []*lsiftyped.SourceFile

func SnapshotTest(t *testing.T, indexFunction indexFunction) {
	cwd, err := os.Getwd()
	require.Nil(t, err)
	inputDirectory := filepath.Join(cwd, "snapshots-input")
	outputDirectory := filepath.Join(cwd, "snapshots-output")
	SnapshotTestDirectories(t, inputDirectory, outputDirectory, indexFunction)
}

func SnapshotTestDirectories(t *testing.T, inputDirectory, outputDirectory string, indexFunction indexFunction) {
	if *updateLsifSnapshots {
		err := os.RemoveAll(outputDirectory)
		require.Nil(t, err)
	}
	testCases, err := os.ReadDir(inputDirectory)
	require.Nil(t, err)
	if len(testCases) == 0 {
		t.Fatalf("no subdirectories: %v", inputDirectory)
	}

	for _, testCase := range testCases {
		if !testCase.IsDir() {
			t.Fatalf("not a directory: %v", testCase.Name())
		}
		t.Run(testCase.Name(), func(t *testing.T) {
			baseInputDirectory := filepath.Join(inputDirectory, testCase.Name())
			baseOutputDirectory := filepath.Join(outputDirectory, testCase.Name())
			sources, err := lsiftyped.NewSourcesFromDirectory(baseInputDirectory)
			require.Nil(t, err)
			obtainedSnapshots := indexFunction(baseInputDirectory, baseOutputDirectory, sources)
			snapshotTestSources(t, baseOutputDirectory, obtainedSnapshots)
		})
	}
}

func snapshotTestSources(t *testing.T, outputDirectory string, obtainedSnapshots []*lsiftyped.SourceFile) {
	for _, document := range obtainedSnapshots {
		t.Run(document.RelativePath, func(t *testing.T) {
			obtained := document.Text
			outputFile := filepath.Join(outputDirectory, document.RelativePath)
			expected, err := os.ReadFile(outputFile)
			if err != nil {
				expected = []byte{}
			}
			if *updateLsifSnapshots {
				err = os.MkdirAll(filepath.Dir(outputFile), 0755)
				require.Nil(t, err)
				err = os.WriteFile(outputFile, []byte(obtained), 0755)
				require.Nil(t, err)
			} else {
				edits := myers.ComputeEdits(span.URIFromPath(outputFile), string(expected), obtained)
				if len(edits) > 0 {
					diff := fmt.Sprint(gotextdiff.ToUnified(
						outputFile+" (obtained)",
						outputFile+" (expected)",
						string(expected),
						edits,
					))
					t.Fatalf("\n" + diff)
				}
			}
		})
	}
}
