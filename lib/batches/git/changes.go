package git

import (
	"strings"

	"github.com/sourcegraph/go-diff/diff"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Changes are the changes made to files in a repository.
type Changes struct {
	Modified []string `json:"modified"`
	Added    []string `json:"added"`
	Deleted  []string `json:"deleted"`
	Renamed  []string `json:"renamed"`
}

// ParseStatus parses the output of `git status` and turns it into Changes.
func ParseGitStatus(out []byte) (Changes, error) {
	result := Changes{}

	stripped := strings.TrimSpace(string(out))
	if stripped == "" {
		return result, nil
	}

	for _, line := range strings.Split(stripped, "\n") {
		if len(line) < 4 {
			return result, errors.Newf("git status line has unrecognized format: %q", line)
		}

		file := line[3:]

		switch line[0] {
		case 'M':
			result.Modified = append(result.Modified, file)
		case 'A':
			result.Added = append(result.Added, file)
		case 'D':
			result.Deleted = append(result.Deleted, file)
		case 'R':
			files := strings.Split(file, " -> ")
			newFile := files[len(files)-1]
			result.Renamed = append(result.Renamed, newFile)
		}
	}

	return result, nil
}

func ChangesInDiff(rawDiff []byte) (Changes, error) {
	result := Changes{}

	fileDiffs, err := diff.ParseMultiFileDiff(rawDiff)
	if err != nil {
		return result, err
	}

	for _, fd := range fileDiffs {
		switch {
		case fd.NewName == "/dev/null":
			result.Deleted = append(result.Deleted, fd.OrigName)
		case fd.OrigName == "/dev/null":
			result.Added = append(result.Added, fd.NewName)
		case fd.OrigName == fd.NewName:
			result.Modified = append(result.Modified, fd.OrigName)
		case fd.OrigName != fd.NewName:
			result.Renamed = append(result.Renamed, fd.NewName)
		}
	}

	return result, nil
}
