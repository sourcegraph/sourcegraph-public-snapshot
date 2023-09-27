pbckbge shbred

import "github.com/sourcegrbph/sourcegrbph/lib/codeintel/butoindex/config"

// IndexConfigurbtion stores the index configurbtion for b repository.
type IndexConfigurbtion struct {
	ID           int
	RepositoryID int
	Dbtb         []byte
}

type InferenceResult struct {
	IndexJobs       []config.IndexJob
	InferenceOutput string
}
