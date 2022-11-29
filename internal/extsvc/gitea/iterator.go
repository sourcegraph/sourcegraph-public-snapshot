package gitea

// RepositoryIterator is an iterator for the list of repositories returned in
// a Gitea request.
//
// Note: there is underlying pagination (not cursor), so if the underlying
// collection is modified between pages it is possible to miss or duplicate
// repositories. As such care needs to be taken for deterministic ordering.
type RepositoryIterator struct {
	current []*Repository
	err     error
	done    bool

	// next is a function which is repeatedly called until no items are
	// returned or there is a non-nil error. These items are returned one by
	// one via Next and Current.
	next func() ([]*Repository, error)
}

func (it *RepositoryIterator) Next() bool {
	if it.done {
		return false
	}

	if len(it.current) > 1 {
		it.current = it.current[1:]
		return true
	}

	it.current, it.err = it.next()
	if len(it.current) == 0 || it.err != nil {
		it.done = true
	}

	return !it.done
}

func (it *RepositoryIterator) Current() *Repository {
	return it.current[0]
}

func (it *RepositoryIterator) Err() error {
	return it.err
}
