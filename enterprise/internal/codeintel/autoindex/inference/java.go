package inference

import (
	"regexp"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindex/config"
)

type lsifJavaJobRecognizer struct{}

func (r lsifJavaJobRecognizer) CanIndex(paths []string, gitserver GitserverClientWrapper) bool {
	for _, path := range paths {
		if r.canIndexPath(path) {
			return true
		}
	}
	return false
}

func (r lsifJavaJobRecognizer) InferIndexJobs(paths []string, gitserver GitserverClientWrapper) (indexes []config.IndexJob) {
	for _, path := range paths {
		if !r.canIndexPath(path) {
			continue
		}
		indexes = append(indexes, config.IndexJob{
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
		})
	}
	return indexes
}

func (lsifJavaJobRecognizer) Patterns() []*regexp.Regexp {
	var patterns []*regexp.Regexp
	for _, filename := range supportedFilenames {
		patterns = append(patterns, suffixPattern(filename))
	}
	return patterns
}

func (r lsifJavaJobRecognizer) canIndexPath(path string) bool {
	for _, filename := range supportedFilenames {
		if filename == path {
			return true
		}
	}
	return false
}

var supportedFilenames = []string{
	"pom.xml",
	"build.gradle",
	"settings.gradle",
	// The "lsif-java.json" file is used to index package repositories such as
	// the JDK sources and published Java libraries.
	"lsif-java.json",
}
