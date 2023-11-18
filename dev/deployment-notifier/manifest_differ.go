package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/grafana/regexp"
)

// git diff myapp/myapp.Deployment.yaml ...
// +         image: index.docker.io/sourcegraph/migrator:137540_2022-03-17_d24138504aea@sha256:2b6efe8f447b22f9396544f885f2f326d21325d652f9b36961f3d105723789df
// -         image: index.docker.io/sourcegraph/migrator:137540_2022-03-17_XXXXXXXXXXXX@sha256:2b6efe8f447b22f9396544f885f2f326d21325d652f9b36961f3d105723789df
var imageCommitRegexp = `(?m)^DIFF_OP\s+image:\s[^/]+\/sourcegraph\/[^:]+:\d{6}_\d{4}-\d{2}-\d{2}_([^@]+)@sha256.*$` // (?m) stands for multiline.

type ServiceVersionDiff struct {
	Old string
	New string
}

type DeploymentDiffer interface {
	Services() (map[string]*ServiceVersionDiff, error)
}
type manifestDeploymentDiffer struct {
	changedFiles []string
	diffs        map[string]*ServiceVersionDiff
}

func NewManifestDeploymentDiffer(changedFiles []string) DeploymentDiffer {
	return &manifestDeploymentDiffer{
		changedFiles: changedFiles,
	}
}

func (m *manifestDeploymentDiffer) Services() (map[string]*ServiceVersionDiff, error) {
	err := m.parseManifests()
	if err != nil {
		return nil, err
	}
	return m.diffs, nil
}

func (m *manifestDeploymentDiffer) parseManifests() error {
	services := map[string]*ServiceVersionDiff{}
	for _, path := range m.changedFiles {
		info, err := os.Stat(path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			// If the file is a directory, skip it.
			continue
		}
		ext := filepath.Ext(path)
		if ext != ".yml" && ext != ".yaml" {
			// If the file is not yaml, skip it.
			continue
		}

		elems := strings.Split(path, string(filepath.Separator))
		if len(elems) < 2 {
			// If the file is at the root, skip it. Services are always in subfolders.
			continue
		}
		if elems[0] != "base" {
			// If the file is not in the base folder where services are, skip it.
			continue
		}

		appName := elems[1] // base/elems[1]/...

		filename := filepath.Base(path)
		components := strings.Split(filename, ".")
		if len(components) < 3 {
			// If the file isn't name like appName.Kind.yaml, skip it.
			continue
		}
		kind := components[1]
		if kind == "Deployment" || kind == "StatefulSet" || kind == "DaemonSet" {
			appDiff, err := diffDeploymentManifest(path)
			if err != nil {
				return err
			}
			if appDiff != nil {
				// It's possible that we find changes that are not bumping the image, when
				// updating environment vars for example. In that case, we don't want to
				// include them.
				services[appName] = appDiff
			}
		}
	}
	m.diffs = services
	return nil
}

// imageDiffRegexp returns a regexp that matches an addition or deletion of an
// image tag in the image field in the manifest of an application.
func imageDiffRegexp(addition bool) *regexp.Regexp {
	var escapedOp string
	if addition {
		// If matching an addition, the + needs to be escaped to not be parsed as a
		// count operator.
		escapedOp = "\\+"
	} else {
		escapedOp = "-"
	}

	re := strings.ReplaceAll(imageCommitRegexp, "DIFF_OP", escapedOp)
	return regexp.MustCompile(re)
}

// parseSourcegraphCommitFromDeploymentManifestsDiff parses the diff output, returning
// the new and old commits that were used to build this specific image.
func parseSourcegraphCommitFromDeploymentManifestsDiff(output []byte) *ServiceVersionDiff {
	var diff ServiceVersionDiff
	addRegexp := imageDiffRegexp(true)
	delRegexp := imageDiffRegexp(false)

	outStr := string(output)
	matches := addRegexp.FindStringSubmatch(outStr)
	if len(matches) > 1 {
		diff.New = matches[1]
	}
	matches = delRegexp.FindStringSubmatch(outStr)
	if len(matches) > 1 {
		diff.Old = matches[1]
	}

	if diff.Old == "" || diff.New == "" {
		return nil
	}

	return &diff
}

func diffDeploymentManifest(path string) (*ServiceVersionDiff, error) {
	diffCommand := []string{"diff", "@^", path}
	output, err := exec.Command("git", diffCommand...).Output()
	if err != nil {
		return nil, err
	}
	imageDiff := parseSourcegraphCommitFromDeploymentManifestsDiff(output)
	return imageDiff, nil
}

type mockDeploymentDiffer struct {
	diffs map[string]*ServiceVersionDiff
}

func (m *mockDeploymentDiffer) Services() (map[string]*ServiceVersionDiff, error) {
	return m.diffs, nil
}

func NewMockManifestDeployementsDiffer(m map[string]*ServiceVersionDiff) DeploymentDiffer {
	return &mockDeploymentDiffer{
		diffs: m,
	}
}
