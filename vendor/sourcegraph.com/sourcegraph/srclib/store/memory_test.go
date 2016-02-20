package store

import "testing"

func TestMemoryUnitStore(t *testing.T) {
	testUnitStore(t, func() UnitStoreImporter {
		return &memoryUnitStore{}
	})
}

func TestMemoryTreeStore(t *testing.T) {
	testTreeStore(t, func() TreeStoreImporter {
		return newMemoryTreeStore()
	})
}

func TestMemoryRepoStore(t *testing.T) {
	testRepoStore(t, func() RepoStoreImporter {
		return newMemoryRepoStore()
	})
}

func TestMemoryMultiRepoStore(t *testing.T) {
	testMultiRepoStore(t, func() MultiRepoStoreImporter {
		return newMemoryMultiRepoStore()
	})
}
