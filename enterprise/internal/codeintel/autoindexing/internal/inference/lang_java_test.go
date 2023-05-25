package inference

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing/internal/inference/libs"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
)

func TestJavaGenerator(t *testing.T) {
	expectedIndexerImage, _ := libs.DefaultIndexerForLang("java")

	autoJob := config.IndexJob{
		Steps:       nil,
		LocalSteps:  nil,
		Root:        "",
		Indexer:     expectedIndexerImage,
		IndexerArgs: []string{"scip-java", "index", "--build-tool=auto"},
		Outfile:     "index.scip",
	}

	singleAutoJob := []config.IndexJob{autoJob}

	testGenerators(t,
		generatorTestCase{
			description: "JVM project with lsif-java.json",
			repositoryContents: map[string]string{
				"lsif-java.json": "",
				"src/java/com/sourcegraph/codeintel/dumb.java": "",
				"src/java/com/sourcegraph/codeintel/fun.scala": "",
			},
			expected: singleAutoJob,
		},
		generatorTestCase{
			description: "JVM project with Gradle",
			repositoryContents: map[string]string{
				"build.gradle": "",
				"src/java/com/sourcegraph/codeintel/dumb.java": "",
				"src/java/com/sourcegraph/codeintel/fun.scala": "",
			},
			expected: singleAutoJob,
		},
		generatorTestCase{
			description: "JVM project with SBT",
			repositoryContents: map[string]string{
				"build.sbt": "",
				"src/java/com/sourcegraph/codeintel/dumb.java": "",
				"src/java/com/sourcegraph/codeintel/fun.scala": "",
			},
			expected: singleAutoJob,
		},
		generatorTestCase{
			description: "JVM project with Maven",
			repositoryContents: map[string]string{
				"pom.xml": "",
				"src/java/com/sourcegraph/codeintel/dumb.java": "",
				"src/java/com/sourcegraph/codeintel/fun.scala": "",
			},
			expected: singleAutoJob,
		},
		generatorTestCase{
			description: "JVM project without build file",
			repositoryContents: map[string]string{
				"src/java/com/sourcegraph/codeintel/dumb.java": "",
				"src/java/com/sourcegraph/codeintel/fun.scala": "",
			},
			expected: []config.IndexJob{},
		},
		generatorTestCase{
			description: "JVM project with Maven build file but no sources",
			repositoryContents: map[string]string{
				"pom.xml": "",
			},
			expected: []config.IndexJob{},
		},
		generatorTestCase{
			description: "JVM project with Gradle build file but no sources",
			repositoryContents: map[string]string{
				"build.gradle": "",
			},
			expected: []config.IndexJob{},
		},
		generatorTestCase{
			description: "JVM project with SBT build file but no sources",
			repositoryContents: map[string]string{
				"build.sbt": "",
			},
			expected: []config.IndexJob{},
		},
		generatorTestCase{
			description: "JVM project with Mill build file but no sources",
			repositoryContents: map[string]string{
				"build.sc": "",
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
				"build.gradle":        "",
				"kt/build.gradle.kts": "",
				"maven/pom.xml":       "",
				"scala/build.sbt":     "",
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
					Root:           "scala",
					Indexer:        expectedIndexerImage,
					HintConfidence: config.HintConfidenceProjectStructureSupported,
				},
			},
		},
	)
}
