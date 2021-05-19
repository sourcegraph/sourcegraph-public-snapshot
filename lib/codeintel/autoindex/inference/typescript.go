package inference

import (
	"context"
	"encoding/json"
	"path/filepath"
	"regexp"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
)

const (
	lsifTscImage = "sourcegraph/lsif-node:autoindex"
	nMuslCommand = "N_NODE_MIRROR=https://unofficial-builds.nodejs.org/download/release n --arch x64-musl auto"
)

func CanIndexTypeScriptRepo(gitclient GitClient, paths []string) bool {
	for _, path := range paths {
		if canIndexTypeScriptPath(path) {
			return true
		}
	}

	return false
}

func InferTypeScriptIndexJobs(gitclient GitClient, paths []string) (indexes []config.IndexJob) {
	for _, path := range paths {
		if !canIndexTypeScriptPath(path) {
			continue
		}

		// check first if anywhere along the ancestor path there is a lerna.json
		isYarn := checkLernaFile(gitclient, path, paths)

		var dockerSteps []config.DockerStep

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

			dockerSteps = append(dockerSteps, config.DockerStep{
				Root:     dir,
				Image:    lsifTscImage,
				Commands: commands,
			})
		}

		var localSteps []string

		if checkCanDeriveNodeVersion(gitclient, path, paths) {
			for i, step := range dockerSteps {
				step.Commands = append([]string{nMuslCommand}, step.Commands...)
				dockerSteps[i] = step
			}
			localSteps = append(localSteps, nMuslCommand)
		}

		n := len(dockerSteps)
		for i := 0; i < n/2; i++ {
			dockerSteps[i], dockerSteps[n-i-1] = dockerSteps[n-i-1], dockerSteps[i]
		}

		indexes = append(indexes, config.IndexJob{
			Steps:       dockerSteps,
			LocalSteps:  localSteps,
			Root:        dirWithoutDot(path),
			Indexer:     lsifTscImage,
			IndexerArgs: []string{"lsif-tsc", "-p", "."},
			Outfile:     "",
		})
	}

	return indexes
}

func checkCanDeriveNodeVersion(gitclient GitClient, path string, paths []string) bool {
	for _, dir := range ancestorDirs(path) {
		packageJSONPath := filepath.Join(dir, "package.json")
		nvmrcPath := filepath.Join(dir, ".nvmrc")
		nodeVersionPath := filepath.Join(dir, ".node-version")
		nnodeVersionPath := filepath.Join(dir, ".n-node-version")
		if (contains(paths, packageJSONPath) && hasEnginesField(gitclient, packageJSONPath)) ||
			contains(paths, nvmrcPath) ||
			contains(paths, nodeVersionPath) ||
			contains(paths, nnodeVersionPath) {
			return true
		}
	}
	return false
}

func hasEnginesField(gitclient GitClient, packageJSONPath string) (hasField bool) {
	type packageJSONEngine struct {
		Engines *struct {
			Node *string `json:"node"`
		} `json:"engines"`
	}

	packageJSON := &packageJSONEngine{}
	if b, err := gitclient.RawContents(context.TODO(), packageJSONPath); err == nil {
		if err := json.Unmarshal(b, packageJSON); err == nil {
			if packageJSON.Engines != nil && packageJSON.Engines.Node != nil {
				return true
			}
		}
	}
	return
}

func checkLernaFile(gitclient GitClient, path string, paths []string) (isYarn bool) {
	type lernaConfig struct {
		NPMClient string `json:"npmClient"`
	}

	for _, dir := range ancestorDirs(path) {
		lernaPath := filepath.Join(dir, "lerna.json")
		if contains(paths, lernaPath) && !isYarn {
			if b, err := gitclient.RawContents(context.TODO(), lernaPath); err == nil {
				var c lernaConfig
				if err := json.Unmarshal(b, &c); err == nil {
					isYarn = c.NPMClient == "yarn"
				}
			}
		}
	}
	return
}

func TypeScriptPatterns() []*regexp.Regexp {
	return []*regexp.Regexp{
		suffixPattern("tsconfig.json"),
		suffixPattern("package.json"),
		suffixPattern("lerna.json"),
		suffixPattern("yarn.lock"),
		suffixPattern(".nvmrc"),
		suffixPattern(".node-version"),
		suffixPattern(".n-node-version"),
	}
}

func canIndexTypeScriptPath(path string) bool {
	// TODO(efritz) - check for javascript files
	return filepath.Base(path) == "tsconfig.json" && containsNoSegments(path, tscSegmentBlockList...)
}

var tscSegmentBlockList = append([]string{
	"node_modules",
}, segmentBlockList...)
