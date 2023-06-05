package diff

import (
	"bytes"
	"sort"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// ParseGitDiffNameStatus returns the paths changedA and changedB for commits
// A and B respectively. It expects to be parsing the output of the command
// git diff -z --name-status --no-renames A B.
func ParseGitDiffNameStatus(out []byte) (changedA, changedB []string, err error) {
	if len(out) == 0 {
		return nil, nil, nil
	}

	slices := bytes.Split(bytes.TrimRight(out, "\x00"), []byte{0})
	if len(slices)%2 != 0 {
		return nil, nil, errors.New("uneven pairs")
	}

	for i := 0; i < len(slices); i += 2 {
		path := string(slices[i+1])
		switch slices[i][0] {
		case 'D': // no longer appears in B
			changedA = append(changedA, path)
		case 'M':
			changedA = append(changedA, path)
			changedB = append(changedB, path)
		case 'A': // doesn't exist in A
			changedB = append(changedB, path)
		}
	}
	sort.Strings(changedA)
	sort.Strings(changedB)

	return changedA, changedB, nil
}
