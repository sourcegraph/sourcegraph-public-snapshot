package inference

import (
	"regexp"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
)

func CanIndexJavaRepo(gitserver GitClient, paths []string) bool {
	for _, path := range paths {
		if isJavaPath(path) {
			return true
		}
	}
	return false
}

func InferJavaIndexJobs(gitserver GitClient, paths []string) (indexes []config.IndexJob) {
	for _, path := range paths {
		if !isJavaPath(path) {
			continue
		}
		indexes = append(indexes, config.IndexJob{
			Indexer: "sourcegraph/lsif-java",
			IndexerArgs: []string{
				"/coursier launch --contrib --ttl 0 lsif-java -- index",
			},
			Outfile: "dump.lsif",
			Root:    "",
			Steps:   []config.DockerStep{},
		})
	}
	return indexes
}

func JavaPatterns() []*regexp.Regexp {
	var patterns []*regexp.Regexp
	for _, filename := range supportedFilenames {
		patterns = append(patterns, suffixPattern(rawPattern(filename)))
	}
	return patterns
}

func isJavaPath(path string) bool {
	for _, filename := range supportedFilenames {
		if filename == path {
			return true
		}
	}
	return false
}

var supportedFilenames = []string{
	// The "lsif-java.json" file is used to index package repositories such as
	// the JDK sources and published Java libraries.
	"lsif-java.json",
	// "build.sbt",
	// Maven and Gradle are intentionally excluded from these patterns to
	// begin with. We want to gain experience with auto-indexing only
	// package repos, which have a higher likelyhood of indexing
	// successfully because they have a simpler build structure compared to
	// Gradle/Maven repos.
	// "pom.xml",
	// "build.gradle",
	// "settings.gradle",
}
