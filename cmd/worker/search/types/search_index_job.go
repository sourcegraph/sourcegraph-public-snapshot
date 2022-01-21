package types

import "github.com/sourcegraph/sourcegraph/internal/api"

type SearchIndexJob struct {
	ID       int
	RepoID   api.RepoID
	Revision string
}

func (s *SearchIndexJob) RecordID() int {
	return s.ID
}
