package inference

import (
	"context"
	"encoding/json"
	"path/filepath"
	"regexp"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
)

func TypeScriptPatterns() []*regexp.Regexp {
	return []*regexp.Regexp{
		pathPattern(rawPattern("tsconfig.json")),
		pathPattern(rawPattern("package.json")),
		pathPattern(rawPattern("lerna.json")),
		pathPattern(rawPattern("yarn.lock")),
		pathPattern(rawPattern(".nvmrc")),
		pathPattern(rawPattern(".node-version")),
		pathPattern(rawPattern(".n-node-version")),
	}
}

func CanIndexTypeScriptRepo(gitclient GitClient, paths []string) bool {
	for _, path := range paths {
		if canIndexTypeScriptPath(path) {
			return true
		}
	}

	return false
}

const lsifTscImage = "sourcegraph/lsif-node:autoindex"
const nMuslCommand = "N_NODE_MIRROR=https://unofficial-builds.nodejs.org/download/release n --arch x64-musl auto"

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

func checkLernaFile(gitclient GitClient, path string, paths []string) (isYarn bool) {
	lernaConfig := struct {
		NPMClient string `json:"npmClient"`
	}{}

	for _, dir := range ancestorDirs(path) {
		lernaPath := filepath.Join(dir, "lerna.json")

		if contains(paths, lernaPath) && !isYarn {
			if b, err := gitclient.RawContents(context.TODO(), lernaPath); err == nil {
				if err := json.Unmarshal(b, &lernaConfig); err == nil {
					isYarn = lernaConfig.NPMClient == "yarn"
				}
			}
		}
	}
	return
}

func checkCanDeriveNodeVersion(gitclient GitClient, path string, paths []string) bool {
	for _, dir := range ancestorDirs(path) {
		packageJSONPath := filepath.Join(dir, "package.json")
		nvmrcPath := filepath.Join(dir, ".nvmrc")
		nodeVersionPath := filepath.Join(dir, ".node-version")
		nnodeVersionPath := filepath.Join(dir, ".n-node-version")

		// TODO - refactor this
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
	packageJSON := struct {
		Engines *struct {
			Node *string `json:"node"`
		} `json:"engines"`
	}{}

	if b, err := gitclient.RawContents(context.TODO(), packageJSONPath); err == nil {
		if err := json.Unmarshal(b, &packageJSON); err == nil {
			if packageJSON.Engines != nil && packageJSON.Engines.Node != nil {
				return true
			}
		}
	}
	return
}

var tscSegmentBlockList = append([]string{"node_modules"}, segmentBlockList...)

func canIndexTypeScriptPath(path string) bool {
	return filepath.Base(path) == "tsconfig.json" && containsNoSegments(path, tscSegmentBlockList...)
}
