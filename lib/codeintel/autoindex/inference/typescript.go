package inference

import (
	"context"
	"encoding/json"
	"fmt"
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

const lsifTscImage = "sourcegraph/lsif-node:autoindex"
const nMuslCommand = "N_NODE_MIRROR=https://unofficial-builds.nodejs.org/download/release n --arch x64-musl auto"

var tscSegmentBlockList = append([]string{"node_modules"}, segmentBlockList...)

func InferTypeScriptIndexJobs(gitclient GitClient, paths []string) (indexes []config.IndexJob) {
	pathMap := newPathMap(paths)
	createTSConfigCommand := map[string][]string{}

	tsConfigPaths := pathMap.pathsFor("tsconfig.json")

	// TypeScript packages on NPM are commonly distributed without a
	// tsconfig.json file. Instead of skipping indexing altogether,
	// synthesize some permissive tsconfig.json files, as a treat.
	if len(tsConfigPaths) == 0 {
		packageJSONPaths := pathMap.pathsFor("package.json")
		if len(packageJSONPaths) == 0 {
			return indexes
		}
		for _, packageJSONPath := range packageJSONPaths {
			if !containsNoSegments(packageJSONPath, tscSegmentBlockList...) {
				continue
			}
			dir := filepath.Dir(packageJSONPath)
			tsConfigPath := filepath.Join(dir, "tsconfig.json")
			createTSConfigCommand[tsConfigPath] = []string{
				fmt.Sprintf("echo '{\"compilerOptions\":{\"allowJs\":true}}' > %s", tsConfigPath),
			}
			pathMap.insert(dir, "tsconfig.json")
		}
		if len(createTSConfigCommand) == 0 {
			return indexes
		}
		tsConfigPaths = pathMap.pathsFor("tsconfig.json")
	}

	for _, tsConfigPath := range tsConfigPaths {
		if !containsNoSegments(tsConfigPath, tscSegmentBlockList...) {
			continue
		}
		isYarn := checkLernaFile(gitclient, tsConfigPath, pathMap)
		var dockerSteps []config.DockerStep
		for _, dir := range ancestorDirs(tsConfigPath) {
			if !pathMap.contains(dir, "package.json") {
				continue
			}

			var commands []string
			if cmds, ok := createTSConfigCommand[tsConfigPath]; ok {
				commands = make([]string, len(cmds))
				copy(commands, cmds)
			}

			if isYarn || pathMap.contains(dir, "yarn.lock") {
				commands = append(commands, "yarn --ignore-engines")
			} else {
				// [QUESTION] Should we make --ignore-scripts conditional for
				// packages? For source repos, the TypeScript source should be
				// present elsewhere...
				//
				// TypeScript code sometimes deletes the 'dist' directory as
				// part of a pre-build step, as it is used for the generated
				// JavaScript code and type definitions.
				// https://sourcegraph.com/search?q=context:global+%22build%22:+%22%28rm+-rf%7Cdel-cli%29+dist+file:package.json+count:all&patternType=regexp
				//
				// This delete step is a problem when dealing with the
				// corresponding NPM package, since the original TypeScript
				// code may not be part of the package tarball. If we just
				// run `npm install` after extracting the tarball, we may get
				// rid of the available source code, and emit an empty index.
				// Avoid that with `--ignore-scripts`.
				commands = append(commands, "npm install --ignore-scripts")
			}

			dockerSteps = append(dockerSteps, config.DockerStep{
				Root:     dir,
				Image:    lsifTscImage,
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

		indexes = append(indexes, config.IndexJob{
			Steps:       dockerSteps,
			LocalSteps:  localSteps,
			Root:        dirWithoutDot(tsConfigPath),
			Indexer:     lsifTscImage,
			IndexerArgs: []string{"lsif-tsc", "-p", "."},
			Outfile:     "",
		})
	}

	return indexes
}

func checkLernaFile(gitclient GitClient, path string, pathMap *pathMap) (isYarn bool) {
	lernaConfig := struct {
		NPMClient string `json:"npmClient"`
	}{}

	for _, dir := range ancestorDirs(path) {
		if pathMap.contains(dir, "lerna.json") {
			lernaPath := filepath.Join(dir, "lerna.json")
			if b, err := gitclient.RawContents(context.TODO(), lernaPath); err == nil {
				if err := json.Unmarshal(b, &lernaConfig); err == nil && lernaConfig.NPMClient == "yarn" {
					return true
				}
			}
		}
	}
	return false
}

func checkCanDeriveNodeVersion(gitclient GitClient, path string, pathMap *pathMap) bool {
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
