package core

import "github.com/sourcegraph/sourcegraph/internal/api"

type UploadLike interface {
	// GetID is purely present to help with debugging.
	GetID() int
	// GetRoot should return the upload's Root path.
	GetRoot() string
	// GetCommit should return the upload's Commit
	GetCommit() api.CommitID
}

type UploadSummary struct {
	ID     int
	Root   string
	Commit api.CommitID
}

func (u *UploadSummary) GetID() int {
	return u.ID
}

func (u *UploadSummary) GetRoot() string {
	return u.Root
}

func (u *UploadSummary) GetCommit() api.CommitID {
	return u.Commit
}

var _ UploadLike = &UploadSummary{}
