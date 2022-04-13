package inference

import (
	"context"
	"encoding/json"
	"path/filepath"

	"github.com/grafana/regexp"

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

const lsifTypescriptImage = "sourcegraph/lsif-typescript:autoindex"
const nMuslCommand = "N_NODE_MIRROR=https://unofficial-builds.nodejs.org/download/release n --arch x64-musl auto"

var tscSegmentBlockList = append([]string{"node_modules"}, segmentBlockList...)

func inferSingleTypeScriptIndexJob(
	gitclient GitClient, pathMap pathMap, tsConfigPath string, shouldInferConfig bool,
) *config.IndexJob {
	if !containsNoSegments(tsConfigPath, tscSegmentBlockList...) {
		return nil
	}
	isYarn := checkLernaFile(gitclient, tsConfigPath, pathMap)
	var dockerSteps []config.DockerStep
	for _, dir := range ancestorDirs(tsConfigPath) {
		if !pathMap.contains(dir, "package.json") {
			continue
		}

		ignoreScripts := ""
		if shouldInferConfig {
			ignoreScripts = " --ignore-scripts"
		}
		var commands []string
		if isYarn || pathMap.contains(dir, "yarn.lock") {
			commands = append(commands, "yarn --ignore-engines"+ignoreScripts)
		} else {
			commands = append(commands, "npm install"+ignoreScripts)
		}

		dockerSteps = append(dockerSteps, config.DockerStep{
			Root:     dir,
			Image:    lsifTypescriptImage,
			Commands: commands,
		})
	}

	var localSteps []string
	if checkCanDeriveNodeVersion(gitclient, tsConfigPath, pathMap) {
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

	indexerArgs := []string{"lsif-typescript-autoindex", "index"}
	if shouldInferConfig {
		indexerArgs = append(indexerArgs, "--infer-tsconfig")
	}

	return &config.IndexJob{
		Steps:       dockerSteps,
		LocalSteps:  localSteps,
		Root:        dirWithoutDot(tsConfigPath),
		Indexer:     lsifTypescriptImage,
		IndexerArgs: indexerArgs,
		Outfile:     "",
	}
}

func InferTypeScriptIndexJobs(gitclient GitClient, paths []string) (indexes []config.IndexJob) {
	pathMap := newPathMap(paths)

	tsConfigEntry, tsConfigPresent := pathMap["tsconfig.json"]
	if !tsConfigPresent {
		indexJob := inferSingleTypeScriptIndexJob(gitclient, pathMap, "tsconfig.json", true)
		if indexJob != nil {
			return []config.IndexJob{*indexJob}
		}
		return []config.IndexJob{}
	}

	for _, tsConfigIndex := range tsConfigEntry.indexes {
		indexJob := inferSingleTypeScriptIndexJob(gitclient, pathMap, paths[tsConfigIndex], false)
		if indexJob != nil {
			indexes = append(indexes, *indexJob)
		}
	}

	return indexes
}

func checkLernaFile(gitclient GitClient, path string, pathMap pathMap) (isYarn bool) {
	lernaConfig := struct {
		NpmClient string `json:"npmClient"`
	}{}

	for _, dir := range ancestorDirs(path) {
		if pathMap.contains(dir, "lerna.json") {
			lernaPath := filepath.Join(dir, "lerna.json")
			if b, err := gitclient.RawContents(context.TODO(), lernaPath); err == nil {
				if err := json.Unmarshal(b, &lernaConfig); err == nil && lernaConfig.NpmClient == "yarn" {
					return true
				}
			}
		}
	}
	return false
}

func checkCanDeriveNodeVersion(gitclient GitClient, path string, pathMap pathMap) bool {
	for _, dir := range ancestorDirs(path) {
		packageJSONPath := filepath.Join(dir, "package.json")
		if (pathMap.contains(dir, "package.json") && hasEnginesField(gitclient, packageJSONPath)) ||
			pathMap.contains(dir, ".nvmrc") ||
			pathMap.contains(dir, ".node-version") ||
			pathMap.contains(dir, ".n-node-version") {
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
