package dbstore

// IndexConfiguration stores the index configuration for a repository.
type IndexConfiguration struct {
	ID           int    `json:"id"`
	RepositoryID int    `json:"repository_id"`
	Data         []byte `json:"data"`
}
