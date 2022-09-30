package types

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

// RevSpecSet is a utility type for a set of RevSpecs.
type RevSpecSet map[api.RevSpec]struct{}

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

// UploadLocation is a path and range pair from within a particular upload. The target commit
// denotes the target commit for which the location was set (the originally requested commit).
type UploadLocation struct {
	Dump         Dump
	Path         string
	TargetCommit string
	TargetRange  Range
}

type Range struct {
	Start Position
	End   Position
}

// Position is a unique position within a file.
type Position struct {
	Line      int
	Character int
}

type UploadLog struct {
	LogTimestamp      time.Time
	RecordDeletedAt   *time.Time
	UploadID          int
	Commit            string
	Root              string
	RepositoryID      int
	UploadedAt        time.Time
	Indexer           string
	IndexerVersion    *string
	UploadSize        *int
	AssociatedIndexID *int
	TransitionColumns []map[string]*string
	Reason            *string
	Operation         string
}

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
	UncompressedSize  *int64
	Rank              *int
	AssociatedIndexID *int
}

func (u Upload) RecordID() int {
	return u.ID
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

type DeleteUploadsOptions struct {
	State        string
	Term         string
	VisibleAtTip bool
}

type GetConfigurationPoliciesOptions struct {
	// RepositoryID indicates that only configuration policies that apply to the
	// specified repository (directly or via pattern) should be returned. This value
	// has no effect when equal to zero.
	RepositoryID int

	// Term is a string to search within the configuration title.
	Term string

	// ForIndexing indicates that only configuration policies with data retention enabled
	// should be returned.
	ForDataRetention bool

	// ForIndexing indicates that only configuration policies with indexing enabled should
	// be returned.
	ForIndexing bool

	// Limit indicates the number of results to take from the result set.
	Limit int

	// Offset indicates the number of results to skip in the result set.
	Offset int
}

type ConfigurationPolicy struct {
	ID                        int
	RepositoryID              *int
	RepositoryPatterns        *[]string
	Name                      string
	Type                      GitObjectType
	Pattern                   string
	Protected                 bool
	RetentionEnabled          bool
	RetentionDuration         *time.Duration
	RetainIntermediateCommits bool
	IndexingEnabled           bool
	IndexCommitMaxAge         *time.Duration
	IndexIntermediateCommits  bool
}

type RetentionPolicyMatchCandidate struct {
	*ConfigurationPolicy
	Matched           bool
	ProtectingCommits []string
}

type GitObjectType string

const (
	GitObjectTypeCommit GitObjectType = "GIT_COMMIT"
	GitObjectTypeTag    GitObjectType = "GIT_TAG"
	GitObjectTypeTree   GitObjectType = "GIT_TREE"
)
