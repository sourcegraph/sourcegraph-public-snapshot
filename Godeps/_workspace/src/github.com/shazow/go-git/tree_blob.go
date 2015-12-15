package git

import (
	"os"
	"path"
	"strings"
)

var ErrNotExist = os.ErrNotExist

func (t *Tree) GetTreeEntryByPath(rpath string) (*TreeEntry, error) {
	if len(rpath) == 0 {
		return nil, ErrNotExist
	}

	parts := strings.Split(path.Clean(rpath), "/")
	var err error
	tree := t
	for i, name := range parts {
		if i == len(parts)-1 {
			entries, err := tree.ListEntries()
			if err != nil {
				return nil, err
			}
			for _, v := range entries {
				if v.name == name {
					return v, nil
				}
			}
		} else {
			tree, err = tree.SubTree(name)
			if err != nil {
				return nil, err
			}
		}
	}

	return nil, ErrNotExist
}

func (t *Tree) GetBlobByPath(rpath string) (*Blob, error) {
	entry, err := t.GetTreeEntryByPath(rpath)
	if err != nil {
		return nil, err
	}

	if !entry.IsDir() {
		return entry.Blob(), nil
	}

	return nil, ErrNotExist
}
