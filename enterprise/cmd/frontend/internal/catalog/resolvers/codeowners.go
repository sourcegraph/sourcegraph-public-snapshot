package resolvers

import (
	"bytes"
	pathpkg "path"
	"strings"

	"github.com/gobwas/glob"
	"github.com/sourcegraph/sourcegraph/internal/api"
)

type codeownersEntry struct {
	pathPattern string
	owners      []string

	glob glob.Glob // set only by codeownersComputer
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

func (c *codeownersComputer) add(repo api.RepoName, commit api.CommitID, path string, codeowners []byte) error {
	dir := pathpkg.Dir(path)
	if dir == ".github" {
		dir = "."
	}

	entries := parseCodeowners(codeowners)
	for i := range entries {
		entries[i].pathPattern = pathpkg.Join(dir, entries[i].pathPattern)

		glob, err := glob.Compile(entries[i].pathPattern)
		if err != nil {
			return err
		}
		entries[i].glob = glob
	}

	key := repoCommitKey{repo: repo, commit: commit}
	if c.at == nil {
		c.at = map[repoCommitKey][]codeownersEntry{}
	}
	c.at[key] = append(c.at[key], entries...)

	return nil
}

func (c *codeownersComputer) get(repo api.RepoName, commit api.CommitID, path string) (owners []string) {
	entries := c.at[repoCommitKey{repo: repo, commit: commit}]
	for _, e := range entries {
		if e.glob.Match(path) {
			return e.owners
		}
	}
	return nil
}
