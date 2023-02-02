package types

import "time"

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
	ContentType       string
	ShouldReindex     bool
}

func (u Upload) RecordID() int {
	return u.ID
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
