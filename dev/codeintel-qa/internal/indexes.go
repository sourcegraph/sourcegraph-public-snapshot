package internal

import (
	"os"

	"github.com/grafana/regexp"
)

var indexFilenamePattern = regexp.MustCompile(`^(.+)\.\d+\.([0-9A-Fa-f]{40})\.dump$`)

// CommitsByRepo returns a map from repository name to a slice of commits for that repository.
// The repositories and commits are read from the filesystem state of the index directory
// supplied by the user. This method assumes that index files have been downloaded or generated
// locally.
func CommitsByRepo(indexDir string) (map[string][]string, error) {
	infos, err := os.ReadDir(indexDir)
	if err != nil {
		return nil, err
	}

	commitsByRepo := map[string][]string{}
	for _, info := range infos {
		if matches := indexFilenamePattern.FindStringSubmatch(info.Name()); len(matches) > 0 {
			commitsByRepo[matches[1]] = append(commitsByRepo[matches[1]], matches[2])
		}
	}

	return commitsByRepo, nil
}
