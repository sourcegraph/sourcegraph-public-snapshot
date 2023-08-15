package inference

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/internal/inference/libs"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
)

func autoJob(root string) config.IndexJob {
	expectedIndexerImage, _ := libs.DefaultIndexerForLang("java")
	return config.IndexJob{
		Steps:       nil,
		LocalSteps:  nil,
		Root:        root,
		Indexer:     expectedIndexerImage,
		IndexerArgs: []string{"scip-java", "index", "--build-tool=auto"},
		Outfile:     "index.scip",
	}
}

func TestJavaGenerator(t *testing.T) {
	singleTopLevelJob := []config.IndexJob{autoJob("")}

	testGenerators(t,
		generatorTestCase{
			description: "JVM project with lsif-java.json",
			repositoryContents: map[string]string{
				"lsif-java.json": "",
				"src/java/com/sourcegraph/codeintel/dumb.java": "",
				"src/java/com/sourcegraph/codeintel/fun.scala": "",
			},
			expected: singleTopLevelJob,
		},
		generatorTestCase{
			description: "JVM project with Gradle",
			repositoryContents: map[string]string{
				"build.gradle": "",
				"src/java/com/sourcegraph/codeintel/dumb.java": "",
				"src/java/com/sourcegraph/codeintel/fun.scala": "",
			},
			expected: singleTopLevelJob,
		},
		generatorTestCase{
			description: "JVM project with SBT",
			repositoryContents: map[string]string{
				"build.sbt": "",
				"src/java/com/sourcegraph/codeintel/dumb.java": "",
				"src/java/com/sourcegraph/codeintel/fun.scala": "",
			},
			expected: singleTopLevelJob,
		},
		generatorTestCase{
			description: "JVM project with Maven",
			repositoryContents: map[string]string{
				"pom.xml": "",
				"src/java/com/sourcegraph/codeintel/dumb.java": "",
				"src/java/com/sourcegraph/codeintel/fun.scala": "",
			},
			expected: singleTopLevelJob,
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
		generatorTestCase{
			description: "Nested JVM project with top-level build file",
			repositoryContents: map[string]string{
				"build.sbt": "",
				"my-module/src/java/com/sourcegraph/codeintel/dumb.java": "",
				"my-module/pom.xml": "",
			},
			expected: singleTopLevelJob,
		},
		generatorTestCase{
			description: "Nested JVM project WITHOUT top-level build file",
			repositoryContents: map[string]string{
				"my-module/src/java/com/sourcegraph/codeintel/dumb.java": "",
				"my-module/pom.xml": "",
				"our-module/src/java/com/sourcegraph/codeintel/dumb.java": "",
				"our-module/pom.xml": "",
			},
			expected: []config.IndexJob{autoJob("my-module"), autoJob("our-module")},
		},
	)
}

func TestJavaHinter(t *testing.T) {
	expectedIndexerImage, _ := libs.DefaultIndexerForLang("java")

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
