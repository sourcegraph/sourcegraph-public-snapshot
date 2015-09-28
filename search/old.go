package search

import "strings"

// ParseRepoAndCommitID parses strings like "example.com/repo" and
// "example.com/repo@myrev".
func ParseRepoAndCommitID(repoAndCommitID string) (uri, commitID string) {
	if i := strings.Index(repoAndCommitID, "@"); i != -1 {
		return repoAndCommitID[:i], repoAndCommitID[i+1:]
	}
	return repoAndCommitID, ""
}
