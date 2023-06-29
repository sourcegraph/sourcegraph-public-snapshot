package internal

import (
	"fmt"
	"os"

	"github.com/grafana/regexp"
)

var indexFilenamePattern = regexp.MustCompile(`^([^.]+)\.([^.]+)\.([0-9A-Fa-f]{40})\.([^.]+)\.(scip|dump)$`)

type ExtensionCommitAndRoot struct {
	Extension string
	Commit    string
	Root      string
}

// ExtensionAndCommitsByRepo returns a map from org+repository name to a slice of commit and extension
// pairs for that repository. The repositories and commits are read from the filesystem state of the
// index directory supplied by the user. This method assumes that index files have been downloaded or
// generated locally.
func ExtensionAndCommitsByRepo(indexDir string) (map[string][]ExtensionCommitAndRoot, error) {
	infos, err := os.ReadDir(indexDir)
	if err != nil {
		return nil, err
	}

	commitsByRepo := map[string][]ExtensionCommitAndRoot{}
	for _, info := range infos {
		if matches := indexFilenamePattern.FindStringSubmatch(info.Name()); len(matches) > 0 {
			orgRepo := fmt.Sprintf("%s/%s", matches[1], matches[2])
			root := matches[4]
			commitsByRepo[orgRepo] = append(commitsByRepo[orgRepo], ExtensionCommitAndRoot{Extension: matches[5], Commit: matches[3], Root: root})
		}
	}

	return commitsByRepo, nil
}
