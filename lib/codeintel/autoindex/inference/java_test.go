package inference

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
)

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

func TestJavaPatterns(t *testing.T) {
	paths := []string{
		"lsif-java.json",
		"A.java",
		"A.scala",
		"A.kt",
		// "settings.gradle",
		// "build.gradle",
		// "pom.xml",
	}

	for _, path := range paths {
		match := false
		for _, pattern := range JavaPatterns() {
			if pattern.MatchString(path) {
				match = true
				break
			}
		}

		if !match {
			t.Error(fmt.Sprintf("failed to match %s", path))
		}
	}
}
