package git

import (
	"github.com/sourcegraph/go-diff/diff"
)

// Changes are the changes made to files in a repository.
type Changes struct {
	Modified []string `json:"modified"`
	Added    []string `json:"added"`
	Deleted  []string `json:"deleted"`
	Renamed  []string `json:"renamed"`
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
