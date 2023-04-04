package shared

// IndexConfiguration stores the index configuration for a repository.
type IndexConfiguration struct {
	ID           int
	RepositoryID int
	Data         []byte
}
