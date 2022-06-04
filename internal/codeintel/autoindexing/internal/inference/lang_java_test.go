package inference

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
)

func TestJavaGenerator(t *testing.T) {
	testGenerators(t,
		generatorTestCase{
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
					Indexer:     "sourcegraph/scip-java",
					IndexerArgs: []string{"scip-java", "index", "--build-tool=lsif"},
					Outfile:     "index.scip",
				},
			},
		},
		generatorTestCase{
			description: "java project without lsif-java.json (no match)",
			repositoryContents: map[string]string{
				"src/java/com/sourcegraph/codeintel/dumb.java": "",
				"src/java/com/sourcegraph/codeintel/fun.scala": "",
			},
			expected: []config.IndexJob{},
		},
	)
}

func TestJavaHinter(t *testing.T) {
	testHinters(t,
		hinterTestCase{
			description: "basic hints",
			repositoryContents: map[string]string{
				"build.gradle":               "",
				"kt/build.gradle.kts":        "",
				"maven/pom.xml":              "",
				"subdir/src/java/App.java":   "",
				"subdir/src/kotlin/App.kt":   "",
				"subdir/src/scala/App.scala": "",
			},
			expected: []config.IndexJobHint{
				{
					Root:           "",
					Indexer:        "sourcegraph/scip-java",
					HintConfidence: config.HintConfidenceProjectStructureSupported,
				},
				{
					Root:           "kt",
					Indexer:        "sourcegraph/scip-java",
					HintConfidence: config.HintConfidenceProjectStructureSupported,
				},
				{
					Root:           "maven",
					Indexer:        "sourcegraph/scip-java",
					HintConfidence: config.HintConfidenceProjectStructureSupported,
				},
				{
					Root:           "subdir/src/java",
					Indexer:        "sourcegraph/scip-java",
					HintConfidence: config.HintConfidenceLanguageSupport,
				},
				{
					Root:           "subdir/src/kotlin",
					Indexer:        "sourcegraph/scip-java",
					HintConfidence: config.HintConfidenceLanguageSupport,
				},
				{
					Root:           "subdir/src/scala",
					Indexer:        "sourcegraph/scip-java",
					HintConfidence: config.HintConfidenceLanguageSupport,
				},
			},
		},
	)
}
