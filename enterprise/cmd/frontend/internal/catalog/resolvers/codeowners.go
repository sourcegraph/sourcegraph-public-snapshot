package resolvers

import (
	"bytes"
	pathpkg "path"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

type codeownersEntry struct {
	pathPattern string
	owners      []string
}

func parseCodeowners(data []byte) []codeownersEntry {
	var entries []codeownersEntry
	lines := bytes.Split(data, []byte("\n"))
	for _, line := range lines {
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		parts := strings.Fields(string(line))
		if len(parts) < 2 {
			continue
		}
		entries = append(entries, codeownersEntry{
			pathPattern: pathpkg.Clean(parts[0]),
			owners:      parts[1:],
		})
	}

	return entries
}

type repoCommitKey struct {
	repo   api.RepoName
	commit api.CommitID
}

type codeownersComputer struct {
	at map[repoCommitKey][]codeownersEntry
}

func (c *codeownersComputer) has(repo api.RepoName, commit api.CommitID, path string) bool {
	return c.at[repoCommitKey{repo: repo, commit: commit}] != nil
}

func (c *codeownersComputer) add(repo api.RepoName, commit api.CommitID, path string, codeowners []byte) {
	dir := pathpkg.Dir(path)
	if dir == ".github" {
		dir = "."
	}

	entries := parseCodeowners(codeowners)
	for i := range entries {
		entries[i].pathPattern = pathpkg.Join(dir, entries[i].pathPattern)
	}

	key := repoCommitKey{repo: repo, commit: commit}
	c.at[key] = append(c.at[key], entries...)
}

func (c *codeownersComputer) get(repo api.RepoName, commit api.CommitID, path string) (owners []string) {
	entries := c.at[repoCommitKey{repo: repo, commit: commit}]
	for _, e := range entries {
		if matched, _ := pathpkg.Match(e.pathPattern, path); matched {
			return e.owners
		}
	}
	return nil
}
