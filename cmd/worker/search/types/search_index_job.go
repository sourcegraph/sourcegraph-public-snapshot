package types

import "github.com/sourcegraph/sourcegraph/internal/api"

type SearchIndexJob struct {
	ID        int
	RepoID    api.RepoID
	BranchRef string
	Revision  string

	UploadIDs []string
}

func (s *SearchIndexJob) RecordID() int {
	return s.ID
}
