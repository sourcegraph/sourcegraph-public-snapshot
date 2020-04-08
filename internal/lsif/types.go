package lsif

import (
	"time"

	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/sourcegraph/internal/api"
)

type LSIFUpload struct {
	ID                int64      `json:"id"`
	RepositoryID      api.RepoID `json:"repositoryId"`
	Commit            string     `json:"commit"`
	Root              string     `json:"root"`
	Indexer           string     `json:"indexer"`
	Filename          string     `json:"filename"`
	State             string     `json:"state"`
	UploadedAt        time.Time  `json:"uploadedAt"`
	StartedAt         *time.Time `json:"startedAt"`
	FinishedAt        *time.Time `json:"finishedAt"`
	FailureSummary    *string    `json:"failureSummary"`
	FailureStacktrace *string    `json:"failureStacktrace"`
	VisibleAtTip      bool       `json:"visibleAtTip"`
	PlaceInQueue      *int32     `json:"placeInQueue"`
}

type LSIFLocation struct {
	RepositoryID api.RepoID `json:"repositoryId"`
	Commit       string     `json:"commit"`
	Path         string     `json:"path"`
	Range        lsp.Range  `json:"range"`
}
