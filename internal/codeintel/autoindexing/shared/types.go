package shared

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/types"
)

type IndexJob struct {
	Indexer string
}

type SourcedCommits struct {
	RepositoryID   int
	RepositoryName string
	Commits        []string
}

// IndexConfiguration stores the index configuration for a repository.
type IndexConfiguration struct {
	ID           int    `json:"id"`
	RepositoryID int    `json:"repository_id"`
	Data         []byte `json:"data"`
}

type IndexesWithRepositoryNamespace struct {
	Root    string
	Indexer string
	Indexes []types.Index
}

// UploadLocation is a path and range pair from within a particular upload. The target commit
// denotes the target commit for which the location was set (the originally requested commit).
type UploadLocation struct {
	Dump         Dump
	Path         string
	TargetCommit string
	TargetRange  Range
}

// Dump is a subset of the lsif_uploads table (queried via the lsif_dumps_with_repository_name view)
// and stores only processed records.
type Dump struct {
	ID                int        `json:"id"`
	Commit            string     `json:"commit"`
	Root              string     `json:"root"`
	VisibleAtTip      bool       `json:"visibleAtTip"`
	UploadedAt        time.Time  `json:"uploadedAt"`
	State             string     `json:"state"`
	FailureMessage    *string    `json:"failureMessage"`
	StartedAt         *time.Time `json:"startedAt"`
	FinishedAt        *time.Time `json:"finishedAt"`
	ProcessAfter      *time.Time `json:"processAfter"`
	NumResets         int        `json:"numResets"`
	NumFailures       int        `json:"numFailures"`
	RepositoryID      int        `json:"repositoryId"`
	RepositoryName    string     `json:"repositoryName"`
	Indexer           string     `json:"indexer"`
	IndexerVersion    string     `json:"indexerVersion"`
	AssociatedIndexID *int       `json:"associatedIndex"`
}

// Range is an inclusive bounds within a file.
type Range struct {
	Start Position
	End   Position
}

// Position is a unique position within a file.
type Position struct {
	Line      int
	Character int
}

type DeleteUploadsOptions struct {
	State string
	Term  string
}
