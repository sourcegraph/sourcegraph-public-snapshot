package shared

import "time"

// Upload is a subset of the lsif_uploads table and stores both processed and unprocessed
// records.
type Upload struct {
	ID                int
	Commit            string
	Root              string
	VisibleAtTip      bool
	UploadedAt        time.Time
	State             string
	FailureMessage    *string
	StartedAt         *time.Time
	FinishedAt        *time.Time
	ProcessAfter      *time.Time
	NumResets         int
	NumFailures       int
	RepositoryID      int
	RepositoryName    string
	Indexer           string
	IndexerVersion    string
	NumParts          int
	UploadedParts     []int
	UploadSize        *int64
	Rank              *int
	AssociatedIndexID *int
}

// Dump is a subset of the lsif_uploads table (queried via the lsif_dumps_with_repository_name view)
// and stores only processed records.
type Dump struct {
	ID                int
	Commit            string
	Root              string
	VisibleAtTip      bool
	UploadedAt        time.Time
	State             string
	FailureMessage    *string
	StartedAt         *time.Time
	FinishedAt        *time.Time
	ProcessAfter      *time.Time
	NumResets         int
	NumFailures       int
	RepositoryID      int
	RepositoryName    string
	Indexer           string
	IndexerVersion    string
	AssociatedIndexID *int
}

type SourcedCommits struct {
	RepositoryID   int
	RepositoryName string
	Commits        []string
}

type GetUploadsOptions struct {
	RepositoryID            int
	State                   string
	Term                    string
	VisibleAtTip            bool
	DependencyOf            int
	DependentOf             int
	UploadedBefore          *time.Time
	UploadedAfter           *time.Time
	LastRetentionScanBefore *time.Time
	AllowExpired            bool
	AllowDeletedRepo        bool
	AllowDeletedUpload      bool
	OldestFirst             bool
	Limit                   int
	Offset                  int

	// InCommitGraph ensures that the repository commit graph was updated strictly
	// after this upload was processed. This condition helps us filter out new uploads
	// that we might later mistake for unreachable.
	InCommitGraph bool
}

// Package pairs a package schem+name+version with the dump that provides it.
type Package struct {
	DumpID  int
	Scheme  string
	Name    string
	Version string
}

// PackageReference is a package scheme+name+version
type PackageReference struct {
	Package
}

type DependencyReferenceCountUpdateType int

const (
	DependencyReferenceCountUpdateTypeNone DependencyReferenceCountUpdateType = iota
	DependencyReferenceCountUpdateTypeAdd
	DependencyReferenceCountUpdateTypeRemove
)
