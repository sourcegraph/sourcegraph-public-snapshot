package inference

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing/internal/inference/libs"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
)

func TestJavaGenerator(t *testing.T) {
	expectedIndexerImage, _ := libs.DefaultIndexerForLang("java")

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
					Indexer:     expectedIndexerImage,
					IndexerArgs: []string{"scip-java", "index", "--build-tool=scip"},
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
	expectedIndexerImage := "sourcegraph/scip-java@sha256:4fee3d21692df0e3cd59245ee2adcf5a522d46e450fb9bc0561dc0907f50844b"

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
					Indexer:        expectedIndexerImage,
					HintConfidence: config.HintConfidenceProjectStructureSupported,
				},
				{
					Root:           "kt",
					Indexer:        expectedIndexerImage,
					HintConfidence: config.HintConfidenceProjectStructureSupported,
				},
				{
					Root:           "maven",
					Indexer:        expectedIndexerImage,
					HintConfidence: config.HintConfidenceProjectStructureSupported,
				},
				{
					Root:           "subdir/src/java",
					Indexer:        expectedIndexerImage,
					HintConfidence: config.HintConfidenceLanguageSupport,
				},
				{
					Root:           "subdir/src/kotlin",
					Indexer:        expectedIndexerImage,
					HintConfidence: config.HintConfidenceLanguageSupport,
				},
				{
					Root:           "subdir/src/scala",
					Indexer:        expectedIndexerImage,
					HintConfidence: config.HintConfidenceLanguageSupport,
				},
			},
		},
	)
}
