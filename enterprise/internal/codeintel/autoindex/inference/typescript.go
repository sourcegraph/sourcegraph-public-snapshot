package inference

import (
	"context"
	"encoding/json"
	"path/filepath"
	"regexp"
)

const (
	lsifTscImage     = "sourcegraph/lsif-node:latest"
	nodeInstallImage = "node:alpine3.12"
)

type lsifTscJobRecognizer struct{}

var _ IndexJobRecognizer = lsifTscJobRecognizer{}

type lernaConfig struct {
	NPMClient string `json:"npmClient"`
}

func (r lsifTscJobRecognizer) CanIndex(paths []string, gitserver GitserverClientWrapper) bool {
	for _, path := range paths {
		if r.canIndexPath(path) {
			return true
		}
	}

	return false
}

func (r lsifTscJobRecognizer) InferIndexJobs(paths []string, gitserver GitserverClientWrapper) (indexes []IndexJob) {
	for _, path := range paths {
		if !r.canIndexPath(path) {
			continue
		}

		// check first if anywhere along the ancestor path there is a lerna.json
		isYarn := checkLernaFile(path, paths, gitserver)

		var dockerSteps []DockerStep

		for _, dir := range ancestorDirs(path) {
			if !contains(paths, filepath.Join(dir, "package.json")) {
				continue
			}

			var commands []string
			if isYarn || contains(paths, filepath.Join(dir, "yarn.lock")) {
				commands = append(commands, "yarn --ignore-engines")
			} else {
				commands = append(commands, "npm install")
			}

			dockerSteps = append(dockerSteps, DockerStep{
				Root:     dir,
				Image:    nodeInstallImage,
				Commands: commands,
			})
		}

		n := len(dockerSteps)
		for i := 0; i < n/2; i++ {
			dockerSteps[i], dockerSteps[n-i-1] = dockerSteps[n-i-1], dockerSteps[i]
		}

		indexes = append(indexes, IndexJob{
			DockerSteps: dockerSteps,
			Root:        dirWithoutDot(path),
			Indexer:     lsifTscImage,
			IndexerArgs: []string{"lsif-tsc", "-p", "."},
			Outfile:     "",
		})
	}

	return indexes
}

func checkLernaFile(path string, paths []string, gitserver GitserverClientWrapper) (isYarn bool) {
	for _, dir := range ancestorDirs(path) {
		lernaPath := filepath.Join(dir, "lerna.json")
		if exists := contains(paths, lernaPath); exists && !isYarn {
			if b, err := gitserver.RawContents(context.TODO(), lernaPath); err == nil {
				var c lernaConfig
				if err := json.Unmarshal(b, &c); err == nil {
					isYarn = c.NPMClient == "yarn"
				}
			}
		}
	}
	return
}

func (lsifTscJobRecognizer) Patterns() []*regexp.Regexp {
	return []*regexp.Regexp{
		suffixPattern("tsconfig.json"),
		suffixPattern("package.json"),
		suffixPattern("lerna.json"),
	}
}

func (r lsifTscJobRecognizer) canIndexPath(path string) bool {
	// TODO(efritz) - check for javascript files
	return filepath.Base(path) == "tsconfig.json" && containsNoSegments(path, tscSegmentBlockList...)
}

var tscSegmentBlockList = append([]string{
	"node_modules",
}, segmentBlockList...)
