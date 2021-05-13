package reposource

import (
	"strings"
)

func DecomposeMavenPath(path string) string {
	return strings.ReplaceAll(strings.TrimPrefix(path, "maven/"), "/", ":")
}
