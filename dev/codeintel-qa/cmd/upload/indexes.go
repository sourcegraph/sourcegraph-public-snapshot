package main

import (
	"os"
	"regexp"
)

var indexFilenamePattern = regexp.MustCompile(`^(.+)\.\d+\.([0-9A-Fa-f]{40})\.dump$`)

func commitsByRepo() (map[string][]string, error) {
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
