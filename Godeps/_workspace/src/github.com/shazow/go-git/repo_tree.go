package git

import ()

// Find the tree object in the repository.
func (repo *Repository) GetTree(idStr string) (*Tree, error) {
	id, err := NewIdFromString(idStr)
	if err != nil {
		return nil, err
	}
	return repo.getTree(id)
}

func (repo *Repository) getTree(id sha1) (*Tree, error) {
	treePath := filepathFromSHA1(repo.Path, id.String())
	if !isFile(treePath) {
		m := false
		for _, indexfile := range repo.indexfiles {
			if offset := indexfile.offsetValues[id]; offset != 0 {
				m = true
				break
			}
		}
		if !m {
			return nil, ErrNotExist
		}
	}

	return NewTree(repo, id), nil
}
