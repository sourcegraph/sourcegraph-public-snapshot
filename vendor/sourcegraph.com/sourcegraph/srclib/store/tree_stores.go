package store

// scopeTrees returns a list of commit IDs that are matched by the
// filters. If potentially all commits could match, or if enough
// commits could potentially match that it would probably be cheaper
// to iterate through all of them, then a nil slice is returned. If
// none match, an empty slice is returned.
//
// scopeTrees is used to select which TreeStores to query.
//
// TODO(sqs): return an error if the filters are mutually exclusive?
func scopeTrees(filters []interface{}) ([]string, error) {
	commitIDs := map[string]struct{}{}
	everHadAny := false // whether unitIDs ever contained any commitIDs

	for _, f := range filters {
		switch f := f.(type) {
		case ByCommitIDsFilter:
			if len(commitIDs) == 0 && !everHadAny {
				everHadAny = true
				for _, c := range f.ByCommitIDs() {
					commitIDs[c] = struct{}{}
				}
			} else {
				// Intersect.
				newCommitIDs := make(map[string]struct{}, (len(commitIDs)+len(f.ByCommitIDs()))/2)
				for _, c := range f.ByCommitIDs() {
					if _, present := commitIDs[c]; present {
						newCommitIDs[c] = struct{}{}
					}
				}
				commitIDs = newCommitIDs
			}
		}
	}

	if len(commitIDs) == 0 && !everHadAny {
		// No unit scoping filters were present, so scope includes
		// potentially all commitIDs.
		return nil, nil
	}

	ids := make([]string, 0, len(commitIDs))
	for commitID := range commitIDs {
		ids = append(ids, commitID)
	}
	return ids, nil
}

// A treeStoreOpener opens the TreeStore for the specified tree.
type treeStoreOpener interface {
	openTreeStore(commitID string) TreeStore
	openAllTreeStores() (map[string]TreeStore, error)
}

// openCommitstores is a helper func that calls o.openTreeStore for
// each tree returned by scopeTrees(filters...).
func openTreeStores(o treeStoreOpener, filters interface{}) (map[string]TreeStore, error) {
	commitIDs, err := scopeTrees(storeFilters(filters))
	if err != nil {
		return nil, err
	}

	if commitIDs == nil {
		return o.openAllTreeStores()
	}

	tss := make(map[string]TreeStore, len(commitIDs))
	for _, commitID := range commitIDs {
		tss[commitID] = o.openTreeStore(commitID)
	}
	return tss, nil
}
