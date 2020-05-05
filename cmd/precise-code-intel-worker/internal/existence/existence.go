package existence

import (
	"path/filepath"
)

type ExistenceChecker struct {
	root              string
	directoryContents map[string]StringSet
}

// NewExistenceChecker constructs a map of directory contents from the given set of paths and the given
// getChildren function pointer that determines which of the given paths exist in the git clone at the
// target commit.
func NewExistenceChecker(root string, paths []string, getChildren GetChildrenFunc) (*ExistenceChecker, error) {
	directoryContents, err := directoryContents(root, paths, getChildren)
	if err != nil {
		return nil, err
	}

	return &ExistenceChecker{root, directoryContents}, nil
}

// Exists determines if the given path exists in the git clone at the target commit.
func (ec *ExistenceChecker) Exists(path string) bool {
	if children, ok := ec.directoryContents[dirWithoutDot(filepath.Join(ec.root, path))]; ok {
		if _, ok := children[path]; ok {
			return true
		}
	}

	return false
}
