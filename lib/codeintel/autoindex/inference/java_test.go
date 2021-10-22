package inference

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
)

func TestCanIndexJavaRepo(t *testing.T) {
	testCases := []struct {
		paths    []string
		expected bool
	}{
		{paths: []string{"pom.xml"}, expected: false},
		{paths: []string{"nested/pom.xml"}, expected: false},
		{paths: []string{"build.gradle"}, expected: false},
		{paths: []string{"nested/build.gradle"}, expected: false},
		{paths: []string{"settings.gradle"}, expected: false},
		{paths: []string{"nested/settings.gradle"}, expected: false},
		{paths: []string{"lsif-java.json"}, expected: false},
		{paths: []string{"lsif-java.json", "A.kt"}, expected: true},
		{paths: []string{"lsif-java.json", "A.java"}, expected: true},
		{paths: []string{"lsif-java.json", "A.scala"}, expected: true},
		{paths: []string{"lsif-java.json", "A.java", "A.scala"}, expected: true},
		{paths: []string{"nested/lsif-java.json"}, expected: false},
		{paths: []string{"build.sbt"}, expected: false},
		{paths: []string{"package.json"}, expected: false},
		{paths: []string{"MyApp.java"}, expected: false},
		{paths: []string{"MyApp.groovy"}, expected: false},
		{paths: []string{"gradle.properties"}, expected: false},
	}

	for _, testCase := range testCases {
		name := strings.Join(testCase.paths, ", ")

		t.Run(name, func(t *testing.T) {
			if value := CanIndexJavaRepo(NewMockGitClient(), testCase.paths); value != testCase.expected {
				t.Errorf("unexpected result from CanIndex. want=%v have=%v", testCase.expected, value)
			}
		})
	}
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
