package types

import "time"

// BatchSpecMount contains the metadata about the mount object for the batch spec.
type BatchSpecMount struct {
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

// Clone returns a clone of a BatchSpecMount.
func (b *BatchSpecMount) Clone() *BatchSpecMount {
	clone := *b
	return &clone
}
