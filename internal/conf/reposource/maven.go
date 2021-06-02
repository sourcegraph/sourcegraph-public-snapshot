package reposource

import (
	"strings"
)

func MavenRepoName(dependency string) string {
	return "maven/" + strings.ReplaceAll(dependency, ":", "/")
}

func DecomposeMavenPath(path string) string {
	return strings.ReplaceAll(strings.TrimPrefix(path, "maven/"), "/", ":")
}
