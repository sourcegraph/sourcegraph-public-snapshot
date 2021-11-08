package batches

// Repository is a repository in which the steps of a batch spec are executed.
//
// It is also part of the cache.ExecutionKey, so changes to the names of fields
// here will lead to cache busts.
type Repository struct {
	ID          string
	Name        string
	BaseRef     string
	BaseRev     string
	FileMatches []string
}
