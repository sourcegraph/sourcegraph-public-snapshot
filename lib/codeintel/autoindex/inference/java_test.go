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
