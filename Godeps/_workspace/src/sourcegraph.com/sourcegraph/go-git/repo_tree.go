package git

func (repo *Repository) GetTree(idStr string) (*Tree, error) {
	return repo.getTree(ObjectIDHex(idStr))
}

func (repo *Repository) getTree(id ObjectID) (*Tree, error) {
	_, err := repo.object(id, true)
	if err != nil {
		if _, ok := err.(ObjectNotFound); ok {
			return nil, ErrNotExist
		}
		return nil, err
	}

	return NewTree(repo, id), nil
}
