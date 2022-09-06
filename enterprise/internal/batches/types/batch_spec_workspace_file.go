package types

import "time"

// BatchSpecWorkspaceFile contains the metadata about the workspace file for the batch spec.
type BatchSpecWorkspaceFile struct {
	ID          int64
	RandID      string
	BatchSpecID int64

	FileName string
	Path     string
	Size     int64
	Content  []byte
	// ModifiedAt is when the file was last touched. Compared to UpdatedAt, this field is the filesystem modtime versus
	// when updated in the database.
	ModifiedAt time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}

// Clone returns a clone of a BatchSpecWorkspaceFile.
func (b *BatchSpecWorkspaceFile) Clone() *BatchSpecWorkspaceFile {
	clone := *b
	return &clone
}
