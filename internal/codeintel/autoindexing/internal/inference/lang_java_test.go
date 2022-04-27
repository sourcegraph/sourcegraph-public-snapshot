package inference

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
)

func TestRecognizersJava(t *testing.T) {
	testRecognizers(t,
		recognizerTestCase{
			description: "java project with lsif-java.json",
			repositoryContents: map[string]string{
				"lsif-java.json": "",
				"src/java/com/sourcegraph/codeintel/dumb.java": "",
				"src/java/com/sourcegraph/codeintel/fun.scala": "",
			},
			expected: []config.IndexJob{
				{
					Steps:       nil,
					LocalSteps:  nil,
					Root:        "",
					Indexer:     "sourcegraph/lsif-java",
					IndexerArgs: []string{"lsif-java", "index", "--build-tool=lsif"},
					Outfile:     "dump.lsif",
				},
			},
		},
		recognizerTestCase{
			description: "java project without lsif-java.json (no match)",
			repositoryContents: map[string]string{
				"src/java/com/sourcegraph/codeintel/dumb.java": "",
				"src/java/com/sourcegraph/codeintel/fun.scala": "",
			},
			expected: nil,
		},
	)
}
