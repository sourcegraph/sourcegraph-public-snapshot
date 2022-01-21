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

const lsifTscImage = "sourcegraph/lsif-node:autoindex"
const nMuslCommand = "N_NODE_MIRROR=https://unofficial-builds.nodejs.org/download/release n --arch x64-musl auto"

type pathMapValue struct {
	// indexes stored in order for deterministic iteration
	indexes  []int
	baseDirs map[string]struct{}
}

type pathMap map[string]*pathMapValue

func newPathMap(paths []string) (p pathMap) {
	p = map[string]*pathMapValue{}
	for i, path := range paths {
		dir, baseName := filepath.Split(path)
		dir = filepath.Clean(dir)
		if entry, ok := p[baseName]; ok {
			if _, dirPresent := entry.baseDirs[dir]; !dirPresent {
				entry.indexes = append(entry.indexes, i)
				entry.baseDirs[dir] = struct{}{}
			} // dirPresent ==> paths must have duplicates; ignore them
		} else {
			p[baseName] = &pathMapValue{[]int{i}, map[string]struct{}{dir: {}}}
		}
	}
	return p
}

func (p pathMap) contains(dir string, baseName string) bool {
	dir = filepath.Clean(dir)
	if entry, ok := p[baseName]; ok {
		if _, ok := entry.baseDirs[dir]; ok {
			return true
		}
	}
	return false
}

var tscSegmentBlockList = append([]string{"node_modules"}, segmentBlockList...)

func InferTypeScriptIndexJobs(gitclient GitClient, paths []string) (indexes []config.IndexJob) {
	pathMap := newPathMap(paths)

	tsConfigEntry, tsConfigPresent := pathMap["tsconfig.json"]
	if !tsConfigPresent {
		return indexes
	}

	for _, tsConfigIndex := range tsConfigEntry.indexes {
		path := paths[tsConfigIndex]
		if !containsNoSegments(path, tscSegmentBlockList...) {
			continue
		}
		isYarn := checkLernaFile(gitclient, path, pathMap)
		var dockerSteps []config.DockerStep
		for _, dir := range ancestorDirs(path) {
			if !pathMap.contains(dir, "package.json") {
				continue
			}

			var commands []string
			if isYarn || pathMap.contains(dir, "yarn.lock") {
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
		if checkCanDeriveNodeVersion(gitclient, path, pathMap) {
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

func checkLernaFile(gitclient GitClient, path string, pathMap pathMap) (isYarn bool) {
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
