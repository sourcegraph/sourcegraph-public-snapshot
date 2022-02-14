package inference

import (
	"path/filepath"

	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
)

func InferRustIndexJobs(gitserver GitClient, paths []string) (indexes []config.IndexJob) {
	for _, path := range paths {
		if canIndexRustPath(path) {
			// We only create one IndexJob, as the complexity of indexing
			// large projects, such as those with workspaces, is handled
			// in lsif-rust.
			indexes = append(indexes, config.IndexJob{
				Indexer:     "sourcegraph/lsif-rust",
				IndexerArgs: []string{"lsif-rust", "index"},
				Outfile:     "dump.lsif",
				Root:        "",
				Steps:       []config.DockerStep{},
			})
			break
		}
	}
	return indexes
}

func RustPatterns() []*regexp.Regexp {
	return []*regexp.Regexp{
		suffixPattern(rawPattern("Cargo.toml")),
	}
}

func canIndexRustPath(path string) bool {
	return filepath.Base(path) == "Cargo.toml"
}
