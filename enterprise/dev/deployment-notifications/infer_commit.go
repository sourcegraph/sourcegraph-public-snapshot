package main

import (
	"os/exec"
	"strings"

	"github.com/grafana/regexp"
)

var imageCommitRegexp = `(?m)^\+\s+image:\s[^/]+\/sourcegraph\/APPNAME:\d{6}_\d{4}-\d{2}-\d{2}_([^@]+)@sha256.*$` // (?m) stands for multiline.

func inferSourcegraphCommit() (string, error) {
	// There are multiple files being updated when a deployment PR occurs. We're picking those two here
	// for no particular reason other than the chances are pretty low for these to change soon.
	files := []string{
		// If we're looking at a continuous deployment, we'll always find the frontend being updated.
		"base/frontend/sourcegraph-frontend-internal.Deployment.yaml",
		// Else, it's either a daily deployment with giterver or Tomas bumping it.
		"base/gitserver/gitserver.StatefulSet.yaml",
	}
	for _, file := range files {
		diffCommand := []string{"diff", "@^", file}
		if output, err := exec.Command("git", diffCommand...).Output(); err != nil {
			commit := extractSourcegraphCommitFromDeploymentManifestsDiff(output, "frontend")
			if commit != "" {
				return commit, nil
			} else {
				continue
			}
		} else {
			return "", err
		}
	}
	return "", nil
}

func extractSourcegraphCommitFromDeploymentManifestsDiff(output []byte, appname string) string {
	appRegexp := regexp.MustCompile(strings.ReplaceAll(imageCommitRegexp, "APPNAME", appname))
	matches := appRegexp.FindStringSubmatch(string(output))
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}
