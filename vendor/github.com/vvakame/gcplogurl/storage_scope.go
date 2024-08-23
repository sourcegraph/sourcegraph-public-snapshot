package gcplogurl

// StorageScope means scope about logs.
type StorageScope interface {
	isStorageScope()
	marshalURL(vs values)
}

var _ StorageScope = StorageScopeProject
var _ StorageScope = (*StorageScopeStorage)(nil)

// StorageScopeProject use scope by project.
const StorageScopeProject = storageScopeProject(0)

type storageScopeProject int

func (storageScopeProject) isStorageScope() {}

func (storageScopeProject) marshalURL(vs values) {
	vs.Set("storageScope", "project")
}

// StorageScopeStorage use scope by storage.
type StorageScopeStorage struct {
	Src []string
}

func (s *StorageScopeStorage) isStorageScope() {}

func (s *StorageScopeStorage) marshalURL(vs values) {
	vs.Set("storageScope", "storage")
	for _, src := range s.Src {
		vs.Add("storageScope", src)
	}
}
