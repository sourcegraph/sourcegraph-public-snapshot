package lsif

import (
	"time"

	"github.com/sourcegraph/go-lsp"
)

type LSIFUpload struct {
	ID                int64      `json:"id"`
	Repository        string     `json:"repository"`
	Commit            string     `json:"commit"`
	Root              string     `json:"root"`
	Filename          string     `json:"filename"`
	State             string     `json:"state"`
	UploadedAt        time.Time  `json:"uploadedAt"`
	StartedAt         *time.Time `json:"startedAt"`
	FinishedAt        *time.Time `json:"finishedAt"`
	FailureSummary    *string    `json:"failureSummary"`
	FailureStacktrace *string    `json:"failureStacktrace"`
	VisibleAtTip      bool       `json:"visibleAtTip"`
}

type LSIFLocation struct {
	Repository string    `json:"repository"`
	Commit     string    `json:"commit"`
	Path       string    `json:"path"`
	Range      lsp.Range `json:"range"`
}
