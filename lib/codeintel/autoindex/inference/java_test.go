package inference

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindex/config"
)

func TestLSIFJavaJobRecognizerCanIndex(t *testing.T) {
	recognizer := lsifJavaJobRecognizer{}
	testCases := []struct {
		paths    []string
		expected bool
	}{
		{paths: []string{"pom.xml"}, expected: true},
		{paths: []string{"nested/pom.xml"}, expected: false},
		{paths: []string{"build.gradle"}, expected: true},
		{paths: []string{"nested/build.gradle"}, expected: false},
		{paths: []string{"settings.gradle"}, expected: true},
		{paths: []string{"nested/settings.gradle"}, expected: false},
		{paths: []string{"lsif-java.json"}, expected: true},
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
			if value := recognizer.CanIndex(testCase.paths, NewMockGitserverClientWrapper()); value != testCase.expected {
				t.Errorf("unexpected result from CanIndex. want=%v have=%v", testCase.expected, value)
			}
		})
	}
}

func TestLsifJavaJobRecognizerInferIndexJobsTsConfigRoot(t *testing.T) {
	recognizer := lsifJavaJobRecognizer{}
	paths := []string{
		"pom.xml",
	}

	expectedIndexJobs := []config.IndexJob{
		{
			Indexer: "gradle:7.0.0-jdk8",
			LocalSteps: []string{
				"apt-get update",
				"apt-get install --yes maven",
				"curl -fLo coursier https://git.io/coursier-cli",
				"chmod +x coursier",
			},
			IndexerArgs: []string{
				"./coursier launch --contrib lsif-java -- --packagehub=https://packagehub-ohcltxh6aq-uc.a.run.app index",
			},
			Outfile: "dump.lsif",
			Root:    "",
			Steps:   []config.DockerStep{},
		},
	}
	if diff := cmp.Diff(expectedIndexJobs, recognizer.InferIndexJobs(paths, NewMockGitserverClientWrapper())); diff != "" {
		t.Errorf("unexpected index jobs (-want +got):\n%s", diff)
	}
}

func TestLSIFJavaJobRecognizerPatterns(t *testing.T) {
	recognizer := lsifJavaJobRecognizer{}
	paths := []string{
		"lsif-java.json",
		"settings.gradle",
		"build.gradle",
		"pom.xml",
	}

	for _, path := range paths {
		match := false
		for _, pattern := range recognizer.Patterns() {
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
