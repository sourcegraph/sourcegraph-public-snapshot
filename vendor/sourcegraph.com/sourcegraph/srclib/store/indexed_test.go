package store

import "testing"

func TestIndexedUnitStore(t *testing.T) {
	useIndexedStore = true
	testUnitStore(t, func() UnitStoreImporter {
		return newIndexedUnitStore(newTestFS(), "")
	})
}

func TestIndexedTreeStore(t *testing.T) {
	useIndexedStore = true
	testTreeStore(t, func() TreeStoreImporter {
		return newIndexedTreeStore(newTestFS(), "test")
	})
}

func TestIndexedFSTreeStore(t *testing.T) {
	useIndexedStore = true
	testTreeStore(t, func() TreeStoreImporter {
		return newFSTreeStore(newTestFS())
	})
}

func TestIndexedFSRepoStore(t *testing.T) {
	useIndexedStore = true
	testRepoStore(t, func() RepoStoreImporter {
		return NewFSRepoStore(newTestFS())
	})
}

func TestIndexedFSMultiRepoStore(t *testing.T) {
	useIndexedStore = true
	testMultiRepoStore(t, func() MultiRepoStoreImporter {
		return NewFSMultiRepoStore(newTestFS(), nil)
	})
}

func TestIndexedFSMultiRepoStore_customRepoPaths(t *testing.T) {
	useIndexedStore = true
	testMultiRepoStore(t, func() MultiRepoStoreImporter {
		return NewFSMultiRepoStore(newTestFS(), &FSMultiRepoStoreConf{RepoPaths: &customRepoPaths{}})
	})
}
