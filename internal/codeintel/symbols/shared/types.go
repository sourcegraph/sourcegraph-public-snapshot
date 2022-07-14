package shared

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

type Symbol struct {
	Name string
}

// Location is an LSP-like location scoped to a dump.
type Location struct {
	DumpID int
	Path   string
	Range  Range
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

type RequestArgs struct {
	RepositoryID int
	Commit       string
	Path         string
	Line         int
	Character    int
	Limit        int
	RawCursor    string
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

// scanDumps scans a slice of dumps from the return value of `*Store.query`.
func scanDump(s dbutil.Scanner) (dump Dump, err error) {
	return dump, s.Scan(
		&dump.ID,
		&dump.Commit,
		&dump.Root,
		&dump.VisibleAtTip,
		&dump.UploadedAt,
		&dump.State,
		&dump.FailureMessage,
		&dump.StartedAt,
		&dump.FinishedAt,
		&dump.ProcessAfter,
		&dump.NumResets,
		&dump.NumFailures,
		&dump.RepositoryID,
		&dump.RepositoryName,
		&dump.Indexer,
		&dbutil.NullString{S: &dump.IndexerVersion},
		&dump.AssociatedIndexID,
	)
}
