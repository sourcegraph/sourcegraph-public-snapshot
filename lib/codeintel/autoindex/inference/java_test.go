package inference

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
)

func TestJavaPatterns(t *testing.T) {
	testLangPatterns(t, JavaPatterns(), []PathTestCase{
		{"lsif-java.json", true},
		{"A.java", true},
		{"A.scala", true},
		{"A.kt", true},
		// TODO: Turn these on after adding support for more complex projects.
		// {"settings.gradle", true}
		// {"build.gradle", true}
		// {"pom.xml", true}
	})
}

func TestInferJavaIndexJobs(t *testing.T) {
	paths := []string{
		"lsif-java.json",
		"A.java",
	}

	expectedIndexJobs := []config.IndexJob{
		{
			Indexer: "sourcegraph/lsif-java",
			IndexerArgs: []string{
				"lsif-java index --build-tool=lsif",
			},
			Outfile: "dump.lsif",
			Root:    "",
			Steps:   []config.DockerStep{},
		},
	}
	if diff := cmp.Diff(expectedIndexJobs, InferJavaIndexJobs(NewMockGitClient(), paths)); diff != "" {
		t.Errorf("unexpected index jobs (-want +got):\n%s", diff)
	}
}

func TestInferJavaIndexJobHints(t *testing.T) {
	paths := []string{
		"build.gradle",
		"kt/build.gradle.kts",
		"maven/pom.xml",
		"subdir/src/java/App.java",
		"subdir/src/kotlin/App.kt",
		"subdir/src/scala/App.scala",
	}

	expectedHints := []config.IndexJobHint{
		{
			Root:           ".",
			Indexer:        "sourcegraph/lsif-java",
			HintConfidence: config.HintConfidenceProjectStructureSupported,
		},
		{
			Root:           "kt",
			Indexer:        "sourcegraph/lsif-java",
			HintConfidence: config.HintConfidenceProjectStructureSupported,
		},
		{
			Root:           "maven",
			Indexer:        "sourcegraph/lsif-java",
			HintConfidence: config.HintConfidenceProjectStructureSupported,
		},
		{
			Root:           "subdir/src/java",
			Indexer:        "sourcegraph/lsif-java",
			HintConfidence: config.HintConfidenceLanguageSupport,
		},
		{
			Root:           "subdir/src/kotlin",
			Indexer:        "sourcegraph/lsif-java",
			HintConfidence: config.HintConfidenceLanguageSupport,
		},
		{
			Root:           "subdir/src/scala",
			Indexer:        "sourcegraph/lsif-java",
			HintConfidence: config.HintConfidenceLanguageSupport,
		},
	}

	if diff := cmp.Diff(expectedHints, InferJavaIndexJobHints(NewMockGitClient(), paths)); diff != "" {
		t.Errorf("unexpected index job hints (-want +got)\n%s", diff)
	}
}
