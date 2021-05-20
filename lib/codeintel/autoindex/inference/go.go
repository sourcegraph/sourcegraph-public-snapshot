package inference

import (
	"path/filepath"
	"regexp"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
)

func GoPatterns() []*regexp.Regexp {
	return []*regexp.Regexp{
		// go.mod in any directory
		pathPattern(rawPattern("go.mod")),
		// *.go file in root directory
		prefixPattern(suffixPattern(extensionPattern(rawPattern("go")))),
	}
}

func CanIndexGoRepo(gitclient GitClient, paths []string) bool {
	for _, path := range paths {
		if isGoModulePath(path) || isPreModuleGoProjectPath(path) {
			return true
		}
	}

	return false
}

const lsifGoImage = "sourcegraph/lsif-go:latest"

func InferGoIndexJobs(gitclient GitClient, paths []string) (indexes []config.IndexJob) {
	for _, path := range paths {
		if !isGoModulePath(path) {
			continue
		}

		root := dirWithoutDot(path)

		dockerSteps := []config.DockerStep{
			{
				Root:     root,
				Image:    lsifGoImage,
				Commands: []string{"go mod download"},
			},
		}

		indexes = append(indexes, config.IndexJob{
			Steps:       dockerSteps,
			Root:        root,
			Indexer:     lsifGoImage,
			IndexerArgs: []string{"lsif-go", "--no-animation"},
			Outfile:     "",
		})
	}
	if len(indexes) > 0 {
		return indexes
	}

	for _, path := range paths {
		if !isPreModuleGoProjectPath(path) {
			continue
		}

		return []config.IndexJob{
			{
				Steps:       nil,
				Root:        "",
				Indexer:     lsifGoImage,
				IndexerArgs: []string{"GO111MODULE=off", "lsif-go", "--no-animation"},
				Outfile:     "",
			},
		}
	}

	return nil
}

var goSegmentBlockList = append([]string{"vendor"}, segmentBlockList...)

func isGoModulePath(path string) bool {
	return filepath.Base(path) == "go.mod" && containsNoSegments(path, goSegmentBlockList...)
}

func isPreModuleGoProjectPath(path string) bool {
	return filepath.Ext(path) == ".go" && containsNoSegments(path, goSegmentBlockList...)
}
