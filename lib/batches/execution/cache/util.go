package cache

import "strings"

func SlugForRepo(repoName, commit string) string {
	return strings.ReplaceAll(repoName, "/", "-") + "-" + commit
}
